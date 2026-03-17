import { ref, onUnmounted } from 'vue'

/**
 * Provides a self-dismissing toast notification.
 * Calling showToast() twice in quick succession cancels the first timer so
 * only one toast is ever visible.
 *
 * @param {number} duration - auto-dismiss delay in ms (default 3500)
 */
export function useToast(duration = 3500) {
  const toast = ref({ visible: false, message: '', type: 'success' })
  let timer = null

  function showToast(message, type = 'success') {
    clearTimeout(timer)
    toast.value = { visible: true, message, type }
    timer = setTimeout(() => { toast.value.visible = false }, duration)
  }

  onUnmounted(() => clearTimeout(timer))

  return { toast, showToast }
}
