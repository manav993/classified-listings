<template>
  <div class="filter-bar">

    <div class="search-wrap">
      <svg class="search-icon" width="15" height="15" viewBox="0 0 24 24" fill="none"
        stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
        <circle cx="11" cy="11" r="8"/><path d="M21 21l-4.35-4.35"/>
      </svg>
      <input
        v-model="localSearch"
        type="text"
        placeholder="Search listings…"
        aria-label="Search listings"
        @input="onSearchInput"
      />
    </div>

    <div class="select-wrap">
      <select v-model="localCategory" @change="submit" aria-label="Filter by category">
        <option value="">All Categories</option>
        <option v-for="c in CATEGORIES" :key="c" :value="c">{{ c }}</option>
      </select>
      <svg class="chevron" width="12" height="12" viewBox="0 0 24 24" fill="none"
        stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
        <polyline points="6 9 12 15 18 9"/>
      </svg>
    </div>

    <div class="select-wrap">
      <select v-model="localStatus" @change="submit" aria-label="Filter by status">
        <option value="">All Statuses</option>
        <option v-for="s in STATUSES" :key="s" :value="s">{{ s }}</option>
      </select>
      <svg class="chevron" width="12" height="12" viewBox="0 0 24 24" fill="none"
        stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
        <polyline points="6 9 12 15 18 9"/>
      </svg>
    </div>

    <button class="reset-btn" @click="reset">
      <svg width="13" height="13" viewBox="0 0 24 24" fill="none"
        stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
        <path d="M3 12a9 9 0 109-9 9.75 9.75 0 00-6.74 2.74L3 8"/>
        <path d="M3 3v5h5"/>
      </svg>
      Reset filters
    </button>

  </div>
</template>

<script setup>
import { ref, watch, onUnmounted } from 'vue'
import { CATEGORIES, STATUSES } from '../constants/listing.js'

const props = defineProps({
  modelValue: {
    type: Object,
    default: () => ({ search: '', category: '', status: '' }),
  },
})

// Standard Vue v-model event name so <FilterBar v-model="..." /> works.
const emit = defineEmits(['update:modelValue'])

const localSearch   = ref(props.modelValue.search   ?? '')
const localCategory = ref(props.modelValue.category ?? '')
const localStatus   = ref(props.modelValue.status   ?? '')

// Deep-watch with guards so we only sync when the parent actually changes a
// value, avoiding redundant updates when the parent re-creates the same object.
watch(() => props.modelValue, (v) => {
  if (v.search   !== localSearch.value)   localSearch.value   = v.search   ?? ''
  if (v.category !== localCategory.value) localCategory.value = v.category ?? ''
  if (v.status   !== localStatus.value)   localStatus.value   = v.status   ?? ''
}, { deep: true })

function submit() {
  emit('update:modelValue', {
    search:   localSearch.value,
    category: localCategory.value,
    status:   localStatus.value,
  })
}

let debounceTimer = null
function onSearchInput() {
  clearTimeout(debounceTimer)
  debounceTimer = setTimeout(submit, 380)
}
onUnmounted(() => clearTimeout(debounceTimer))

function reset() {
  localSearch.value   = ''
  localCategory.value = ''
  localStatus.value   = ''
  submit()
}
</script>

<style scoped>
.filter-bar {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 10px;
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-card);
  padding: 12px 16px;
  margin-bottom: 24px;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.05);
}

/* Search */
.search-wrap {
  flex: 1;
  min-width: 180px;
  position: relative;
  display: flex;
  align-items: center;
}

.search-icon {
  position: absolute;
  left: 11px;
  color: #9ca3af;
  pointer-events: none;
  flex-shrink: 0;
}

.search-wrap input {
  width: 100%;
  padding: 8px 12px 8px 34px;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-pill);
  font-size: 14px;
  font-family: inherit;
  background: var(--color-surface-alt);
  color: var(--color-text);
  transition: border-color 0.15s, box-shadow 0.15s;
}

.search-wrap input::placeholder { color: #9ca3af; }

.search-wrap input:focus {
  outline: none;
  border-color: var(--color-primary);
  background: var(--color-surface);
  box-shadow: 0 0 0 2px rgba(16, 185, 129, 0.15);
}

/* Selects */
.select-wrap {
  position: relative;
  display: flex;
  align-items: center;
}

.select-wrap select {
  appearance: none;
  padding: 8px 30px 8px 14px;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-pill);
  font-size: 14px;
  font-family: inherit;
  background: var(--color-surface-alt);
  color: var(--color-text-secondary);
  cursor: pointer;
  transition: border-color 0.15s, box-shadow 0.15s;
  white-space: nowrap;
}

.select-wrap select:focus {
  outline: none;
  border-color: var(--color-primary);
  box-shadow: 0 0 0 2px rgba(16, 185, 129, 0.15);
}

.chevron {
  position: absolute;
  right: 11px;
  color: #6b7280;
  pointer-events: none;
}

/* Reset */
.reset-btn {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  padding: 8px 14px;
  border: none;
  border-radius: 999px;
  background: transparent;
  font-size: 13.5px;
  font-weight: 500;
  color: #6b7280;
  cursor: pointer;
  font-family: inherit;
  transition: color 0.14s, background 0.14s;
  white-space: nowrap;
}

.reset-btn:hover {
  color: var(--color-text);
  background: var(--color-bg);
}
</style>
