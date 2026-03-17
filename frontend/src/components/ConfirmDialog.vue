<template>
  <Teleport to="body">
    <div class="backdrop" @click.self="$emit('cancel')">
      <div
        class="dialog"
        role="alertdialog"
        aria-modal="true"
        aria-labelledby="confirm-title"
        aria-describedby="confirm-desc"
      >
        <div class="icon">{{ icon }}</div>
        <h3 id="confirm-title">{{ title }}</h3>
        <p id="confirm-desc">
          <strong>{{ listing?.title }}</strong> {{ message }}
        </p>
        <div class="actions">
          <button class="btn btn-outline" @click="$emit('cancel')">Cancel</button>
          <button class="btn btn-danger" :disabled="loading" @click="$emit('confirm')">
            {{ loading ? 'Working…' : confirmLabel }}
          </button>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<script setup>
import { onMounted, onUnmounted } from 'vue'

const props = defineProps({
  listing:      { type: Object,  default: null },
  loading:      { type: Boolean, default: false },
  title:        { type: String,  default: 'Delete listing?' },
  message:      { type: String,  default: 'will be permanently removed. This cannot be undone.' },
  confirmLabel: { type: String,  default: 'Yes, delete' },
  icon:         { type: String,  default: '🗑️' },
})
const emit = defineEmits(['confirm', 'cancel'])

function onKeydown(e) {
  if (e.key === 'Escape') emit('cancel')
}
onMounted(() => document.addEventListener('keydown', onKeydown))
onUnmounted(() => document.removeEventListener('keydown', onKeydown))
</script>

<style scoped>
.backdrop {
  position: fixed;
  inset: 0;
  background: rgba(15, 23, 42, 0.55);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 200;
  padding: 1rem;
}

.dialog {
  background: #fff;
  border-radius: 16px;
  padding: 2rem;
  width: 100%;
  max-width: 380px;
  text-align: center;
  box-shadow: 0 20px 60px rgba(0,0,0,0.2);
}

.icon { font-size: 2.5rem; margin-bottom: 0.5rem; }

h3 {
  margin: 0 0 0.5rem;
  font-size: 1.15rem;
  font-weight: 700;
  color: #1e293b;
}

p {
  margin: 0 0 1.5rem;
  font-size: 0.9rem;
  color: #64748b;
  line-height: 1.55;
}

.actions { display: flex; gap: 0.75rem; justify-content: center; }
</style>
