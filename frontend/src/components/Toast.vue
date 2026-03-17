<template>
  <Transition name="toast">
    <div v-if="toast.visible" :class="['toast', `toast--${toast.type}`]" role="status">
      <svg v-if="toast.type === 'success'" width="16" height="16" viewBox="0 0 24 24"
        fill="none" stroke="currentColor" stroke-width="2.5"
        stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
        <polyline points="20 6 9 17 4 12"/>
      </svg>
      <svg v-else width="16" height="16" viewBox="0 0 24 24"
        fill="none" stroke="currentColor" stroke-width="2.5"
        stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
        <circle cx="12" cy="12" r="10"/>
        <line x1="12" y1="8" x2="12" y2="12"/>
        <line x1="12" y1="16" x2="12.01" y2="16"/>
      </svg>
      {{ toast.message }}
    </div>
  </Transition>
</template>

<script setup>
defineProps({
  toast: { type: Object, required: true },
})
</script>

<style scoped>
.toast {
  position: fixed;
  bottom: 28px;
  left: 50%;
  transform: translateX(-50%);
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px 20px;
  border-radius: var(--radius-pill);
  font-size: 14px;
  font-weight: 500;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.15);
  z-index: 9999;
  white-space: nowrap;
  pointer-events: none;
}

.toast--success {
  background: #052e16;
  color: #d1fae5;
}

.toast--error {
  background: #450a0a;
  color: #fecaca;
}

.toast-enter-active,
.toast-leave-active {
  transition: opacity 0.25s ease, transform 0.25s ease;
}

.toast-enter-from,
.toast-leave-to {
  opacity: 0;
  transform: translateX(-50%) translateY(12px);
}
</style>
