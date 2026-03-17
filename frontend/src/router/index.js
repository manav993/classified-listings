import { createRouter, createWebHistory } from 'vue-router'
import ListingsView from '../views/ListingsView.vue'

const routes = [
  {
    path: '/',
    name: 'listings',
    component: ListingsView,
  },
  {
    path: '/listings/:id',
    name: 'listing-detail',
    component: () => import('../views/ListingDetailView.vue'),
  },
  // Catch-all: redirect any unmatched path to the listings page instead of a blank screen.
  {
    path: '/:pathMatch(.*)*',
    redirect: '/',
  },
]

export default createRouter({
  history: createWebHistory(),
  routes,
})
