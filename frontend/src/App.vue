<template>
  <div class="app-shell">
    <header class="site-header">
      <div class="header-inner">

        <div class="brand">
          <div class="brand-icon">
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none"
              stroke="currentColor" stroke-width="2.5"
              stroke-linecap="round" stroke-linejoin="round">
              <path d="M6 2L3 6v14a2 2 0 002 2h14a2 2 0 002-2V6l-3-4z"/>
              <line x1="3" y1="6" x2="21" y2="6"/>
              <path d="M16 10a4 4 0 01-8 0"/>
            </svg>
          </div>
          <span class="brand-name">Marketplace</span>
        </div>


      </div>
    </header>

    <main class="main-content">
      <RouterView />
    </main>
  </div>
</template>

<script setup>
import { RouterView } from 'vue-router'
</script>

<style scoped>
.app-shell {
  display: flex;
  flex-direction: column;
  min-height: 100vh;
}

.site-header {
  position: sticky;
  top: 0;
  z-index: 50;
  background: #fff;
  border-bottom: 1px solid #e5e7eb;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.06);
}

.header-inner {
  max-width: 1140px;
  margin: 0 auto;
  padding: 0 24px;
  height: 56px;
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.brand {
  display: flex;
  align-items: center;
  gap: 10px;
  text-decoration: none;
}

.brand-icon {
  width: 34px;
  height: 34px;
  background: linear-gradient(135deg, #10b981, #059669);
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
  flex-shrink: 0;
}

.brand-name {
  font-size: 16px;
  font-weight: 700;
  color: #111827;
  letter-spacing: -0.01em;
}


.main-content {
  flex: 1;
}
</style>
