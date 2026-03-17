/**
 * useListings — composable that owns all listing state and actions.
 *
 * Components simply import this and call the exposed methods; they never
 * talk to the API layer directly.
 */
import { ref, computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { getListings, createListing, updateListing, deleteListing } from '../api/listings.js'

export const PAGE_SIZE = 9

export function useListings() {
  const router = useRouter()
  const route  = useRoute()

  // State
  const listings = ref([])
  const total    = ref(0)
  const loading  = ref(false)
  const error    = ref(null)

  // Filters — initialised from URL query params so the page is bookmarkable.
  const search   = ref(route.query.search   ?? '')
  const category = ref(route.query.category ?? '')
  const status   = ref(route.query.status   ?? '')
  const rawPage  = parseInt(route.query.page, 10)
  const page     = ref(isNaN(rawPage) || rawPage < 1 ? 1 : rawPage)

  // Derived
  const totalPages = computed(() => Math.ceil(total.value / PAGE_SIZE) || 1)
  const offset     = computed(() => (page.value - 1) * PAGE_SIZE)

  // Helpers
  /** Push current filter state into the URL so it's preserved on refresh. */
  function syncUrl() {
    router.replace({
      query: {
        ...(search.value   && { search:   search.value }),
        ...(category.value && { category: category.value }),
        ...(status.value   && { status:   status.value }),
        ...(page.value > 1 && { page:     page.value }),
      },
    })
  }

  // Incremented on every fetch so stale responses from earlier calls are silently
  // discarded when a newer fetch is already in flight.
  let fetchSeq = 0

  // Actions
  async function fetchListings() {
    const seq = ++fetchSeq
    loading.value = true
    error.value   = null
    try {
      const data = await getListings({
        limit:    PAGE_SIZE,
        offset:   offset.value,
        search:   search.value   || undefined,
        category: category.value || undefined,
        status:   status.value   || undefined,
      })
      if (seq !== fetchSeq) return
      listings.value = data.listings
      total.value    = data.total
    } catch (e) {
      if (seq !== fetchSeq) return
      error.value = e.message
    } finally {
      if (seq === fetchSeq) loading.value = false
    }
  }

  /** Apply new filters and reset to page 1. */
  function applyFilters(filters) {
    search.value   = filters.search   ?? ''
    category.value = filters.category ?? ''
    status.value   = filters.status   ?? ''
    page.value     = 1
    syncUrl()
    fetchListings()
  }

  function goToPage(n) {
    page.value = n
    syncUrl()
    fetchListings()
  }

  async function saveListing(payload, id = null) {
    if (id !== null) {
      await updateListing(id, payload)
      await fetchListings()
    } else {
      await createListing(payload)
      // Clear all filters after creation so the new listing is always visible.
      search.value   = ''
      category.value = ''
      status.value   = ''
      page.value     = 1
      syncUrl()
      await fetchListings()
    }
  }

  async function removeListing(id) {
    await deleteListing(id)
    // If the last item on a page was deleted, step back one page.
    if (listings.value.length === 1 && page.value > 1) {
      page.value -= 1
      syncUrl()
    }
    await fetchListings()
  }

  return {
    listings,
    total,
    loading,
    error,
    search,
    category,
    status,
    page,
    totalPages,
    fetchListings,
    applyFilters,
    goToPage,
    saveListing,
    removeListing,
  }
}
