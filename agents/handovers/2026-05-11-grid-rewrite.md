---
title: The Grid — Rewrite Handover
date: 2026-05-11
project: TheGrid
session: thegrid-rewrite-2026-05-11
status: rewrite-complete-pending-cutover
tags: [handover, thegrid, infra, rewrite]
derived-from:
  - type: session
    id: thegrid-rewrite-2026-05-11
    date: "2026-05-11T10:14:18Z"
    actor: mcp-agent
---

# The Grid — Rewrite Handover

Picking up a clean-rewrite of `/home/brad/Documents/GitHub/TheGrid`. New code is on disk, passing 7/7 Playwright smoke tests on `:4003`. Legacy `node index.js` is still running on `:3003` (pid `3614356`) — **not yet cut over**.

## State at handover

- **Branch:** `master`, no commits yet for the rewrite (everything is staged-only / new files on disk; legacy `index.js` and `package.json` still show as modified — those will be removed/replaced on cutover).
- **Old grid still running** on `:3003` (`node index.js`, pid `3614356`).
- **New grid verified** by running `GRID_PORT=4003 node src/server.js` and Playwright smoke suite — 7/7 passed.
- **DB:** `db/grid.sqlite` is gitignored; rebuilt from `artefacts/` on each boot via `src/reindex.js`. Test DB was wiped after smoke test, so first boot will reindex all 15 existing artefacts again.
- **Token:** `.grid-token` was generated automatically (gitignored). Skill calling `/generate` will need to read this file.

## What was done

Plan file: `/home/brad/.claude/plans/squishy-dancing-umbrella.md`. Locked decisions during planning:

- Clean rewrite, vanilla Node ≥18 + esbuild, single runtime dep `better-sqlite3`.
- Three-pane home: stats strip → services grid → artefacts grid.
- Persistent **header** nav (replaces old footer + hamburger).
- Artefacts load in **sandboxed iframes** under the header.
- SQLite registry replaces `artefacts.json` (FTS5 search, tags, pins, services, health_history).
- Skill-driven generation: `POST /generate` writes file + DB row + kiwifs companion stub.
- Categorised on-disk hierarchy: `artefacts/<category>/<path_parts...>/<slug>.html`.

### New file layout

```
src/
  server.js                regex route table, https, CSP headers
  routes/
    home.js                GET /
    artefacts.js           GET /a/* (shell), /_a/* (raw sandboxed), /api/artefacts*
    services.js            /api/services, POST /api/services/:id/start
    stats.js               /api/stats, SSE /api/events
    kiwi.js                /api/kiwi/* → http://localhost:3333 proxy
    search.js              /api/search?q=, /search redirect
    generate.js            POST /generate (token-gated)
  services/
    stats.js               /proc/stat, /proc/meminfo, df, loadavg
    health.js              http.get probes every 10s, in-memory cache + history
  db/
    index.js               better-sqlite3 + prepared statements + helpers
    migrate.js             PRAGMA user_version stepper
    migrations/001_init.sql
  util/
    paths.js, slug.js, auth.js, respond.js
  reindex.js               walk artefacts/, upsert rows, prune missing
public/
  css/grid.css
  js/shell.src.js → shell.js  (esbuild bundle: SSE, theme, drawer, iframe swap)
templates/
  shell.html, artefact-404.html
config/
  grid.config.json, services.json
scripts/
  build.js, dev.js, reindex.js
tests/
  smoke.spec.js            7 tests, all passing on :4003
playwright.config.js
.gitignore
.grid-token                (gitignored, auto-generated)
db/grid.sqlite             (gitignored, rebuilt from artefacts/ on boot)
```

### Key contracts

- **URL spaces:** `/a/<cat>/<...>` renders shell + iframe pointing at `/_a/<cat>/<...>`. The raw `/_a/*` response carries `Content-Security-Policy: sandbox allow-scripts allow-same-origin allow-forms allow-popups`.
- **`POST /generate`** — header `X-Grid-Token: <hex>`. Body: `{ slug, category, path_parts[], title, summary, html, prompt, source, tags[], kiwi_stub: { space, path } }`. Validates category against `config/grid.config.json` allowlist (`projects | research | agents | shaders | misc`). Path segments must match `^[a-z0-9][a-z0-9-]*$`, max 6 deep.
- **SSE `/api/events`** — pushes two named events every 5s: `stats` and `health`.
- **Health probe** — every 10s into in-memory snapshot + `services` + `health_history` tables. Snapshot is what `/api/services` and SSE serve.

