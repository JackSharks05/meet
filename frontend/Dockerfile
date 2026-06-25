# Builds the ADMIN frontend (create/summary UI, VUE_APP_ADMIN=true) and serves it
# via Caddy with SPA fallback + a same-origin /api reverse-proxy to the admin Go
# listener. Used for the Tailscale-only admin site. Build context = repo root.

# --- build stage ---
FROM node:20-alpine AS build
WORKDIR /app
COPY frontend/package.json frontend/package-lock.json ./
RUN --mount=type=cache,target=/root/.npm npm install
COPY frontend/ ./
# Admin build: shows the create UI; serverURL stays "/api" (same-origin via Caddy).
RUN VUE_APP_ADMIN=true npm run build

# --- serve stage ---
FROM caddy:2-alpine
COPY --from=build /app/dist /srv
COPY deploy/Caddyfile /etc/caddy/Caddyfile
EXPOSE 80
