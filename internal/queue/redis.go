package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/zerodha/logf"
)

const (
	// StreamName is the Redis stream for campaign jobs
	StreamName = "whatomate:campaigns"

	// ConsumerGroup is the consumer group name for workers
	ConsumerGroup = "campaign-workers"

	// BlockTimeout is how long to block waiting for new messages
	BlockTimeout = 5 * time.Second

	// ClaimMinIdleTime is the minimum idle time before claiming a pending message
	ClaimMinIdleTime = 5 * time.Minute

	// ClaimInterval is how often a running consumer looks for stale pending
	// messages left behind by crashed workers. Claiming only at startup meant
	// a job stuck in another consumer's pending list was never retried until
	// the next restart.
	ClaimInterval = time.Minute

	// MaxDeliveries caps how many times one message is handed to a worker.
	// A job that keeps failing (unmarshal error, deleted campaign) is dropped
	// and logged after this many attempts instead of being retried forever.
	MaxDeliveries = 5
)

// RedisQueue implements the Queue interface using Redis Streams
type RedisQueue struct {
	client *redis.Client
	log    logf.Logger
}

// NewRedisQueue creates a new Redis queue
func NewRedisQueue(client *redis.Client, log logf.Logger) *RedisQueue {
	return &RedisQueue{
		client: client,
		log:    log,
	}
}

// EnqueueRecipient adds a single recipient job to the queue
func (q *RedisQueue) EnqueueRecipient(ctx context.Context, job *RecipientJob) error {
	if job.EnqueuedAt.IsZero() {
		job.EnqueuedAt = time.Now()
	}

	payload, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("failed to marshal recipient job: %w", err)
	}

	_, err = q.client.XAdd(ctx, &redis.XAddArgs{
		Stream: StreamName,
		Values: map[string]any{
			"type":    string(JobTypeRecipient),
			"payload": string(payload),
		},
	}).Result()

	if err != nil {
		return fmt.Errorf("failed to enqueue recipient job: %w", err)
	}

	return nil
}

// EnqueueRecipients adds multiple recipient jobs to the queue using pipeline
func (q *RedisQueue) EnqueueRecipients(ctx context.Context, jobs []*RecipientJob) error {
	if len(jobs) == 0 {
		return nil
	}

	pipe := q.client.Pipeline()
	now := time.Now()

	for _, job := range jobs {
		if job.EnqueuedAt.IsZero() {
			job.EnqueuedAt = now
		}

		payload, err := json.Marshal(job)
		if err != nil {
			return fmt.Errorf("failed to marshal recipient job: %w", err)
		}

		pipe.XAdd(ctx, &redis.XAddArgs{
			Stream: StreamName,
			Values: map[string]any{
				"type":    string(JobTypeRecipient),
				"payload": string(payload),
			},
		})
	}

	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to enqueue recipient jobs: %w", err)
	}

	q.log.Info("Recipient jobs enqueued", "count", len(jobs), "campaign_id", jobs[0].CampaignID)
	return nil
}

// Close closes the queue connection
func (q *RedisQueue) Close() error {
	return nil // Redis client is managed externally
}

// RedisConsumer implements the Consumer interface using Redis Streams
type RedisConsumer struct {
	client     *redis.Client
	log        logf.Logger
	consumerID string
}

// NewRedisConsumer creates a new Redis consumer
func NewRedisConsumer(client *redis.Client, log logf.Logger) (*RedisConsumer, error) {
	// Generate unique consumer ID
	hostname, _ := os.Hostname()
	consumerID := fmt.Sprintf("worker-%s-%d", hostname, os.Getpid())

	consumer := &RedisConsumer{
		client:     client,
		log:        log,
		consumerID: consumerID,
	}

	if err := consumer.ensureGroup(context.Background()); err != nil {
		return nil, err
	}

	log.Info("Redis consumer initialized", "consumer_id", consumerID)
	return consumer, nil
}

// ensureGroup creates the stream and consumer group when they do not exist.
// It is also used to recover when the stream is deleted while a consumer is
// running (for example by an operator or a cache flush), which otherwise makes
// every XREADGROUP fail with NOGROUP until the process is restarted.
func (c *RedisConsumer) ensureGroup(ctx context.Context) error {
	err := c.client.XGroupCreateMkStream(ctx, StreamName, ConsumerGroup, "0").Err()
	if err != nil && !strings.HasPrefix(err.Error(), "BUSYGROUP") {
		return fmt.Errorf("failed to create consumer group: %w", err)
	}
	return nil
}