### Verified working

- Home dashboard renders with stats strip + service health pills + artefacts grid.
- Artefact URLs load the shell + iframe with correct sandbox CSP.
- API endpoints respond (`/api/health`, `/api/stats`, `/api/services`, `/api/artefacts`).
- `/generate` works end-to-end (writes file, indexes, returns artefact URL).
- Unauthenticated `/generate` rejected with 401.
- FTS5 search returns hits.
- Kiwifs proxy returns 502 when kiwifs is down (graceful, not a crash).

## Outstanding — for next session to pick up

1. **Cutover** (manual, destructive — needs user OK):
   ```bash
   kill 3614356
   cd /home/brad/Documents/GitHub/TheGrid && npm start
   ```
   Then verify `https://100.92.148.41:3003/` shows the new dashboard from another device on Tailscale.

2. **Delete legacy files** after cutover is verified:
   - `index.js`
   - `hermes-proxy.js`
   - `test-grid.js`
   - `artefacts.json`
   (Reindex on boot already migrated their data into `db/grid.sqlite`.)

3. **Trust the self-signed cert** on every Tailscale device. Chrome/Firefox refuse to render iframes from an untrusted cert. Options:
   - `mkcert -install` and re-issue (cleanest)
   - Manually trust the existing `certs/cert.pem` per device
   - Issue trusted certs via Caddy/Tailscale-cert

4. **Decide kiwifs companion-md path convention** before the `grid-artefact` skill ships. Current existing artefact metadata has inconsistent paths (`pages/...`, `Projects/...`, `AgentVault/Projects/...`). Recommendation in plan: dedicate `<space>/artefacts/<category>/<slug>.md` so it doesn't clutter `pages/`.

5. **Build the `grid-artefact` Claude skill** (separate session — explicit user note from planning). Requirements from the user:
   - Generate responsive HTML for mobile + desktop.
   - Reference the `impeccable` design skill for design tokens / patterns / styling consistency.
   - Categorize on its own (`projects > <name> > docs`, `research > <category> > <topic>`, etc.) and `AskUserQuestion` only when ambiguous.
   - Read the token from `~/Documents/GitHub/TheGrid/.grid-token` (or wherever it ends up on the host) and POST `https://localhost:3003/generate`.
   - Validate output (single self-contained HTML, has `<meta viewport>`, mobile-first, touch targets ≥ 44px, `clamp()` typography, `dvh` for full-bleed iframes).
   - Write a companion kiwifs stub via the `kiwi_stub` payload field.

6. **Backup** `db/grid.sqlite` — it's now the source of truth for tags, pins, summaries, and `(prompt, source)` that aren't reconstructible from on-disk HTML.

## Commands the next session will use

```bash
# Dev with watch + esbuild bundling
cd /home/brad/Documents/GitHub/TheGrid && npm run dev

# Run alone on alt port while legacy still on :3003
GRID_PORT=4003 node src/server.js

# Rebuild client bundle only
npm run build

# Re-walk artefacts/ and reconcile DB
npm run reindex

# Playwright smoke suite (auto-starts on :4003)
npm test
```

## Reference docs

- Plan: `/home/brad/.claude/plans/squishy-dancing-umbrella.md`
- Project guide: `/home/brad/Documents/GitHub/TheGrid/CLAUDE.md`
- Approved decisions during planning (chat history this session): clean rewrite ✓, vanilla Node + esbuild ✓, SQLite ✓, iframe sandbox ✓, header nav ✓, three-pane home ✓, skill picks category + asks if unclear ✓, kiwifs companion stub on every generate ✓.

## Pitfalls noted

- Initial schema had `UNIQUE(category, slug)` — collided on `webgpu-shaders/{overview,index}.html`. Dropped to just `UNIQUE(path)` + non-unique index on `(category, slug)`. Anyone re-deriving from the plan should match the as-shipped schema, not the plan version.
- Server reads `GRID_PORT` env var to override the config port — added for parallel dev on `:4003` while `:3003` stays live.
- Reindex on boot is **idempotent + destructive of stale rows**: any DB row whose `path` no longer exists on disk is deleted. Don't drop tags/pins/summaries into the DB without also writing the file.
