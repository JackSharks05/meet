# meet.jackdehaan.com — implementation plan

A private, dark-themed **AGPL fork of Timeful** for finding the best time for a
group to meet. Public site is response-only; event creation/admin is operator-
only over Tailscale; data is self-hosted.

## Decisions (locked)

| Topic | Decision |
|---|---|
| Derivation | **Fork** Timeful (`schej-it/timeful.app`) and license the result **AGPL-3.0**. |
| Backend | **Go + MongoDB in Docker** (run Timeful's real backend ≈ unchanged). |
| Public hosting | **Vercel** static frontend (respond-only) + **Cloudflare Tunnel** to the home API. |
| DNS | **Cloudflare** (zone for jackdehaan.com) → Tunnel + custom hostname is trivial. |
| Admin | Create/summary/calendar served over **Tailscale Serve**; admin API bound to the tailnet. |
| Visual style | Keep the operator's **dark "light-pillar"** brand; keep Timeful's green/yellow/red availability semantics. |
| Calendar sync | **Operator-only** Google OAuth, reusing **Mensa's** Google Cloud project, Testing mode → $0. |
| Repo | Separate **public** repo `JackSharks05/meet`. |

## Target architecture

```
PUBLIC (internet)                      HOME SERVER (Arch)              PRIVATE (tailnet)
─────────────────                      ─────────────────              ─────────────────
Respondents                            ┌── Docker ───────────┐        Operator (any device)
   │                                   │  Go/Gin server       │            │
   ▼                                   │   ┌ public listener ─┼──┐         ▼
meet.jackdehaan.com ── HTTPS ──┐       │   └ admin  listener ─┼──┼──► Tailscale Serve
(Vercel, Vue respond-only)     │       │  MongoDB (container) │  │     admin Vue build
   │ calls API                 ▼       └─────────────────────┘  │     (create/summary/cal)
api.meet.jackdehaan.com ◄─ Cloudflare Tunnel ─► public listener │◄── tailnet ──┘
   (CORS, guest, no cookies)                                     
                                       Google Calendar OAuth (operator-only)
```

- **Public path:** Vercel serves the respond-only build → `api.meet.jackdehaan.com`
  → Cloudflare Tunnel → Go **public listener** (read-event + submit-response only;
  guest, no cookies → clean cross-origin).
- **Admin path:** admin build over **Tailscale Serve** → Go **admin listener**
  bound to the tailnet interface, **never routed through the tunnel**. "Only I can
  make events" is enforced by network isolation + an admin auth check.
- **Data:** one MongoDB container shared by both listeners.

## AGPL-3.0 compliance checklist

- [x] Preserve `LICENSE` (AGPL-3.0) verbatim.
- [x] `NOTICE` with upstream attribution, forked commit, and change summary.
- [x] README states derivation + license + §13 source link.
- [ ] Mark each modified source file with a "changed by … / date" notice (§5a) as files are edited.
- [ ] Footer **"Source"** link on both public and admin builds → this repo (§13).
- [ ] Keep the published repo in sync with what is actually deployed.
- [ ] License the whole combined work AGPL-3.0 (modifications included).
- [x] Secrets via env only; never committed (`.env` gitignored, `.env.template` documents them).
- [x] Don't use Timeful's name/logo as our branding; attribution-only references.

## Phases

- **Phase 0 — Prereqs:** ✅ repo created (`JackSharks05/meet`); Cloudflare DNS confirmed;
  reuse Mensa's Google Cloud project; Docker present on the Arch box.
- **Phase 1 — Fork + license + scaffold:** ✅ copy Timeful `frontend/` + `server/`;
  add `LICENSE`/`NOTICE`/`README`/`PLAN`; wire git remote.
- **Phase 2 — Strip to core (frontend, in progress):** remove the third-party,
  billing, and marketing subsystems where it's safe and build-verified. Keep event
  creation (specific-dates + DOW), `ScheduleOverlap` grid, responses, heatmap/best-times,
  `RespondentsList`, timezone selector, optional Google overlay, CSV export.
  - ✅ **Analytics removed**: PostHog replaced with a no-op `$posthog` stub; Google Tag
    Manager + cookie-consent removed; `posthog-js` + `@gtm-support/vue2-gtm` dropped;
    poster-scan geolocation call removed.
  - ✅ **Third-party/marketing widgets removed**: Discord banner, cookie-consent dialog,
    "upvote on Reddit" snackbar, `/test` + `/stripe-redirect` routes/views.
  - ✅ **Billing/paywall neutralized**: `isPremiumUser` always true + `enablePaywall`
    default off → all features unlocked, no upgrade prompts. (Paywall is woven into the
    core `ScheduleOverlap.vue`/`ToolRow.vue`, so it is disabled rather than excised; the
    dead Premium/Donate/Upgrade UI is removed during the Phase 4 restyle.)
  - ⏳ **Deferred to Phase 3 (needs a Go compiler)**: backend removal of Stripe, Slack,
    Discord, folders/groups, and email/Listmonk. Done inside the Docker build so each cut
    is compile-verified rather than edited blind on a machine without Go.
  - ⏳ **Folded into Phases 5/6 (don't excise code we're about to replace)**: friends,
    availability groups, contacts, and the marketing Landing/dashboard Home/Settings views
    are dropped naturally when the minimal **public** (poll-only) and **admin** (create/
    summary) builds get their own slim routers — surgically gutting the soon-to-be-replaced
    dashboard now would be wasted effort.
  - Verification: `cd frontend && npm run build` stays green after each cut.
- **Phase 3 — Backend in Docker + split listeners (mostly done, verified):**
  - ✅ **Docker stack** in [`deploy/`](deploy/): multi-stage `Dockerfile` (static Go
    binary, non-root) + `docker-compose.yml` running `mongo` + `server-public` +
    `server-admin` (one image, `MODE` env). Mongo not published off-host; each listener
    bound to `127.0.0.1` (public→Cloudflare Tunnel, admin→Tailscale Serve).
  - ✅ **Runs with no external creds**: `db/init.go` reads `MONGO_URI`/`MONGO_DB`;
    `gcloud.InitTasks` skips without a service-account key (email reminders off);
    Slack/Listmonk are no-ops without config; `loadDotEnv` tolerates a missing `.env`.
  - ✅ **`main.go` rewritten**: `MODE=public` registers only auth-status + events;
    `MODE=admin` adds user/analytics/folders/swagger. `CORS_ORIGINS` + `PORT` env-driven.
    Frontend static-serving removed (API-only server).
  - ✅ **Removed (compile-verified in Docker)**: Stripe (`routes/stripe.go` + main.go
    refs) and the self-contained `discord_bot/`.
  - ✅ **Verified**: `docker build` succeeds; stack boots; `/api/auth/status` returns
    `{"error":"not-signed-in"}` on both; **`/api/user/profile` → 404 on the public
    listener but 401 on admin** (the MODE split blocks admin routes publicly).
  - ✅ **Backend hardened — "only I can make events" (verified)**: three layers.
    (1) *Network*: admin listener bound to `127.0.0.1` → Tailscale-only. (2) *Route
    registration*: `InitEvents`/`InitAuth` take a `mode` flag — the public listener
    registers ONLY respondent routes (view poll, view responses, submit/edit/delete own
    response, rename self); event create/edit/delete/duplicate/archive/decline, calendar
    overlay, sign-in, user, analytics, folders, and swagger are registered only on admin.
    (3) *Auth*: `createEvent`/`editEvent` now require `AuthRequired`, and `signInHelper`
    enforces an `ADMIN_EMAILS` operator allowlist (callers guard on `c.IsAborted()`).
    Smoke-verified: `POST /api/events`, `PUT/DELETE /api/events/:id`, `/api/auth/sign-in`,
    `/api/user/profile` all return **404 on the public listener** but **401 on admin**;
    respondent routes (`/api/auth/status`, `/api/events/:id/response`) remain on public.
  - ✅ **Dead-code removed (compile-verified + smoke-tested)**: contacts
    (`services/contacts`, `searchContacts`), Outlook (`services/microsoftgraph`,
    `services/calendar/outlook_calendar.go`, `OutlookCalendarType`, the calendar-factory
    and `services/auth` cases, `addOutlookCalendarAccount`), and folders/groups
    (`routes/folders.go`, `db/folders.go`, `models/folder*.go`, the two folder collections,
    `setEventFolder`, `InitFolders`, the event-delete folder cleanup). `go build .` is green;
    booted stack confirms removed endpoints (`/user/searchContacts`, `/user/...set-folder`,
    `/user/add-outlook-calendar-account`, `/folders`) now 404 while core routes are unchanged.
    Build now uses BuildKit cache mounts for fast rebuilds.
  - ⏳ **Left intentionally**: the email/Cloud-Tasks/Listmonk code is **neutered (inert)**
    rather than excised — it's woven through ~12 sites in the 45 KB core `events.go`
    (some storing `taskIds` on the event), so removing it risks the core scheduling logic
    for zero functional gain. Also pending: regenerate swagger (`docs/docs.go` still lists a
    couple removed endpoints — cosmetic, admin-only), prune obsolete Timeful migration
    `scripts/` (pre-existing, unrelated breakage; not in the server binary), `go mod tidy`,
    and a `/health` endpoint.
- **Phase 4 — Dark-brand restyle (foundation done, build-verified):** style inspired by
  jacksharks05.github.io (black bg, purple #621F6D→red #FE0000 light pillar, glassy cards,
  Spectral serif). Availability stays green/yellow.
  - ✅ **LightPillar** Three.js shader ported React→Vue 2 (`components/LightPillar.vue`,
    shaders verbatim); added `three` dep; mounted as a fixed full-viewport backdrop in `App.vue`.
  - ✅ **Dark foundation**: Vuetify dark theme (green primary kept), `index.css` dark tokens
    + glassy `.jdh-card` + transparent app surface, Spectral loaded in `index.html`.
  - ✅ **Shell**: glassy dark header + `meet` wordmark (dropped Logo/Premium/Donate/Feedback);
    cleaned `index.html` (removed GTM/AdSense + the broken Go-template `<title>`/meta).
  - ✅ **Landing = booking page**: the default route is now a dark hero with a "Book a time"
    CTA → the operator's Google Calendar appointment link (`VUE_APP_BOOKING_URL`, with Jack's
    link as the default); poll links (`/e/:id`) still open the grid. Replaced Timeful's
    marketing homepage. Includes the AGPL §13 source link.
  - ✅ **Dark-surface remap**: Timeful's light utility classes (`tw-bg-white`, `tw-text-black`,
    `tw-text-very-dark-gray`, light-gray borders, …) remapped to dark equivalents in
    `index.css`, scoped under `.v-application` so they win specificity. Availability
    green/yellow untouched. Removed the "formerly Schej" banner from the Event view.
  - ✅ `npm run build` green throughout.
  - ⏳ **Remaining (needs a visible poll to tune)**: per-pixel polish of the `ScheduleOverlap`
    grid + dialogs/inputs and pillar-intensity tuning behind the dense grid. Blocked on a
    test poll to preview (createEvent now needs auth → Phase 7, or a DB seed).
