import { describe, it, expect } from 'vitest'
import { TITLE_MIN_LENGTH, TITLE_MAX_LENGTH, DESCRIPTION_MIN_LENGTH, DESCRIPTION_MAX_LENGTH, PRICE_MAX } from '../constants/listing.js'

// validate() lives inside ListingModal.vue. It is replicated here as a plain
// function so the rules can be tested without mounting the full component.
// Keep this in sync with the validate() function in ListingModal.vue.
function validate(form) {
  const e = {}
  const trimmedTitle = form.title?.trim() ?? ''
  if (!trimmedTitle)                                        e.title = 'Title is required'
  else if (trimmedTitle.length < TITLE_MIN_LENGTH)          e.title = `Title must be at least ${TITLE_MIN_LENGTH} characters`
  else if (form.title.length > TITLE_MAX_LENGTH)            e.title = `Title must be ${TITLE_MAX_LENGTH} characters or fewer`

  const trimmedDesc = form.description?.trim() ?? ''
  if (!trimmedDesc)                                         e.description = 'Description is required'
  else if (trimmedDesc.length < DESCRIPTION_MIN_LENGTH)    e.description = `Description must be at least ${DESCRIPTION_MIN_LENGTH} characters`
  else if (form.description.length > DESCRIPTION_MAX_LENGTH) e.description = `Description must be ${DESCRIPTION_MAX_LENGTH} characters or fewer`

  if (!form.price || Number(form.price) <= 0) e.price = 'Price must be greater than zero'
  else if (Number(form.price) > PRICE_MAX)    e.price = `Price must not exceed £${PRICE_MAX.toLocaleString()}`
  if (!form.category) e.category = 'Please select a category'
  if (!form.status)   e.status   = 'Please select a status'
  return e
}

const VALID = {
  title: 'Mountain Bike',
  description: 'Good condition, barely used.',
  price: 150,
  category: 'Vehicle',
  status: 'Active',
}

describe('ListingModal validate()', () => {
  // Baseline: a correctly filled form must never be incorrectly rejected.
  it('returns no errors for a fully valid form', () => {
    expect(validate(VALID)).toEqual({})
  })

  // Spaces-only input must be caught; without trim() it silently passes as non-empty.
  it('rejects whitespace-only title and description', () => {
    expect(validate({ ...VALID, title: '   ' }).title).toBeDefined()
    expect(validate({ ...VALID, description: '   ' }).description).toBeDefined()
  })

  // Single characters like "a" or "b" must be rejected as too short.
  it('rejects title under minimum length but accepts exactly the minimum', () => {
    expect(validate({ ...VALID, title: 'ab' }).title).toBeDefined()
    expect(validate({ ...VALID, title: 'abc' }).title).toBeUndefined()
  })

  // Short descriptions like "Nice" must be rejected; 20+ characters must pass.
  it('rejects description under minimum length but accepts exactly the minimum', () => {
    expect(validate({ ...VALID, description: 'Too short' }).description).toBeDefined()
    expect(validate({ ...VALID, description: 'a'.repeat(DESCRIPTION_MIN_LENGTH) }).description).toBeUndefined()
  })

  // Price rule is strictly > 0: zero must be rejected and 0.01 must be accepted.
  it('rejects price = 0 but accepts 0.01 (lower boundary)', () => {
    expect(validate({ ...VALID, price: 0 }).price).toBeDefined()
    expect(validate({ ...VALID, price: 0.01 }).price).toBeUndefined()
  })

  // Upper boundary: exactly at the cap passes; one cent over must be rejected.
  it('rejects price above max but accepts exactly the max (upper boundary)', () => {
    expect(validate({ ...VALID, price: PRICE_MAX }).price).toBeUndefined()
    expect(validate({ ...VALID, price: PRICE_MAX + 0.01 }).price).toBeDefined()
  })

  // The create form initialises price as null; ensure it is caught before submit.
  it('rejects null price (create form starts with null)', () => {
    expect(validate({ ...VALID, price: null }).price).toBeDefined()
  })

  // Titles at or under the limit pass; one character over must be rejected.
  it('rejects title over 100 chars but accepts exactly 100', () => {
    const at  = 'a'.repeat(TITLE_MAX_LENGTH)
    const over = 'a'.repeat(TITLE_MAX_LENGTH + 1)
    expect(validate({ ...VALID, title: at }).title).toBeUndefined()
    expect(validate({ ...VALID, title: over }).title).toBeDefined()
  })

  // Descriptions at or under the limit pass; one character over must be rejected.
  it('rejects description over 1000 chars but accepts exactly 1000', () => {
    const at   = 'a'.repeat(DESCRIPTION_MAX_LENGTH)
    const over = 'a'.repeat(DESCRIPTION_MAX_LENGTH + 1)
    expect(validate({ ...VALID, description: at }).description).toBeUndefined()
    expect(validate({ ...VALID, description: over }).description).toBeDefined()
  })

  // All field errors must be returned together so the user can fix everything in one go.
  it('collects all errors at once rather than stopping at the first', () => {
    const result = validate({ title: '', description: '', price: 0, category: '', status: '' })
    expect(Object.keys(result).length).toBe(5)
  })
})
