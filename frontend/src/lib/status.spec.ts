import { describe, expect, it } from 'vitest'
import { campaignStatusMeta, messageStatusMeta, directionMeta } from './status'

describe('messageStatusMeta', () => {
  it('maps every delivery state to a badge variant and a label key', () => {
    for (const status of ['pending', 'sent', 'delivered', 'read', 'failed']) {
      const meta = messageStatusMeta(status)
      expect(meta.variant).toBeTruthy()
      expect(meta.labelKey).toMatch(/^chat\.status/)
      expect(meta.icon).toBeTruthy()
    }
  })

  it('uses a distinct label for read vs delivered so status is not colour-only', () => {
    expect(messageStatusMeta('read').labelKey).not.toBe(messageStatusMeta('delivered').labelKey)
  })

  it('falls back to pending for unknown or missing values', () => {
    expect(messageStatusMeta(undefined)).toEqual(messageStatusMeta('pending'))
    expect(messageStatusMeta('bogus')).toEqual(messageStatusMeta('pending'))
  })
})

describe('campaignStatusMeta', () => {
  it('marks failure as destructive and completion as success', () => {
    expect(campaignStatusMeta('failed').variant).toBe('destructive')
    expect(campaignStatusMeta('completed').variant).toBe('success')
    expect(campaignStatusMeta('paused').variant).toBe('warning')
  })

  it('never throws for unexpected statuses', () => {
    expect(campaignStatusMeta('weird').labelKey).toBe('common.status')
  })
})

describe('directionMeta', () => {
  it('distinguishes incoming from outgoing', () => {
    expect(directionMeta('incoming').variant).not.toBe(directionMeta('outgoing').variant)
  })
})
