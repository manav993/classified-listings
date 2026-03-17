<template>
  <span class="badge" :class="status === 'Active' ? 'badge-active' : 'badge-inactive'" :style="sm ? { fontSize: '10px', padding: '3px 8px' } : {}">
    {{ status }}
  </span>
</template>

<script setup>
defineProps({
  status: { type: String, required: true },
  sm:     { type: Boolean, default: false },
})
</script>

<style scoped>
.badge {
  display: inline-block;
  font-size: 11px;
  font-weight: 700;
  padding: 4px 10px;
  border-radius: var(--radius-pill);
  text-transform: uppercase;
  letter-spacing: 0.06em;
}

.badge-active   { background: var(--color-badge-active);   color: #fff; }
.badge-inactive { background: var(--color-badge-inactive); color: #fff; }
</style>
