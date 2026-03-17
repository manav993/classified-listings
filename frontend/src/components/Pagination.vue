<template>
  <nav class="pagination" aria-label="Pagination">

    <button class="page-btn nav-btn" :disabled="current <= 1" @click="$emit('change', current - 1)">
      <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor"
        stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
        <polyline points="15 18 9 12 15 6"/>
      </svg>
    </button>

    <template v-for="p in pages" :key="p">
      <span v-if="p === '...'" class="ellipsis">…</span>
      <button
        v-else
        class="page-btn"
        :class="{ active: p === current }"
        @click="$emit('change', p)"
      >
        {{ p }}
      </button>
    </template>

    <button class="page-btn nav-btn" :disabled="current >= totalPages" @click="$emit('change', current + 1)">
      <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor"
        stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
        <polyline points="9 18 15 12 9 6"/>
      </svg>
    </button>

  </nav>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({
  current:    { type: Number, required: true },
  totalPages: { type: Number, required: true },
})

defineEmits(['change'])

const pages = computed(() => {
  const { current, totalPages } = props
  if (totalPages <= 7) return range(1, totalPages)
  if (current <= 4)   return [...range(1, 5), '...', totalPages]
  if (current >= totalPages - 3) return [1, '...', ...range(totalPages - 4, totalPages)]
  return [1, '...', ...range(current - 1, current + 1), '...', totalPages]
})

function range(from, to) {
  return Array.from({ length: to - from + 1 }, (_, i) => from + i)
}
</script>

<style scoped>
.pagination {
  display: flex;
  align-items: center;
  gap: 4px;
}

.page-btn {
  min-width: 34px;
  height: 34px;
  padding: 0 10px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-input);
  background: var(--color-surface);
  font-size: 13.5px;
  font-family: inherit;
  font-weight: 500;
  color: var(--color-text-secondary);
  cursor: pointer;
  transition: background 0.13s, border-color 0.13s, color 0.13s;
}

.page-btn:hover:not(:disabled) {
  background: var(--color-bg);
  border-color: var(--color-border-dark);
}

.page-btn.active {
  background: var(--color-text);
  border-color: var(--color-text);
  color: #fff;
}

.page-btn:disabled {
  opacity: 0.35;
  cursor: not-allowed;
}

.nav-btn {
  color: #6b7280;
}

.ellipsis {
  min-width: 28px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  font-size: 14px;
  color: #9ca3af;
}
</style>
