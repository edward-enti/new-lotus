# SonLotus Web Platform — AGENTS.md

Vietnamese water-based paint brand site (sonlotus.com) rewritten in Go.
Stack: Go 1.23.12 (module `sonlotus`, stdlib ServeMux, `html/template`),
SQLite via modernc.org/sqlite v1.34.3 (pure Go, WAL, MaxOpenConns(1),
FTS5 `unicode61 remove_diacritics 2`), HTMX 2.0.7 + Alpine 3.14.8
vendored in `web/static/js/`, Tailwind standalone CLI v3.4.17.
Single binary: templates, static assets and migrations are embedded.

## Layout

- `cmd/server/main.go` — HTTP server, fixed middleware chain, graceful shutdown.
- `cmd/migrator/main.go` — `--seed-only`, `--scrape-sitemap`, `--dry-run`.
- `internal/config`, `internal/database` (migrations 001–003, `Seed`),
  `internal/domain` (incl. pure `calculator.go`), `internal/repository`,
  `internal/service` (catalog, content, seo, media), `internal/handler`
  (public + admin + uploads), `internal/middleware` (logging, recovery,
  auth, security, redirect, ratelimit, csrf, chain).
- `web/embed.go`, `web/templates/{layouts,pages,partials,admin}`,
  `web/static/{css,input.css + generated app.css,js,img}`.
- `data/` (git-ignored): SQLite DB + `uploads/`. `SPEC.md` is the source
  of truth; `PLAN.md` is the reconciled plan; `SECURITY-NOTE.md` holds
  the Supervisor-approved Phase 3 security values — do not change them
  without Supervisor approval.

## Commands

PATH in this VM must include the user-local toolchains first:

```sh
export PATH=$HOME/local/go/bin:$HOME/local/bin:$PATH
make test          # go test ./...
make vet           # go vet ./...
make fmt           # gofmt
make css           # Tailwind: input.css -> web/static/css/app.css (artifact)
make build         # css + bin/server + bin/migrator
make dev           # run server (requires ADMIN_USER/ADMIN_PASSWORD)
make seed          # migrator --seed-only (shared seed function)
go test -race ./...  # Gate 4
```

## Conventions

- Go stdlib only for routing/templates; no chi, no JS bundler, no SPA.
- All SQL parameterised; FTS input goes through `database.FTSQuery`
  (parameterised SQL alone does not cover FTS MATCH injection).
- HTML from the DB is NEVER rendered raw: escape by default
  (`html/template` auto-escaping); JSON-LD only via `encoding/json`
  emitted as `template.JS`.
- Public reads and sitemap filter `products.is_active` and
  `articles.is_published`; inactive product = 404.
- Middleware order is fixed: recovery → logging → security headers →
  redirect → rate limit → admin auth → CSRF. New routes win over
  redirect lookup; redirect `hits` increment; DB `status_code` honoured.
- Admin: env credentials only, fail closed when missing/empty; bcrypt
  cost 12 (memory-only hash); crypto/rand 32-byte tokens stored only as
  SHA-256, 12h expiry, rotation on login, invalidation on logout;
  HttpOnly + SameSite=Lax cookies, `Path=/admin`; generic login error;
  login limit 5 failed / 15 min / IP reusing the Limiter; CSRF 403 on
  rejection; uploads images + PDF only, 10 MB via MaxBytesReader,
  sanitised basename, nosniff, non-allowlisted extensions 404.
- Images: `<picture>` WebP fallback is the default (templates use
  `webpURL`); upload-time WebP conversion is NOT implemented — Go's
  stdlib has no WebP encoder and CGO would break the pure-Go guardrail.
  Flag conversion to the Supervisor instead of inventing a pipeline.
- Inquiry form: honeypot `website_url`, RemoteAddr IP limiting only
  (never `X-Forwarded-For`). Never log passwords, tokens, phone
  numbers, emails, or other PII/secrets.
- Contact placeholders stay placeholders: do not invent address,
  hotline, or Zalo link.

## Forbidden

- Do not commit `data/`, `*.db*`, `bin/`, `.env`, credentials, or the
  generated `web/static/css/app.css` (it is git-ignored; build it).
- Do not hand-edit `web/static/css/app.css`; do not render DB HTML as
  `template.HTML`; do not run a live scrape in tests/gates (migrator
  scrape is code + `--dry-run` demo only, needs separate human approval).
- Do not claim a 95–100 PageSpeed score locally; only the proxies in
  PLAN §7 (SSR without JS, semantic HTML, parseable JSON-LD, no SPA
  bundle, no render-blocking bundle beyond HTMX/Alpine).
- Do not commit or push from a Lead task unless the Supervisor says so;
  leave changes in the working tree for Peer review and human acceptance.

## Verify after changes

- `go test -count=1 ./...` and `go vet ./...` must pass; `go test -race ./...`
  for Gate 4. New admin behaviour needs an `httptest` test (no E2E/browser
  suite). Binaries: `go build -o bin/server ./cmd/server` and
  `go build -o bin/migrator ./cmd/migrator` must both build cleanly.
- Clean-directory check: copy a built binary alone to an empty directory
  and run it there (with a test admin credential) — it must serve
  without any repo files present, because templates/assets/migrations
  are embedded.
