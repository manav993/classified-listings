<template>
  <div class="card" :class="{ inactive: listing.status === 'Inactive' }">

    <!-- RouterLink wraps the visual area only — action buttons sit outside
         so Edit/Delete don't trigger navigation when clicked -->
    <RouterLink :to="`/listings/${listing.id}`" class="card-link">

      <div class="card-image">
        <img
          v-if="listing.image_url"
          :src="listing.image_url"
          :alt="listing.title"
          loading="lazy"
        />
        <div v-else class="placeholder" :style="{ background: placeholderGradient }">
          <span class="placeholder-icon">{{ categoryEmoji }}</span>
        </div>

        <div class="badge-row">
          <StatusBadge :status="listing.status" sm />
          <span class="chip">{{ listing.category }}</span>
        </div>
      </div>

      <div class="card-body">
        <h3 class="title">{{ listing.title }}</h3>
        <p class="desc">{{ listing.description }}</p>

        <div class="meta">
          <span class="price">£{{ listing.price.toLocaleString() }}</span>
          <span class="posted">Posted {{ relativeTime }}</span>
        </div>
      </div>

    </RouterLink>

    <div class="card-actions">
      <button class="action-edit" @click="$emit('edit', listing)">
        <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor"
          stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <path d="M11 4H4a2 2 0 00-2 2v14a2 2 0 002 2h14a2 2 0 002-2v-7"/>
          <path d="M18.5 2.5a2.121 2.121 0 013 3L12 15l-4 1 1-4 9.5-9.5z"/>
        </svg>
        Edit
      </button>
      <button class="action-delete" @click="$emit('delete', listing)" aria-label="Delete listing">
        <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor"
          stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <polyline points="3 6 5 6 21 6"/>
          <path d="M19 6v14a2 2 0 01-2 2H7a2 2 0 01-2-2V6m3 0V4a1 1 0 011-1h4a1 1 0 011 1v2"/>
        </svg>
      </button>
    </div>

  </div>
</template>

<script setup>
import { computed } from 'vue'
import { RouterLink } from 'vue-router'
import { CATEGORY_GRADIENTS, CATEGORY_EMOJIS, DEFAULT_GRADIENT, DEFAULT_EMOJI, relativeTime as formatRelativeTime } from '../utils/listingDisplay.js'
import StatusBadge from './StatusBadge.vue'

const props = defineProps({
  listing: { type: Object, required: true },
})
defineEmits(['edit', 'delete'])

const placeholderGradient = computed(
  () => CATEGORY_GRADIENTS[props.listing.category] ?? DEFAULT_GRADIENT
)
const categoryEmoji = computed(
  () => CATEGORY_EMOJIS[props.listing.category] ?? DEFAULT_EMOJI
)
const relativeTime = computed(() => formatRelativeTime(props.listing.date_posted))
</script>

<style scoped>
.card {
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-card);
  overflow: hidden;
  display: flex;
  flex-direction: column;
  transition: box-shadow 0.18s, transform 0.18s;
}

.card:hover {
  box-shadow: 0 8px 28px rgba(0, 0, 0, 0.09);
  transform: translateY(-2px);
}

.card.inactive {
  opacity: 0.7;
}

/* Clickable link wrapper */
.card-link {
  display: flex;
  flex-direction: column;
  flex: 1;
  text-decoration: none;
  color: inherit;
  min-width: 0;
}

/* Image area */
.card-image {
  position: relative;
  aspect-ratio: 16 / 9;
  overflow: hidden;
  background: #f3f4f6;
}

.card-image img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  display: block;
}

.placeholder {
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
}

.placeholder-icon {
  font-size: 2.8rem;
  filter: drop-shadow(0 2px 4px rgba(0,0,0,0.1));
}

/* Badges on top of image */
.badge-row {
  position: absolute;
  top: 10px;
  left: 10px;
  display: flex;
  gap: 6px;
  align-items: center;
}

.chip {
  font-size: 10px;
  font-weight: 700;
  padding: 3px 8px;
  border-radius: 999px;
  background: rgba(255, 255, 255, 0.92);
  color: #374151;
  text-transform: uppercase;
  letter-spacing: 0.07em;
  backdrop-filter: blur(4px);
}

/* Body */
.card-body {
  padding: 14px 16px 10px;
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.title {
  font-size: 15px;
  font-weight: 700;
  color: #111827;
  line-height: 1.35;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.desc {
  font-size: 13px;
  color: #6b7280;
  line-height: 1.5;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
  flex: 1;
}

.meta {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  gap: 8px;
  margin-top: 6px;
}

.price {
  font-size: 17px;
  font-weight: 800;
  color: var(--color-primary-dark);
  letter-spacing: -0.01em;
}

.posted {
  font-size: 12px;
  color: #9ca3af;
  white-space: nowrap;
}

/* Actions */
.card-actions {
  display: flex;
  gap: 8px;
  padding: 10px 16px 14px;
  border-top: 1px solid var(--color-bg);
}

.action-edit {
  flex: 1;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  padding: 8px 14px;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-input);
  background: var(--color-surface-alt);
  color: var(--color-text-secondary);
  font-size: 13px;
  font-weight: 500;
  font-family: inherit;
  cursor: pointer;
  transition: background 0.14s, border-color 0.14s;
}

.action-edit:hover {
  background: var(--color-bg);
  border-color: var(--color-border-dark);
}

.action-delete {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 36px;
  height: 36px;
  border: 1px solid #fecaca;
  border-radius: 8px;
  background: #fef2f2;
  color: #ef4444;
  font-family: inherit;
  cursor: pointer;
  transition: background 0.14s, border-color 0.14s;
  flex-shrink: 0;
}

.action-delete:hover {
  background: #fee2e2;
  border-color: #fca5a5;
}
</style>