// Consume starts consuming jobs from the queue
func (c *RedisConsumer) Consume(ctx context.Context, handler JobHandler) error {
	c.log.Info("Starting to consume jobs", "consumer_id", c.consumerID)

	// First, try to claim any stale pending messages from crashed workers
	if err := c.claimPendingMessages(ctx, handler); err != nil {
		c.log.Warn("Failed to claim pending messages", "error", err)
	}
	lastClaim := time.Now()

	for {
		select {
		case <-ctx.Done():
			c.log.Info("Consumer shutting down")
			return ctx.Err()
		default:
		}

		// Periodically re-check for stale pending messages so a job orphaned
		// by a crashed worker is retried while this consumer is running.
		if time.Since(lastClaim) >= ClaimInterval {
			if err := c.claimPendingMessages(ctx, handler); err != nil && ctx.Err() == nil {
				c.log.Warn("Failed to claim pending messages", "error", err)
			}
			lastClaim = time.Now()
		}

		// Read new messages from the stream
		streams, err := c.client.XReadGroup(ctx, &redis.XReadGroupArgs{
			Group:    ConsumerGroup,
			Consumer: c.consumerID,
			Streams:  []string{StreamName, ">"},
			Count:    1,
			Block:    BlockTimeout,
		}).Result()

		if err != nil {
			if err == redis.Nil {
				// No messages available, continue waiting
				continue
			}
			if ctx.Err() != nil {
				return ctx.Err()
			}
			if strings.HasPrefix(err.Error(), "NOGROUP") {
				c.log.Warn("Stream or consumer group missing, recreating", "stream", StreamName, "group", ConsumerGroup)
				if gerr := c.ensureGroup(ctx); gerr != nil {
					c.log.Error("Failed to recreate consumer group", "error", gerr)
				} else {
					continue
				}
			} else {
				c.log.Error("Failed to read from stream", "error", err)
			}
			time.Sleep(time.Second) // Back off on error
			continue
		}

		for _, stream := range streams {
			for _, msg := range stream.Messages {
				if err := c.processMessage(ctx, msg, handler); err != nil {
					c.log.Error("Failed to process message", "error", err, "message_id", msg.ID)
					// Don't ACK failed messages - they'll be reclaimed later
					continue
				}

				// Acknowledge the message
				if err := c.ack(msg.ID); err != nil {
					c.log.Error("Failed to ACK message", "error", err, "message_id", msg.ID)
				}
			}
		}
	}
}

// ack acknowledges a processed message using a detached context, so a
// shutdown that cancels the consumer context while a job is finishing cannot
// leave the completed job pending in the stream (and therefore redelivered,
// i.e. sent to the customer again, on the next start).
func (c *RedisConsumer) ack(id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return c.client.XAck(ctx, StreamName, ConsumerGroup, id).Err()
}

// claimPendingMessages claims stale pending messages from crashed workers
func (c *RedisConsumer) claimPendingMessages(ctx context.Context, handler JobHandler) error {
	// Get pending messages that have been idle for too long
	pending, err := c.client.XPendingExt(ctx, &redis.XPendingExtArgs{
		Stream: StreamName,
		Group:  ConsumerGroup,
		Start:  "-",
		End:    "+",
		Count:  100,
		Idle:   ClaimMinIdleTime,
	}).Result()

	if err != nil {
		return fmt.Errorf("failed to get pending messages: %w", err)
	}

	if len(pending) == 0 {
		return nil
	}

	c.log.Info("Found stale pending messages to claim", "count", len(pending))

	// Claim and process each pending message
	for _, p := range pending {
		// Drop poison messages instead of retrying them forever. The payload
		// is logged so an operator can replay it once the cause is fixed.
		if p.RetryCount > MaxDeliveries {
			c.log.Error("Dropping message after too many delivery attempts",
				"message_id", p.ID, "deliveries", p.RetryCount, "consumer", p.Consumer)
			if err := c.ack(p.ID); err != nil {
				c.log.Error("Failed to ACK poison message", "error", err, "message_id", p.ID)
			}
			continue
		}

		// Claim the message
		messages, err := c.client.XClaim(ctx, &redis.XClaimArgs{
			Stream:   StreamName,
			Group:    ConsumerGroup,
			Consumer: c.consumerID,
			MinIdle:  ClaimMinIdleTime,
			Messages: []string{p.ID},
		}).Result()

		if err != nil {
			c.log.Error("Failed to claim message", "error", err, "message_id", p.ID)
			continue
		}

		for _, msg := range messages {
			if err := c.processMessage(ctx, msg, handler); err != nil {
				c.log.Error("Failed to process claimed message", "error", err, "message_id", msg.ID)
				continue
			}

			// Acknowledge the message
			if err := c.ack(msg.ID); err != nil {
				c.log.Error("Failed to ACK claimed message", "error", err, "message_id", msg.ID)
			}
		}
	}

	return nil
}

// processMessage processes a single message from the stream
func (c *RedisConsumer) processMessage(ctx context.Context, msg redis.XMessage, handler JobHandler) error {
	jobType, ok := msg.Values["type"].(string)
	if !ok {
		return fmt.Errorf("invalid message: missing type")
	}

	payload, ok := msg.Values["payload"].(string)
	if !ok {
		return fmt.Errorf("invalid message: missing payload")
	}

	switch JobType(jobType) {
	case JobTypeRecipient:
		var job RecipientJob
		if err := json.Unmarshal([]byte(payload), &job); err != nil {
			return fmt.Errorf("failed to unmarshal recipient job: %w", err)
		}
		c.log.Debug("Processing recipient job", "campaign_id", job.CampaignID, "recipient_id", job.RecipientID, "message_id", msg.ID)
		return handler.HandleRecipientJob(ctx, &job)

	default:
		return fmt.Errorf("unknown job type: %s", jobType)
	}
}

// Close closes the consumer connection
func (c *RedisConsumer) Close() error {
	return nil // Redis client is managed externally
}
