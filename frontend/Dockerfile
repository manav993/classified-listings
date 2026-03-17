# ─── Stage 1: Build ───────────────────────────────────────────────────────────
FROM node:20-alpine AS builder

WORKDIR /app

# Copy package files first so this layer is cached when only source changes.
COPY package.json package-lock.json ./
RUN npm ci

COPY . .

# VITE_API_BASE_URL is empty in Docker; the nginx proxy handles /api routing.
RUN npm run build

# ─── Stage 2: Serve ───────────────────────────────────────────────────────────
FROM nginx:alpine

# Copy the built SPA into nginx's web root.
COPY --from=builder /app/dist /usr/share/nginx/html

# Custom nginx config to support Vue Router history mode and proxy /api to backend.
COPY nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
