import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest'
import { relativeTime } from '../utils/listingDisplay.js'

// Covers the three distinct output branches of relativeTime: missing input,
// a recent timestamp, and a date old enough to show a full calendar string.
describe('relativeTime', () => {
  let now

  beforeEach(() => {
    now = Date.now()
    vi.spyOn(Date, 'now').mockReturnValue(now)
  })

  afterEach(() => vi.restoreAllMocks())

  // Listing cards must never crash or show blank when date_posted is absent.
  it('returns "unknown date" for a missing value', () => {
    expect(relativeTime(null)).toBe('unknown date')
    expect(relativeTime(undefined)).toBe('unknown date')
  })

  // Confirms the relative-time branch runs and produces a readable label.
  it('returns a relative string for a recent timestamp', () => {
    const recent = new Date(now - 5 * 60_000).toISOString()
    expect(relativeTime(recent)).toBe('5 min ago')
  })

  // Older listings fall through to a formatted calendar date, not a relative label.
  it('returns a formatted date for timestamps older than 30 days', () => {
    const result = relativeTime(new Date('2020-01-15').toISOString())
    expect(result).toMatch(/Jan/)
    expect(result).toMatch(/2020/)
  })
})