- **Phase 5 — Public deploy (config ready; deploy steps are operator actions):**
  - ✅ **API base configurable**: `serverURL` reads `VUE_APP_API_URL` (constants.js); the
    public build bakes in `https://api.meet.jackdehaan.com/api` (verified in the bundle).
  - ✅ **Rebrand fix**: the "add to calendar" share text no longer hardcodes
    `timeful.app` — it uses the deployment origin and says "Scheduled with meet".
  - ✅ **Build modes**: `npm run build` (public, respond-only) / `npm run build:admin`
    (sets `VUE_APP_ADMIN=true`); same for `serve` / `serve:admin`.
  - ✅ **Configs**: `frontend/vercel.json` (SPA rewrites, Vue preset),
    `deploy/cloudflared/config.yml` (tunnel → `api.meet.jackdehaan.com` → `:3002` only),
    `deploy/DEPLOY.md` runbook.
  - ⏳ **Operator actions** (in `DEPLOY.md`): run the Docker stack + `cloudflared` on the
    Arch box, set `PUBLIC_CORS_ORIGINS=https://meet.jackdehaan.com`, import the repo to
    Vercel (root dir `frontend`, env `VUE_APP_API_URL`), attach the domain, verify.
- **Phase 6 — Admin deploy (built + verified; Tailscale step is an operator action):**
  - ✅ **admin-web container**: `frontend/Dockerfile` builds the admin SPA (`VUE_APP_ADMIN=true`)
    and Caddy (`deploy/Caddyfile`) serves it + proxies `/api` **same-origin** to `server-admin`
    (so sessions work, no CORS). Added as the `admin-web` compose service on `127.0.0.1:8081`.
  - ✅ **Verified**: `GET /` → admin SPA (200), `/create` → SPA fallback (200), `/api/auth/status`
    → 401 and `/api/events/testpoll` → 200 (both proxied to server-admin).
  - ✅ **Defense-in-depth** already in place (Phase 3): admin listener is tailnet-only +
    `AuthRequired` on create + `ADMIN_EMAILS` operator allowlist.
  - ⏳ **Operator action** (`DEPLOY.md`): set `ADMIN_EMAILS`, `tailscale serve --bg 8081`,
    optionally link from Mensa. Note: sign-in / create needs **Phase 7 (OAuth)**.
- **Phase 7 — Calendar sync (operator-only):** Google OAuth via Mensa's Cloud project,
  `calendar.events.readonly`, Testing mode, localhost/tailnet redirect URI; wire
  your-calendar autofill in the admin build.
- **Phase 8 — AGPL §13:** footer "Source" link on both builds; publish repo; document
  deploy↔repo sync.
- **Phase 9 — Ops:** `docker compose` autostart (systemd) + Tunnel + Tailscale Serve as
  services; `mongodump` backup cron with an offsite copy.
- **Phase 10 — Cutover:** retire the old React `meet` (Upstash/Redis) app; redirect if needed.

## Notes / open questions

- **Cross-origin sessions:** public flow should be cookieless/guest; keep admin
  sessions same-origin (admin build + admin API both on the tailnet).
- **Mongo on Arch:** runs only inside Docker, isolated from the Node/SQLite (Mensa) stack.
- **Backups:** self-hosted data → `mongodump` cron + offsite copy is mandatory.
- **Old app:** `jdh-playground/apps/meet` (React + Upstash Redis) is superseded by this repo.
