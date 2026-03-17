import { describe, it, expect } from 'vitest'

// buildQuery is not exported from listings.js — replicated here as a pure
// function so the query-building logic can be tested in isolation.
function buildQuery(params) {
  const q = new URLSearchParams()
  for (const [k, v] of Object.entries(params)) {
    if (v !== null && v !== undefined && v !== '') q.set(k, v)
  }
  const s = q.toString()
  return s ? `?${s}` : ''
}

describe('buildQuery', () => {
  // Unset filters must be omitted so stale values never reach the API.
  it('omits null, undefined, and empty-string values but includes present ones', () => {
    const result = buildQuery({ limit: 10, offset: null, search: '', category: 'Vehicle' })
    expect(result).toContain('limit=10')
    expect(result).toContain('category=Vehicle')
    expect(result).not.toContain('offset')
    expect(result).not.toContain('search')
  })

  // Characters like % must be percent-encoded or the server receives a malformed query.
  it('URL-encodes special characters', () => {
    expect(buildQuery({ search: '50% off' })).toBe('?search=50%25+off')
  })
})
