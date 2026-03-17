<template>
  <div class="page">

    <!-- Back -->
    <button class="back-btn" @click="goBack">
      <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor"
        stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
        <polyline points="15 18 9 12 15 6"/>
      </svg>
      Back to listings
    </button>

    <!-- Loading skeleton -->
    <div v-if="loading" class="detail-skeleton">
      <div class="skel-image" />
      <div class="skel-body">
        <div class="skel-line w-20" />
        <div class="skel-line w-60" />
        <div class="skel-line w-40" />
        <div class="skel-line w-full" />
        <div class="skel-line w-full" />
        <div class="skel-line w-80" />
      </div>
    </div>

    <!-- Error -->
    <div v-else-if="error" class="error-state">
      <div class="error-icon">
        <svg width="28" height="28" viewBox="0 0 24 24" fill="none" stroke="currentColor"
          stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
          <circle cx="12" cy="12" r="10"/>
          <line x1="12" y1="8" x2="12" y2="12"/>
          <line x1="12" y1="16" x2="12.01" y2="16"/>
        </svg>
      </div>
      <p class="error-title">{{ error }}</p>
      <button class="btn btn-primary" @click="$router.push('/')">Back to listings</button>
    </div>

    <!-- Detail layout -->
    <div v-else-if="listing" class="detail-layout">

      <!-- Left: Image -->
      <div class="detail-image-wrap">
        <img
          v-if="listing.image_url"
          :src="listing.image_url"
          :alt="listing.title"
          class="detail-img"
        />
        <div v-else class="detail-placeholder" :style="{ background: placeholderGradient }">
          <span class="placeholder-icon">{{ categoryEmoji }}</span>
        </div>
      </div>

      <!-- Right: Info -->
      <div class="detail-info">

        <div class="badge-row">
          <StatusBadge :status="listing.status" />
          <span class="chip">{{ listing.category }}</span>
        </div>

        <h1 class="detail-title">{{ listing.title }}</h1>
        <p class="detail-price">£{{ listing.price.toLocaleString() }}</p>

        <div class="divider" />

        <p class="detail-desc">{{ listing.description }}</p>

        <div class="divider" />

        <dl class="meta-list">
          <div class="meta-row">
            <dt>Category</dt>
            <dd>{{ listing.category }}</dd>
          </div>
          <div class="meta-row">
            <dt>Status</dt>
            <dd><StatusBadge :status="listing.status" sm /></dd>
          </div>
          <div class="meta-row">
            <dt>Date posted</dt>
            <dd>{{ formattedDate }}</dd>
          </div>
          <div class="meta-row">
            <dt>Listing ID</dt>
            <dd>{{ listing.id }}</dd>
          </div>
        </dl>

        <div class="detail-actions">
          <button class="btn btn-secondary" @click="openEdit">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor"
              stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <path d="M11 4H4a2 2 0 00-2 2v14a2 2 0 002 2h14a2 2 0 002-2v-7"/>
              <path d="M18.5 2.5a2.121 2.121 0 013 3L12 15l-4 1 1-4 9.5-9.5z"/>
            </svg>
            Edit listing
          </button>
          <button class="btn btn-danger" @click="deleteOpen = true">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor"
              stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <polyline points="3 6 5 6 21 6"/>
              <path d="M19 6v14a2 2 0 01-2 2H7a2 2 0 01-2-2V6m3 0V4a1 1 0 011-1h4a1 1 0 011 1v2"/>
            </svg>
            Delete
          </button>
        </div>

      </div>
    </div>

    <!-- Edit modal -->
    <ListingModal
      v-if="modal.open"
      :listing="modal.listing"
      :saving="modal.saving"
      :api-error="modal.apiError"
      @close="modal.open = false"
      @save="handleSave"
    />

    <!-- Delete confirm -->
    <ConfirmDialog
      v-if="deleteOpen"
      :listing="listing"
      :loading="deleteLoading"
      @confirm="handleDelete"
      @cancel="deleteOpen = false"
    />

    <Toast :toast="toast" />

  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import ListingModal  from '../components/ListingModal.vue'
import ConfirmDialog from '../components/ConfirmDialog.vue'
import Toast         from '../components/Toast.vue'
import StatusBadge   from '../components/StatusBadge.vue'
import { getListing, updateListing, deleteListing } from '../api/listings.js'
import { useToast } from '../composables/useToast.js'
import { CATEGORY_GRADIENTS, CATEGORY_EMOJIS, DEFAULT_GRADIENT, DEFAULT_EMOJI } from '../utils/listingDisplay.js'

const route  = useRoute()
const router = useRouter()

// If there is no browser history (direct link or bookmark), fall back to the
// listings page rather than navigating the user out of the app.
function goBack() {
  if (window.history.length > 1) {
    router.back()
  } else {
    router.push('/')
  }
}

const listing      = ref(null)
const loading      = ref(true)
const error        = ref(null)

// Fetch
async function fetchListing() {
  loading.value = true
  error.value   = null
  try {
    listing.value = await getListing(route.params.id)
  } catch (e) {
    error.value = e.status === 404 ? 'Listing not found.' : e.message
  } finally {
    loading.value = false
  }
}

