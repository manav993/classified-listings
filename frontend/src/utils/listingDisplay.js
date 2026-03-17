// Pure display helpers for listings - no Vue dependency, fully unit-testable.

export const CATEGORY_GRADIENTS = {
  Property:    'linear-gradient(135deg, #d1fae5 0%, #6ee7b7 100%)',
  Vehicle:     'linear-gradient(135deg, #dbeafe 0%, #93c5fd 100%)',
  Electronics: 'linear-gradient(135deg, #ede9fe 0%, #c4b5fd 100%)',
}

export const CATEGORY_EMOJIS = {
  Property: '🏠', Vehicle: '🚗', Electronics: '💻',
}

export const DEFAULT_GRADIENT = 'linear-gradient(135deg, #f3f4f6, #e5e7eb)'
export const DEFAULT_EMOJI     = '📦'

/**
 * Returns a human-readable relative time string for a given date string.
 * Falls back to a formatted date for older items.
 *
 * @param {string|null} dateStr - ISO date string
 * @returns {string}
 */
export function relativeTime(dateStr) {
  if (!dateStr) return 'unknown date'
  const diff = (Date.now() - new Date(dateStr)) / 1000
  if (diff < 60)      return 'just now'
  if (diff < 3600)    return `${Math.floor(diff / 60)} min ago`
  if (diff < 86400)   return `${Math.floor(diff / 3600)} hours ago`
  if (diff < 604800)  return `${Math.floor(diff / 86400)} days ago`
  if (diff < 2592000) return `${Math.floor(diff / 604800)} weeks ago`
  return new Date(dateStr).toLocaleDateString('en-US', {
    month: 'short', day: 'numeric', year: 'numeric',
  })
}
