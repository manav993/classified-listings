<template>
  <Teleport to="body">
    <div class="backdrop" @click.self="$emit('close')">
      <div class="modal" role="dialog" aria-modal="true" :aria-label="isEdit ? 'Edit listing' : 'New listing'">
        <div class="modal-header">
          <h2>{{ isEdit ? 'Edit Listing' : 'New Listing' }}</h2>
          <button class="close-btn" @click="$emit('close')" aria-label="Close">×</button>
        </div>

        <form @submit.prevent="submit" novalidate>
          <div class="modal-body">
            <!-- API-level errors -->
            <div v-if="apiError" class="alert alert-error">{{ apiError }}</div>

            <div class="field" :class="{ 'has-error': errors.title }">
              <label for="m-title">Title <span class="req">*</span></label>
              <input id="m-title" v-model="form.title" type="text" :maxlength="TITLE_MAX_LENGTH" placeholder="e.g. Mountain bike" />
              <div class="field-footer">
                <span v-if="errors.title" class="field-error">{{ errors.title }}</span>
                <span class="char-count" :class="{ 'near-limit': form.title.length >= TITLE_MAX_LENGTH * 0.9 }">
                  {{ form.title.length }}/{{ TITLE_MAX_LENGTH }}
                </span>
              </div>
            </div>

            <div class="field" :class="{ 'has-error': errors.description }">
              <label for="m-desc">Description <span class="req">*</span></label>
              <textarea id="m-desc" v-model="form.description" rows="3" :maxlength="DESCRIPTION_MAX_LENGTH" placeholder="Describe the listing…" />
              <div class="field-footer">
                <span v-if="errors.description" class="field-error">{{ errors.description }}</span>
                <span class="char-count" :class="{ 'near-limit': form.description.length >= DESCRIPTION_MAX_LENGTH * 0.9 }">
                  {{ form.description.length }}/{{ DESCRIPTION_MAX_LENGTH }}
                </span>
              </div>
            </div>

            <div class="row-2">
              <div class="field" :class="{ 'has-error': errors.price }">
                <label for="m-price">Price (£) <span class="req">*</span></label>
                <input id="m-price" v-model.number="form.price" type="number" min="0.01" :max="PRICE_MAX" step="0.01" placeholder="e.g. 99.99" />
                <span v-if="errors.price" class="field-error">{{ errors.price }}</span>
              </div>

              <div class="field" :class="{ 'has-error': errors.category }">
                <label for="m-cat">Category <span class="req">*</span></label>
                <select id="m-cat" v-model="form.category">
                  <option value="" disabled>Select…</option>
                  <option v-for="c in CATEGORIES" :key="c" :value="c">{{ c }}</option>
                </select>
                <span v-if="errors.category" class="field-error">{{ errors.category }}</span>
              </div>
            </div>

            <div class="field" :class="{ 'has-error': errors.status }">
              <label>Status <span class="req">*</span></label>
              <div class="radio-group">
                <label v-for="s in STATUSES" :key="s" class="radio-label">
                  <input type="radio" v-model="form.status" :value="s" />
                  {{ s }}
                </label>
              </div>
              <span v-if="errors.status" class="field-error">{{ errors.status }}</span>
            </div>

            <div class="field">
              <label>
                Image
                <span class="optional">(optional)</span>
              </label>
              <div class="image-upload-wrap">
                <label class="upload-btn" :class="{ uploading }">
                  <input
                    type="file"
                    accept="image/*"
                    @change="onFileChange"
                    :disabled="uploading"
                  />
                  <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor"
                    stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                    <path d="M21 15v4a2 2 0 01-2 2H5a2 2 0 01-2-2v-4"/>
                    <polyline points="17 8 12 3 7 8"/><line x1="12" y1="3" x2="12" y2="15"/>
                  </svg>
                  <span v-if="uploading">Uploading…</span>
                  <span v-else-if="uploadedUrl">Change image</span>
                  <span v-else>Choose image</span>
                </label>
                <div v-if="uploadedUrl" class="image-preview">
                  <img :src="uploadedUrl" alt="Preview" />
                  <button type="button" class="remove-img" @click="uploadedUrl = ''" aria-label="Remove image">×</button>
                </div>
                <span v-if="uploadError" class="field-error">{{ uploadError }}</span>
              </div>
            </div>
          </div>

          <div class="modal-footer">
            <button type="button" class="btn btn-outline" @click="$emit('close')">Cancel</button>
            <button type="submit" class="btn btn-primary" :disabled="saving">
              <span v-if="saving">Saving…</span>
              <span v-else>{{ isEdit ? 'Save changes' : 'Create listing' }}</span>
            </button>
          </div>
        </form>
      </div>
    </div>
  </Teleport>
