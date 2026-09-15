/**
 * Single source of truth for status presentation (message delivery,
 * campaign lifecycle). Views previously kept their own copies of these
 * colour maps, several of them diverging for the same status.
 *
 * Every entry pairs a Badge variant with an i18n key so the status is always
 * conveyed by text as well as colour.
 */
import {
  AlertCircle,
  Check,
  CheckCheck,
  CircleDashed,
  CircleSlash,
  Clock,
  Loader2,
  Pause,
  Send,
} from 'lucide-vue-next'
import type { Component } from 'vue'

export type BadgeVariant = 'default' | 'secondary' | 'destructive' | 'outline' | 'success' | 'warning' | 'info' | 'active'

export interface StatusMeta {
  /** Badge variant to render with */
  variant: BadgeVariant
  /** i18n key of the human label */
  labelKey: string
  /** Optional icon reinforcing the state */
  icon?: Component
}

const messageStatuses: Record<string, StatusMeta> = {
  pending: { variant: 'secondary', labelKey: 'chat.statusPending', icon: Clock },
  sent: { variant: 'secondary', labelKey: 'chat.statusSent', icon: Check },
  delivered: { variant: 'info', labelKey: 'chat.statusDelivered', icon: CheckCheck },
  read: { variant: 'success', labelKey: 'chat.statusRead', icon: CheckCheck },
  failed: { variant: 'destructive', labelKey: 'chat.statusFailed', icon: AlertCircle },
  received: { variant: 'secondary', labelKey: 'chat.statusDelivered', icon: CheckCheck },
}

export function messageStatusMeta(status: string | undefined): StatusMeta {
  return messageStatuses[status || 'pending'] || messageStatuses.pending
}

const campaignStatuses: Record<string, StatusMeta> = {
  draft: { variant: 'secondary', labelKey: 'campaigns.draft', icon: CircleDashed },
  scheduled: { variant: 'info', labelKey: 'campaigns.scheduled', icon: Clock },
  queued: { variant: 'info', labelKey: 'campaigns.queued', icon: Clock },
  processing: { variant: 'info', labelKey: 'campaigns.processing', icon: Loader2 },
  paused: { variant: 'warning', labelKey: 'campaigns.paused', icon: Pause },
  completed: { variant: 'success', labelKey: 'campaigns.completed', icon: Check },
  cancelled: { variant: 'secondary', labelKey: 'campaigns.cancelled', icon: CircleSlash },
  failed: { variant: 'destructive', labelKey: 'campaigns.failed', icon: AlertCircle },
}

export function campaignStatusMeta(status: string | undefined): StatusMeta {
  return campaignStatuses[status || 'draft'] || { variant: 'secondary', labelKey: 'common.status', icon: Send }
}

/** Direction badge for message lists (incoming vs outgoing). */
export function directionMeta(direction: string | undefined): StatusMeta {
  return direction === 'incoming'
    ? { variant: 'success', labelKey: 'chat.customer' }
    : { variant: 'info', labelKey: 'chat.you' }
}
