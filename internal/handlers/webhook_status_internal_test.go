package handlers

import (
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/shridarpatil/whatomate/internal/config"
	"github.com/shridarpatil/whatomate/internal/models"
	"github.com/shridarpatil/whatomate/test/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStatusesBelow(t *testing.T) {
	t.Parallel()

	assert.ElementsMatch(t,
		[]models.MessageStatus{models.MessageStatusReceived, models.MessageStatusPending},
		statusesBelow(models.MessageStatusSent))
	assert.ElementsMatch(t,
		[]models.MessageStatus{models.MessageStatusReceived, models.MessageStatusPending, models.MessageStatusSent, models.MessageStatusDelivered},
		statusesBelow(models.MessageStatusRead))
	assert.ElementsMatch(t,
		[]models.MessageStatus{models.MessageStatusReceived, models.MessageStatusPending, models.MessageStatusSent, models.MessageStatusDelivered, models.MessageStatusRead},
		statusesBelow(models.MessageStatusFailed), "failed overrides every non-failed status")
	assert.NotContains(t, statusesBelow(models.MessageStatusDelivered), models.MessageStatusRead,
		"a read message must never regress to delivered")
}

// TestUpdateMessageStatus_ConcurrentDuplicatesCountOnce reproduces Meta
// delivering the same status callback several times at once. Every copy used
// to pass the read-then-write progression check and increment the campaign
// counter; the conditional UPDATE lets exactly one through.
func TestUpdateMessageStatus_ConcurrentDuplicatesCountOnce(t *testing.T) {
	db := testutil.SetupTestDB(t)
	app := &App{
		DB:     db,
		Log:    testutil.NopLogger(),
		Config: &config.Config{},
	}

	org := testutil.CreateTestOrganization(t, db)
	contact := testutil.CreateTestContact(t, db, org.ID)
	user := testutil.CreateTestUser(t, db, org.ID)
	template := testutil.CreateTestTemplate(t, db, org.ID, "status-race-acc")

	campaign := &models.BulkMessageCampaign{
		BaseModel:       models.BaseModel{ID: uuid.New()},
		OrganizationID:  org.ID,
		WhatsAppAccount: "status-race-acc",
		Name:            "status race",
		TemplateID:      template.ID,
		Status:          models.CampaignStatusProcessing,
		SentCount:       1,
		CreatedBy:       user.ID,
	}
	require.NoError(t, db.Create(campaign).Error)

	wamid := "wamid.status-race-" + uuid.New().String()
	msg := &models.Message{
		BaseModel:         models.BaseModel{ID: uuid.New()},
		OrganizationID:    org.ID,
		WhatsAppAccount:   "status-race-acc",
		ContactID:         contact.ID,
		WhatsAppMessageID: wamid,
		Direction:         models.DirectionOutgoing,
		MessageType:       models.MessageTypeTemplate,
		Status:            models.MessageStatusSent,
		Metadata:          models.JSONB{"campaign_id": campaign.ID.String()},
	}
	require.NoError(t, db.Create(msg).Error)

	const copies = 8
	var wg sync.WaitGroup
	for i := 0; i < copies; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			app.updateMessageStatus(uuid.Nil, wamid, string(models.MessageStatusDelivered), nil)
		}()
	}
	wg.Wait()

	var updatedMsg models.Message
	require.NoError(t, db.Where("id = ?", msg.ID).First(&updatedMsg).Error)
	assert.Equal(t, models.MessageStatusDelivered, updatedMsg.Status)

	var updatedCampaign models.BulkMessageCampaign
	require.NoError(t, db.Where("id = ?", campaign.ID).First(&updatedCampaign).Error)
	assert.Equal(t, 1, updatedCampaign.DeliveredCount, "duplicate callbacks must not inflate delivered_count")

	// A later, lower-priority callback (delivered after read) must not regress the row.
	app.updateMessageStatus(uuid.Nil, wamid, string(models.MessageStatusRead), nil)
	app.updateMessageStatus(uuid.Nil, wamid, string(models.MessageStatusDelivered), nil)
	require.NoError(t, db.Where("id = ?", msg.ID).First(&updatedMsg).Error)
	assert.Equal(t, models.MessageStatusRead, updatedMsg.Status)
	require.NoError(t, db.Where("id = ?", campaign.ID).First(&updatedCampaign).Error)
	assert.Equal(t, 1, updatedCampaign.ReadCount)
	assert.Equal(t, 1, updatedCampaign.DeliveredCount)
}

// A status callback resolved to one organization must not update a message
// row that belongs to another organization, even when the WhatsApp message id
// collides.
func TestUpdateMessageStatus_ScopedToOrganization(t *testing.T) {
	db := testutil.SetupTestDB(t)
	app := &App{DB: db, Log: testutil.NopLogger(), Config: &config.Config{}}

	orgA := testutil.CreateTestOrganization(t, db)
	orgB := testutil.CreateTestOrganization(t, db)
	contactA := testutil.CreateTestContact(t, db, orgA.ID)
	wamid := "wamid.scoped-" + uuid.New().String()
	msg := &models.Message{
		BaseModel:         models.BaseModel{ID: uuid.New()},
		OrganizationID:    orgA.ID,
		WhatsAppAccount:   "scoped-acc",
		ContactID:         contactA.ID,
		WhatsAppMessageID: wamid,
		Direction:         models.DirectionOutgoing,
		MessageType:       models.MessageTypeText,
		Status:            models.MessageStatusSent,
	}
	require.NoError(t, db.Create(msg).Error)

	app.updateMessageStatus(orgB.ID, wamid, string(models.MessageStatusDelivered), nil)
	var got models.Message
	require.NoError(t, db.Where("id = ?", msg.ID).First(&got).Error)
	assert.Equal(t, models.MessageStatusSent, got.Status, "another org's callback must not touch the row")

	app.updateMessageStatus(orgA.ID, wamid, string(models.MessageStatusDelivered), nil)
	require.NoError(t, db.Where("id = ?", msg.ID).First(&got).Error)
	assert.Equal(t, models.MessageStatusDelivered, got.Status)
}