</template>

<script setup>
import { ref, computed, watch, onMounted, onUnmounted } from 'vue'
import { CATEGORIES, STATUSES, TITLE_MIN_LENGTH, TITLE_MAX_LENGTH, DESCRIPTION_MIN_LENGTH, DESCRIPTION_MAX_LENGTH, PRICE_MAX } from '../constants/listing.js'
import { uploadImage } from '../api/listings.js'

const props = defineProps({
  listing: { type: Object, default: null }, // null = create mode
  saving:  { type: Boolean, default: false },
  apiError:{ type: String, default: null },
})

const emit = defineEmits(['close', 'save'])

const isEdit = computed(() => !!props.listing)

const EMPTY = { title: '', description: '', price: null, category: '', status: 'Active' }

const form        = ref({ ...EMPTY })
const errors      = ref({})
const uploadedUrl = ref('')   // URL returned by /api/upload (or pre-filled from existing listing)
const uploading   = ref(false)
const uploadError = ref('')

function onKeydown(e) {
  if (e.key === 'Escape') emit('close')
}
onMounted(() => document.addEventListener('keydown', onKeydown))
onUnmounted(() => document.removeEventListener('keydown', onKeydown))

watch(() => props.listing, (l) => {
  form.value    = l
    ? { title: l.title, description: l.description, price: l.price, category: l.category, status: l.status }
    : { ...EMPTY }
  uploadedUrl.value = l?.image_url ?? ''
  uploadError.value = ''
  errors.value  = {}
}, { immediate: true })

async function onFileChange(e) {
  const file = e.target.files[0]
  if (!file) return
  uploading.value   = true
  uploadError.value = ''
  try {
    uploadedUrl.value = await uploadImage(file)
  } catch (err) {
    uploadError.value = err.message
  } finally {
    uploading.value = false
  }
}

// Client-side validation
function validate() {
  const e = {}
  const trimmedTitle = form.value.title.trim()
  if (!trimmedTitle)                                              e.title = 'Title is required'
  else if (trimmedTitle.length < TITLE_MIN_LENGTH)               e.title = `Title must be at least ${TITLE_MIN_LENGTH} characters`
  else if (form.value.title.length > TITLE_MAX_LENGTH)           e.title = `Title must be ${TITLE_MAX_LENGTH} characters or fewer`

  const trimmedDesc = form.value.description.trim()
  if (!trimmedDesc)                                              e.description = 'Description is required'
  else if (trimmedDesc.length < DESCRIPTION_MIN_LENGTH)         e.description = `Description must be at least ${DESCRIPTION_MIN_LENGTH} characters`
  else if (form.value.description.length > DESCRIPTION_MAX_LENGTH) e.description = `Description must be ${DESCRIPTION_MAX_LENGTH} characters or fewer`

  if (!form.value.price || Number(form.value.price) <= 0) e.price = 'Price must be greater than zero'
  else if (Number(form.value.price) > PRICE_MAX)           e.price = `Price must not exceed £${PRICE_MAX.toLocaleString()}`
  if (!form.value.category)           e.category    = 'Please select a category'
  if (!form.value.status)             e.status      = 'Please select a status'
  errors.value = e
  return Object.keys(e).length === 0
}

function submit() {
  if (!validate()) return
  emit('save', {
    title:       form.value.title,
    description: form.value.description,
    price:       Number(form.value.price),
    category:    form.value.category,
    status:      form.value.status,
    image_url:   uploadedUrl.value,
  })
}
</script>

<style scoped>
.backdrop {
  position: fixed;
  inset: 0;
  background: rgba(15, 23, 42, 0.55);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 100;
  padding: 1rem;
}

