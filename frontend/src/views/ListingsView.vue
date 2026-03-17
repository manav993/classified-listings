<template>
  <div class="page">

    <!-- Page header -->
    <div class="page-header">
      <div>
        <h1 class="page-title">Marketplace Listings</h1>
        <p class="page-subtitle">Browse, search, and manage classified listings with ease.</p>
      </div>
      <button class="btn btn-primary" @click="openCreate">
        <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor"
          stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
          <line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/>
        </svg>
        New Listing
      </button>
    </div>

    <!-- Filters -->
    <FilterBar :model-value="currentFilters" @update:model-value="applyFilters" />

    <!-- Error -->
    <p v-if="error" class="status-msg error-msg">
      {{ error }} -
      <button class="retry-btn" @click="fetchListings">retry</button>
    </p>

    <!-- Loading -->
    <div v-else-if="loading" class="loading-grid">
      <div v-for="n in 6" :key="n" class="skeleton-card" />
    </div>

    <!-- Empty -->
    <div v-else-if="listings.length === 0" class="empty-state">
      <div class="empty-icon">
        <svg width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="currentColor"
          stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
          <circle cx="11" cy="11" r="8"/><path d="M21 21l-4.35-4.35"/>
        </svg>
      </div>
      <p class="empty-title">No listings found</p>
      <p class="empty-hint">Try adjusting your filters or create a new listing.</p>
    </div>

    <!-- Cards -->
    <div v-else class="cards-grid">
      <ListingCard
        v-for="l in listings"
        :key="l.id"
        :listing="l"
        @edit="openEdit"
        @delete="openDelete"
      />
    </div>

    <!-- Footer: count + pagination -->
    <div v-if="!loading && !error && total > 0" class="list-footer">
      <p class="results-caption">
        Showing <strong>{{ rangeFrom }}</strong> to <strong>{{ rangeTo }}</strong>
        of <strong>{{ total }}</strong> {{ total === 1 ? 'listing' : 'listings' }}
      </p>
      <Pagination
        v-if="totalPages > 1"
        :current="page"
        :total-pages="totalPages"
        @change="goToPage"
      />
    </div>

    <!-- Create / Edit modal -->
    <ListingModal
      v-if="modal.open"
      :listing="modal.listing"
      :saving="modal.saving"
      :api-error="modal.apiError"
      @close="closeModal"
      @save="handleSave"
    />

    <!-- Delete confirm -->
    <ConfirmDialog
      v-if="deleteTarget"
      :listing="deleteTarget"
      :loading="deleteLoading"
      @confirm="handleDelete"
      @cancel="deleteTarget = null"
    />

    <Toast :toast="toast" />

  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import FilterBar     from '../components/FilterBar.vue'
import Pagination    from '../components/Pagination.vue'
import ListingCard   from '../components/ListingCard.vue'
import ListingModal  from '../components/ListingModal.vue'
import ConfirmDialog from '../components/ConfirmDialog.vue'
import Toast         from '../components/Toast.vue'
import { useListings, PAGE_SIZE } from '../composables/useListings.js'
import { useToast } from '../composables/useToast.js'

const {
  listings, total, loading, error,
  search, category, status, page, totalPages,
  fetchListings, applyFilters, goToPage, saveListing, removeListing,
} = useListings()

const currentFilters = computed(() => ({
  search:   search.value,
  category: category.value,
  status:   status.value,
}))

const rangeFrom = computed(() =>
  total.value === 0 ? 0 : (page.value - 1) * PAGE_SIZE + 1
)
const rangeTo = computed(() => Math.min(page.value * PAGE_SIZE, total.value))

const { toast, showToast } = useToast()

// Modal
const modal = ref({ open: false, listing: null, saving: false, apiError: null })

function openCreate() {
  modal.value = { open: true, listing: null, saving: false, apiError: null }
}
function openEdit(listing) {
  modal.value = { open: true, listing, saving: false, apiError: null }
}
function closeModal() {
  modal.value.open = false
}
async function handleSave(payload) {
  const isEdit = !!modal.value.listing
  modal.value.saving   = true
  modal.value.apiError = null
  try {
    await saveListing(payload, modal.value.listing?.id ?? null)
    closeModal()
    showToast(isEdit ? 'Listing updated successfully.' : 'Listing created successfully.')
  } catch (e) {
    modal.value.apiError = e.message
  } finally {
    modal.value.saving = false
  }
}

// Delete
const deleteTarget  = ref(null)
const deleteLoading = ref(false)

function openDelete(listing) {
  deleteTarget.value = listing
}
async function handleDelete() {
  deleteLoading.value = true
  try {
    await removeListing(deleteTarget.value.id)
    deleteTarget.value = null
    showToast('Listing deleted.')
  } catch (e) {
    showToast(`Delete failed: ${e.message}`, 'error')
  } finally {
    deleteLoading.value = false
  }
}

onMounted(fetchListings)
</script>

<style scoped>
.page {
  max-width: 1140px;
  margin: 0 auto;
  padding: 32px 24px 48px;
}

/* Page header */
.page-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
  margin-bottom: 28px;
}

.page-title {
  font-size: 26px;
  font-weight: 800;
  color: #111827;
  letter-spacing: -0.02em;
  line-height: 1.2;
}

.page-subtitle {
  margin-top: 4px;
  font-size: 14px;
  color: #6b7280;
}

/* States */
.status-msg {
  text-align: center;
  padding: 48px 0;
  color: #6b7280;
}

.error-msg { color: #ef4444; }

.retry-btn {
  background: none;
  border: none;
  color: #059669;
  text-decoration: underline;
  cursor: pointer;
  padding: 0;
  font-size: inherit;
}

/* Skeleton loading */
.loading-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
  gap: 1.1rem;
}

.skeleton-card {
  background: var(--color-surface);
  border-radius: var(--radius-card);
  height: 340px;
  border: 1px solid var(--color-border);
  animation: pulse 1.4s ease-in-out infinite;
}

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50%       { opacity: 0.5; }
}

/* Empty state */
.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 72px 24px;
  gap: 10px;
}

.empty-icon {
  width: 60px;
  height: 60px;
  background: #f3f4f6;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #9ca3af;
  margin-bottom: 4px;
}

.empty-title {
  font-size: 16px;
  font-weight: 600;
  color: #374151;
}

.empty-hint {
  font-size: 14px;
  color: #9ca3af;
}

/* Cards grid */
.cards-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
  gap: 1.1rem;
}

/* List footer */
.list-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  flex-wrap: wrap;
  gap: 12px;
  margin-top: 28px;
  padding-top: 20px;
  border-top: 1px solid var(--color-border);
}

.results-caption {
  font-size: 13.5px;
  color: #6b7280;
}

.results-caption strong {
  font-weight: 600;
  color: #374151;
}

</style>
