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

---

# Phase 9 — reliability (autostart + backups)

Goal: the whole stack comes back after a reboot with no manual steps, and the
poll data is backed up off the box.

## 1. Everything starts on boot

```bash
# Docker daemon itself (containers use restart: unless-stopped, but Docker must run)
sudo systemctl enable --now docker

# meet compose stack as a unit (recreates the stack even if containers were removed)
sudo cp ~/meet/deploy/systemd/meet.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable --now meet.service

# Cloudflare Tunnel (this is the shared home tunnel that also serves the mensa
# shortener + meet-api ingress) — confirm it's enabled:
systemctl is-enabled cloudflared || sudo systemctl enable --now cloudflared
```

**Tailscale Serve** (admin-web) persists across reboots because `--bg` writes the
serve config, which `tailscaled` restores on start. Verify after a reboot with
`tailscale serve status`; if it's ever empty, re-run `sudo tailscale serve --bg 8081`.

Quick reboot test: `sudo reboot`, then from a tailnet device check
`https://meet.jackdehaan.com/e/<id>` (public) and
`https://<machine>.<tailnet>.ts.net` (admin) both load.

## 2. MongoDB backups

`backup-mongo.sh` streams a gzipped `mongodump` out of the container to the host,
prunes old archives, and can copy the latest offsite via rclone. A systemd timer
runs it daily.

```bash
# One-off (writes to /home/jdh/meet-backups by default):
~/meet/deploy/backup-mongo.sh

# Schedule it daily at 03:30:
sudo cp ~/meet/deploy/systemd/meet-backup.service ~/meet/deploy/systemd/meet-backup.timer /etc/systemd/system/
#   → optionally set MEET_BACKUP_RCLONE_REMOTE / KEEP_DAYS in meet-backup.service first
sudo systemctl daemon-reload
sudo systemctl enable --now meet-backup.timer
systemctl list-timers meet-backup.timer     # confirm next run
```

**Offsite copy (recommended):** install `rclone`, run `rclone config` to add a
remote (e.g. Google Drive as `gdrive`), then uncomment
`Environment=MEET_BACKUP_RCLONE_REMOTE=gdrive:meet-backups` in
`meet-backup.service`. Without this, backups only live on the same disk as the DB.

**Restore** a backup:

```bash
cd ~/meet/deploy
docker compose exec -T mongo mongorestore --archive --gzip --drop \
  < /home/jdh/meet-backups/meet-YYYYMMDD-HHMMSS.archive.gz
```

> If `mongodump` isn't found in the container, install the MongoDB Database Tools
> in the image or run a one-off `mongo/mongo-tools` container joined to the compose
> network — but the official `mongo:7` image ships the tools.
