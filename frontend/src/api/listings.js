/**
 * listings.js — thin API client for the Go listings backend.
 *
 * All functions return the parsed JSON directly and throw on non-2xx responses
 * so callers can handle errors in a single try/catch.
 */

const BASE = import.meta.env.VITE_API_BASE_URL ?? ''
// 8 s gives a comfortable margin below the server's 10 s WriteTimeout so the
// JS abort always fires first and the user gets a clean error message.
const REQUEST_TIMEOUT_MS = 8_000

/**
 * Low-level fetch wrapper. Throws an Error with the server message on failure.
 * Aborts automatically after REQUEST_TIMEOUT_MS to prevent infinite loading states.
 */
async function request(path, options = {}) {
  const controller = new AbortController()
  const timeoutId = setTimeout(() => controller.abort(), REQUEST_TIMEOUT_MS)

  let res
  try {
    res = await fetch(`${BASE}${path}`, {
      signal: controller.signal,
      headers: { 'Content-Type': 'application/json', ...options.headers },
      ...options,
    })
  } catch (err) {
    if (err.name === 'AbortError') {
      throw new Error('Request timed out. Please try again.')
    }
    throw err
  } finally {
    clearTimeout(timeoutId)
  }

  if (res.status === 204) return null // DELETE success

  const data = await res.json()

  if (!res.ok) {
    // The backend returns either {"error":"..."} or {"errors":[...]}
    const msg =
      data?.error ??
      data?.errors?.map((e) => `${e.field}: ${e.message}`).join(', ') ??
      `HTTP ${res.status}`
    throw Object.assign(new Error(msg), { status: res.status, data })
  }

  return data
}

/**
 * Build a query string from a plain object, omitting null / undefined / '' values.
 */
function buildQuery(params) {
  const q = new URLSearchParams()
  for (const [k, v] of Object.entries(params)) {
    if (v !== null && v !== undefined && v !== '') q.set(k, v)
  }
  const s = q.toString()
  return s ? `?${s}` : ''
}

// Upload API
/**
 * POST /api/upload
 * Uploads an image file and returns its public URL.
 * Uses a raw fetch rather than request() because multipart/form-data must not
 * have a manually-set Content-Type header — the browser sets it with the boundary.
 * @param {File} file
 * @returns {Promise<string>} the public URL of the uploaded image
 */
export async function uploadImage(file) {
  const form = new FormData()
  form.append('file', file)

  const controller = new AbortController()
  const timeoutId = setTimeout(() => controller.abort(), REQUEST_TIMEOUT_MS)

  let res
  try {
    res = await fetch(`${BASE}/api/upload`, {
      method: 'POST',
      body: form,
      signal: controller.signal,
    })
  } catch (err) {
    if (err.name === 'AbortError') throw new Error('Upload timed out. Please try again.')
    throw err
  } finally {
    clearTimeout(timeoutId)
  }

  const data = await res.json()
  if (!res.ok) {
    throw new Error(data?.error ?? `Upload failed (HTTP ${res.status})`)
  }
  return data.url
}

// Listings API
/**
 * GET /api/listings
 * @param {Object} params - limit, offset, search, category, status
 * @returns {Promise<{listings: Array, total: number, limit: number, offset: number}>}
 */
export function getListings(params = {}) {
  return request(`/api/listings/${buildQuery(params)}`)
}

/**
 * GET /api/listings/:id
 * @param {number} id
 * @returns {Promise<Object>}
 */
export function getListing(id) {
  return request(`/api/listings/${id}/`)
}

/**
 * POST /api/listings
 * @param {Object} payload - listing fields
 * @returns {Promise<Object>}
 */
export function createListing(payload) {
  return request('/api/listings/', {
    method: 'POST',
    body: JSON.stringify(payload),
  })
}

/**
 * PUT /api/listings/:id
 * @param {number} id
 * @param {Object} payload - listing fields
 * @returns {Promise<Object>}
 */
export function updateListing(id, payload) {
  return request(`/api/listings/${id}/`, {
    method: 'PUT',
    body: JSON.stringify(payload),
  })
}

/**
 * DELETE /api/listings/:id
 * @param {number} id
 * @returns {Promise<null>}
 */
export function deleteListing(id) {
  return request(`/api/listings/${id}/`, { method: 'DELETE' })
}
