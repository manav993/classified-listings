// Single source of truth for listing domain constants used across components.
// Adding a new category or status here propagates to every dropdown and validator.
export const CATEGORIES = ['Property', 'Vehicle', 'Electronics']
export const STATUSES   = ['Active', 'Inactive']

// Character limits — must match the backend validator to avoid split-brain errors.
export const TITLE_MIN_LENGTH       = 3
export const TITLE_MAX_LENGTH       = 100
export const DESCRIPTION_MIN_LENGTH = 20
export const DESCRIPTION_MAX_LENGTH = 1000

// Price bounds — must match the backend validator to avoid split-brain errors.
export const PRICE_MIN =          0.01
export const PRICE_MAX = 10_000_000
