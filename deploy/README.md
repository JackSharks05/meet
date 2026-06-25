# deploy/

Deployment artifacts for the home-server side of **meet** (the Go API + MongoDB).
The public frontend deploys separately to Vercel from `frontend/`.

> Status: 🚧 scaffold only. Concrete files land in Phase 3 (Docker/listeners) and
> Phase 5 (Cloudflare Tunnel) — see [`../PLAN.md`](../PLAN.md).

## Planned contents

| File / dir            | Purpose                                                                 |
|-----------------------|-------------------------------------------------------------------------|
| `docker-compose.yml`  | Go server (built from `../server`) + MongoDB, on a private Docker net.   |
| `.env.template`       | Documents required env vars (Mongo URI, session secret, Google OAuth…). |
| `cloudflared/`        | Cloudflare Tunnel config exposing **only** the public poll/respond API. |
| `tailscale/`          | Tailscale Serve config for the operator-only admin build + admin API.   |
| `systemd/`            | Units to autostart the compose stack, tunnel, and serve on boot.        |
| `backup/`             | `mongodump` cron script + offsite copy.                                  |

## Topology recap

- **Public listener** (`MODE=public`): read-event + submit-response only → Cloudflare
  Tunnel → `api.meet.jackdehaan.com`. Cookieless/guest, CORS-allowed for the Vercel origin.
- **Admin listener** (`MODE=admin`): full API (create/edit/delete/summary + Google
  OAuth), bound to the tailnet interface, **never** routed through the tunnel.
- **MongoDB**: single container, shared by both listeners, not exposed off-host.