onMounted(fetchListing)

// Computed display
const placeholderGradient = computed(
  () => CATEGORY_GRADIENTS[listing.value?.category] ?? DEFAULT_GRADIENT
)
const categoryEmoji = computed(
  () => CATEGORY_EMOJIS[listing.value?.category] ?? DEFAULT_EMOJI
)

const formattedDate = computed(() => {
  if (!listing.value?.date_posted) return 'Unknown'
  return new Date(listing.value.date_posted).toLocaleDateString(undefined, {
    year: 'numeric', month: 'long', day: 'numeric',
  })
})

const { toast, showToast } = useToast()

// Edit
const modal = ref({ open: false, listing: null, saving: false, apiError: null })

function openEdit() {
  modal.value = { open: true, listing: listing.value, saving: false, apiError: null }
}

async function handleSave(payload) {
  modal.value.saving   = true
  modal.value.apiError = null
  try {
    listing.value    = await updateListing(listing.value.id, payload)
    modal.value.open = false
    showToast('Listing updated successfully.')
  } catch (e) {
    modal.value.apiError = e.message
  } finally {
    modal.value.saving = false
  }
}

// Delete
const deleteOpen    = ref(false)
const deleteLoading = ref(false)

async function handleDelete() {
  deleteLoading.value = true
  try {
    await deleteListing(listing.value.id)
    router.push('/')
  } catch (e) {
    deleteOpen.value = false
    showToast(`Delete failed: ${e.message}`, 'error')
  } finally {
    deleteLoading.value = false
  }
}
</script>

<style scoped>
.page {
  max-width: 1000px;
  margin: 0 auto;
  padding: 28px 24px 60px;
}

/* Back button */
.back-btn {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 6px 0;
  background: none;
  border: none;
  font-size: 14px;
  font-weight: 500;
  color: #6b7280;
  cursor: pointer;
  font-family: inherit;
  margin-bottom: 24px;
  transition: color 0.14s;
}

.back-btn:hover { color: #111827; }

/* Loading skeleton */
.detail-skeleton {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 32px;
}

.skel-image {
  border-radius: 16px;
  aspect-ratio: 4/3;
  background: #e5e7eb;
  animation: pulse 1.4s ease-in-out infinite;
}

.skel-body {
  display: flex;
  flex-direction: column;
  gap: 14px;
  padding-top: 8px;
}

.skel-line {
  height: 14px;
  border-radius: 6px;
  background: #e5e7eb;
  animation: pulse 1.4s ease-in-out infinite;
}
.w-20  { width: 20%; }
.w-40  { width: 40%; }
.w-60  { width: 60%; }
.w-80  { width: 80%; }
.w-full { width: 100%; }

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50%       { opacity: 0.45; }
}

/* Error state */
.error-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 80px 24px;
  gap: 12px;
}

.error-icon {
  width: 56px;
  height: 56px;
  border-radius: 50%;
  background: #fef2f2;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #ef4444;
  margin-bottom: 4px;
}

.error-title {
  font-size: 15px;
  font-weight: 500;
  color: #374151;
}

/* Detail layout */
.detail-layout {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 40px;
  align-items: start;
}

@media (max-width: 700px) {
  .detail-layout { grid-template-columns: 1fr; gap: 24px; }
  .detail-skeleton { grid-template-columns: 1fr; }
}

/* Image */
.detail-image-wrap {
  border-radius: 16px;
  overflow: hidden;
  border: 1px solid var(--color-border);
  aspect-ratio: 4/3;
  background: var(--color-bg);
}

.detail-img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  display: block;
}

.detail-placeholder {
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
}

.placeholder-icon { font-size: 5rem; }

/* Info panel */
.detail-info {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.badge-row {
  display: flex;
  gap: 8px;
  align-items: center;
}

.chip {
  font-size: 11px;
  font-weight: 700;
  padding: 4px 10px;
  border-radius: 999px;
  background: #f3f4f6;
  color: #374151;
  text-transform: uppercase;
  letter-spacing: 0.06em;
}

.detail-title {
  font-size: 24px;
  font-weight: 800;
  color: #111827;
  line-height: 1.3;
  letter-spacing: -0.02em;
}

.detail-price {
  font-size: 28px;
  font-weight: 800;
  color: var(--color-primary-dark);
  letter-spacing: -0.02em;
}

.divider {
  height: 1px;
  background: var(--color-bg);
}

.detail-desc {
  font-size: 15px;
  color: #4b5563;
  line-height: 1.7;
  white-space: pre-line;
}

/* Meta list */
.meta-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.meta-row {
  display: flex;
  align-items: center;
  gap: 8px;
}

.meta-row dt {
  font-size: 13px;
  color: #9ca3af;
  font-weight: 500;
  min-width: 100px;
}

.meta-row dd {
  font-size: 13.5px;
  color: #374151;
  font-weight: 500;
}

/* Actions */
.detail-actions {
  display: flex;
  gap: 10px;
  padding-top: 4px;
}

</style>
