# meet

A private, self-hosted group-availability scheduler — find the best time for a
group to meet. **meet** is an AGPL-licensed fork of
[Timeful](https://github.com/schej-it/timeful.app), re-themed to a dark style
and adapted for single-operator use behind `meet.jackdehaan.com`.

## Source & license (AGPL-3.0)

**meet** is free software under the GNU Affero General Public License v3.0. It is
a derivative work of Timeful by schej-it. See [`LICENSE`](./LICENSE) and
[`NOTICE`](./NOTICE) for the full terms and attribution.

> Per **AGPL-3.0 §13**, the complete corresponding source of the running
> deployment is published here: **https://github.com/JackSharks05/meet**
> This link is also surfaced in the footer of the running site.

## What's different from Timeful

- **Public site is response-only.** Anyone with a poll link can mark their
  availability; nobody but the operator can create or manage events.
- **Admin is private.** Event creation, editing, summaries, and Google Calendar
  sync live on a build served only over the operator's **Tailscale** network.
- **Self-hosted data.** Backend is **Go + MongoDB in Docker** on a home server;
  the public poll/respond API is exposed through **Cloudflare Tunnel**.
- Removed billing, bots, friends, groups, contacts, analytics, and email
  reminders (see `NOTICE`).

## Architecture (target)

```
Respondents ─► meet.jackdehaan.com (Vercel, static respond-only Vue build)
                       │ calls
                       ▼
              api.meet.jackdehaan.com ─ Cloudflare Tunnel ─► Go "public" listener ┐
                                                                                   ├─► MongoDB
Operator ─► Tailscale Serve ─► admin Vue build ─► Go "admin" listener (tailnet) ──┘   (Docker)
                                                   + Google Calendar OAuth (operator-only)
```

## Repository layout

| Path        | Description                                                        |
|-------------|--------------------------------------------------------------------|
| `frontend/` | Forked Timeful Vue 2 SPA. Builds in `public` or `admin` mode.       |
| `server/`   | Forked Timeful Go/Gin API, backed by MongoDB.                      |
| `deploy/`   | Docker Compose, Cloudflare Tunnel, Tailscale, and backup configs.  |
| `PLAN.md`   | Full architecture decisions and the phased implementation plan.    |
| `LICENSE`   | GNU AGPL-3.0 (verbatim).                                            |
| `NOTICE`    | Attribution to Timeful + summary of modifications.                 |

## Status

🚧 **Work in progress.** Phase 1 (fork + license + scaffold) is complete. See
[`PLAN.md`](./PLAN.md) for the remaining phases.

## Development

Local development requires Docker (for the Go server + MongoDB) and Node 18+
(for the Vue frontend). Detailed setup lands with `deploy/` in a later phase;
for now:

```bash
# Frontend (Vue dev server)
cd frontend && npm install && npm run serve

# Backend + DB (once deploy/ is wired up)
cd deploy && docker compose up
```

Secrets are provided via environment variables and never committed; see
`server/.env.template` and `frontend/.env.template`.
