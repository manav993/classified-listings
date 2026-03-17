# Classified Listings — Frontend

Vue 3 single-page application for the classified listings marketplace.

---

## Tech Stack

| Tool        | Purpose                          |
|-------------|----------------------------------|
| Vue 3       | UI framework (Composition API)   |
| Vite        | Build tool and dev server        |
| Vue Router  | Client-side routing              |

---

## Setup

**Requirements:** Node.js 18+

```bash
cd frontend
npm install
```

### Development

```bash
npm run dev
# App runs at http://localhost:5173
# API is proxied to http://localhost:8080 (see vite.config.js)
```

### Production build

```bash
npm run build
# Output written to dist/
```

### Preview production build locally

```bash
npm run preview
```

### Run tests

```bash
npm test
```

---

## Project Structure

```
src/
├── views/
│   ├── ListingsView.vue       # Main listings page (grid, filters, pagination)
│   └── ListingDetailView.vue  # Single listing page (details, edit, delete)
│
├── components/
│   ├── ListingCard.vue        # Individual listing card with image and actions
│   ├── FilterBar.vue          # Search + category and status filters
│   ├── Pagination.vue         # Page navigation
│   ├── ListingModal.vue       # Create / edit listing form with image upload
│   ├── StatusBadge.vue        # Reusable Active / Inactive status pill
│   ├── ConfirmDialog.vue      # Delete confirmation dialog
│   └── Toast.vue              # Success / error toast notification
│
├── composables/
│   └── useListings.js         # Shared state: fetch, filter, paginate, save, delete
│
├── utils/
│   └── listingDisplay.js      # Formatting helpers: relativeTime, price display
│
├── api/
│   └── listings.js            # Thin fetch wrapper for all API calls including uploadImage()
│
├── constants/
│   └── listing.js             # Domain constants: categories, statuses, field limits
│
├── router/
│   └── index.js               # Route definitions (/ and /listings/:id)
│
├── tests/
│   ├── listingDisplay.test.js      # Unit tests for relativeTime utility
│   ├── listings.api.test.js        # Unit tests for buildQuery helper
│   └── listingModal.validate.test.js # Unit tests for form validation rules
│
├── App.vue                    # Root layout with navbar
├── main.js                    # App entry point
└── style.css                  # Global styles and base button classes
```

---

## Key Design Decisions

### State management via composable
All listing state (fetching, filtering, pagination, CRUD) lives in `useListings.js`. Components import only what they need. This avoids a dedicated store library (Pinia/Vuex) for an app of this size while keeping logic centralised and testable.

### URL-persisted filters
Filter state (search, category, status, page) is reflected in the URL query string so the page is bookmarkable and survives a refresh.

### Filters cleared on create
After creating a new listing, all filters are reset so the new listing is always visible regardless of active filter state.

### Image upload
The create/edit form uses a file picker. On selection the file is sent to `POST /api/upload`, which returns a URL. That URL is stored in `image_url` when the listing is saved. The user never types a URL manually.

### Vue 3 instead of Nuxt
This is a pure client-side SPA with no SEO requirements, so Nuxt's SSR adds complexity without benefit. The composable and component structure maps directly to Nuxt's `pages/` and `composables/` conventions — migration is straightforward if SSR is needed.

---

## Environment Variables

| Variable            | Default | Description                      |
|---------------------|---------|----------------------------------|
| `VITE_API_BASE_URL` | `''`    | Base URL for API calls. Empty string means requests go to the same origin (handled by Nginx proxy in Docker). Set to `http://localhost:8080` for standalone dev. |