.modal {
  background: #fff;
  border-radius: 16px;
  width: 100%;
  max-width: 520px;
  max-height: 90vh;
  overflow-y: auto;
  box-shadow: 0 20px 60px rgba(0,0,0,0.2);
  display: flex;
  flex-direction: column;
}

.modal-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 1.25rem 1.5rem;
  border-bottom: 1px solid #f1f5f9;
}

.modal-header h2 {
  margin: 0;
  font-size: 1.15rem;
  font-weight: 700;
  color: #1e293b;
}

.close-btn {
  background: none;
  border: none;
  font-size: 1.6rem;
  color: #94a3b8;
  cursor: pointer;
  line-height: 1;
  padding: 0;
}
.close-btn:hover { color: #475569; }

.modal-body {
  padding: 1.5rem;
  display: flex;
  flex-direction: column;
  gap: 1.1rem;
}

.modal-footer {
  display: flex;
  justify-content: flex-end;
  gap: 0.75rem;
  padding: 1rem 1.5rem;
  border-top: 1px solid #f1f5f9;
}

.row-2 {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 1rem;
}

.field {
  display: flex;
  flex-direction: column;
  gap: 0.35rem;
}

.field label {
  font-size: 0.8rem;
  font-weight: 600;
  color: #475569;
}

.req { color: #ef4444; }

input[type="text"],
input[type="number"],
input[type="url"],
select,
textarea {
  padding: 0.55rem 0.75rem;
  border: 1px solid #cbd5e1;
  border-radius: 8px;
  font-size: 0.9rem;
  font-family: inherit;
  background: #f8fafc;
  transition: border-color 0.15s;
  width: 100%;
  box-sizing: border-box;
}

input:focus, select:focus, textarea:focus {
  outline: none;
  border-color: #6366f1;
  background: #fff;
}

textarea { resize: vertical; }

.has-error input,
.has-error select,
.has-error textarea {
  border-color: #ef4444;
}

.field-error {
  font-size: 0.75rem;
  color: #ef4444;
  font-weight: 500;
}

.radio-group {
  display: flex;
  gap: 1.5rem;
}

.radio-label {
  display: flex;
  align-items: center;
  gap: 0.4rem;
  font-size: 0.9rem;
  font-weight: 500;
  color: #374151;
  cursor: pointer;
}

.optional {
  font-weight: 400;
  color: #94a3b8;
  font-size: 0.75rem;
  margin-left: 4px;
}

.image-upload-wrap {
  display: flex;
  flex-direction: column;
  gap: 0.6rem;
}

.upload-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.45rem;
  padding: 0.6rem 1rem;
  border: 1.5px dashed #cbd5e1;
  border-radius: 8px;
  background: #f8fafc;
  color: #475569;
  font-size: 0.875rem;
  font-weight: 500;
  cursor: pointer;
  transition: border-color 0.15s, background 0.15s, color 0.15s;
  width: 100%;
  box-sizing: border-box;
}

.upload-btn:hover {
  border-color: #6366f1;
  background: #eef2ff;
  color: #4f46e5;
}

.upload-btn.uploading {
  opacity: 0.6;
  cursor: not-allowed;
}

.upload-btn input[type="file"] {
  display: none;
}

.image-preview {
  position: relative;
  border-radius: 8px;
  overflow: hidden;
  border: 1px solid #e2e8f0;
  max-height: 160px;
}

.image-preview img {
  width: 100%;
  max-height: 160px;
  object-fit: cover;
  display: block;
}

.remove-img {
  position: absolute;
  top: 6px;
  right: 6px;
  width: 22px;
  height: 22px;
  border-radius: 50%;
  background: rgba(0, 0, 0, 0.5);
  color: #fff;
  border: none;
  font-size: 1rem;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  line-height: 1;
  padding: 0;
}

.alert {
  padding: 0.75rem 1rem;
  border-radius: 8px;
  font-size: 0.875rem;
  font-weight: 500;
}

.alert-error {
  background: #fef2f2;
  color: #dc2626;
  border: 1px solid #fecaca;
}

.field-footer {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  min-height: 1.25rem;
}

.char-count {
  font-size: 0.75rem;
  color: #94a3b8;
  margin-left: auto;
  white-space: nowrap;
}

.char-count.near-limit {
  color: #f59e0b;
  font-weight: 500;
}
</style>
