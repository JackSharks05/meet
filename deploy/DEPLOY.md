# Deploying meet — Phase 5 (public site)

The public site is a **static respond-only frontend on Vercel** that calls the
**home server's poll/respond API through a Cloudflare Tunnel**. The admin side
(create/summary) is Phase 6 (Tailscale) and is intentionally not exposed here.

```
Respondents ─► meet.jackdehaan.com (Vercel static)  ──calls──►  meet-api.jackdehaan.com
                                                                  └ Cloudflare Tunnel ─► localhost:3002 (MODE=public Go listener)
```

Prereqs: the `meet` repo pushed to GitHub, DNS for `jackdehaan.com` on Cloudflare,
a Vercel account, and the Docker stack from this folder running on the Arch server.

---

## 1. Backend up on the Arch server

```bash
cd deploy
cp .env.template .env      # fill in SESSION_SECRET (openssl rand -base64 48), etc.
# IMPORTANT for prod: set PUBLIC_CORS_ORIGINS=https://meet.jackdehaan.com
docker compose up -d --build
# server-public is now on 127.0.0.1:3002 (not exposed to the internet directly)
```

## 2. Cloudflare Tunnel → public API

Install `cloudflared` on the server, then:

```bash
cloudflared tunnel login
cloudflared tunnel create meet
cloudflared tunnel route dns meet meet-api.jackdehaan.com
# copy the printed <TUNNEL_ID>.json into deploy/cloudflared/ and set <TUNNEL_ID> in config.yml
cloudflared tunnel --config deploy/cloudflared/config.yml run     # test it
```

`config.yml` routes `meet-api.jackdehaan.com → http://localhost:3002` only.
Make it permanent with the systemd unit in Phase 9.

Verify: `curl https://meet-api.jackdehaan.com/api/auth/status` → `{"error":"not-signed-in"}`.

## 3. Frontend on Vercel

- Import `JackSharks05/meet` into Vercel; set **Root Directory = `frontend`**
  (Vercel reads `frontend/vercel.json` for the SPA rewrites).
- **Environment variables:**
  - `VUE_APP_API_URL = https://meet-api.jackdehaan.com/api`
  - leave **`VUE_APP_ADMIN` unset** → public, respond-only build (no create UI).
- Deploy, then add the custom domain **`meet.jackdehaan.com`**.

## 4. Verify end to end

- `https://meet.jackdehaan.com/` → dark booking landing.
- `https://meet.jackdehaan.com/e/<id>` → poll loads (data via the tunnel; CORS allows the Vercel origin).
- Submitting availability works; there is **no create-event UI** (operator-only).

## Notes
- Keep `PUBLIC_CORS_ORIGINS` = the exact Vercel origin (`https://meet.jackdehaan.com`).
- The home server must be up for live poll data; Vercel still serves the page if it blips.
- Phase 6 wires the **admin** build over Tailscale Serve (below).

---

# Phase 6 — admin site (Tailscale, operator-only)

The `admin-web` container (Caddy) serves the **admin build** (create/summary UI)
and proxies `/api` **same-origin** to the `server-admin` listener — so sessions
work with no CORS. It listens on `127.0.0.1:8081`; you expose it to your own
devices with Tailscale Serve. It is never tunneled to the public internet.

```
You (any device on the tailnet) ─► tailscale serve ─► 127.0.0.1:8081 (Caddy)
                                                         ├ /        → admin SPA (create UI)
                                                         └ /api/*   → server-admin (Go)
```

### Steps (on the Arch server)

```bash
cd deploy
# 1) Lock sign-in to your account (operator allowlist) in deploy/.env:
#      ADMIN_EMAILS=you@gmail.com
docker compose up -d --build        # brings up admin-web on 127.0.0.1:8081

# 2) Expose it to your tailnet (HTTPS, tailnet-only):
tailscale serve --bg 8081
#   → reachable at https://<machine>.<tailnet>.ts.net  (syntax varies by version;
#     older: `tailscale serve https / http://127.0.0.1:8081`)
```

Verify (from a tailnet device): the page loads the **create UI**, and
`/api/auth/status` returns `{"error":"not-signed-in"}`.

### Important
- **Creating events needs Google sign-in** → that's **Phase 7 (OAuth)**. Until
  then the admin UI loads and you can navigate, but sign-in / create won't work
  (the API enforces `AuthRequired` + the `ADMIN_EMAILS` allowlist).
- Optionally add a link to `https://<machine>.<tailnet>.ts.net` from the **Mensa**
  dashboard; longer-term the create UI moves into Mensa.
- The public tunnel (Phase 5) only exposes `server-public`; `admin-web` /
  `server-admin` are never in the Cloudflare ingress.
