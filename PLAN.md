# PLAN — SonLotus Web Platform Rewrite (reconciled final plan)

## 0. Goal and source of truth

- Rewrite the official web platform for Sơn Lotus (`https://sonlotus.com`), a Vietnamese
  water-based paint brand by Công ty TNHH Bích Trang, as a new project in this repo
  (`~/new-lotus`, git branch `main`, currently uncommitted).
- Source of truth: `SPEC.md` in this repo root (copied verbatim from the user-supplied
  `SonLotus_Web_Platform_Rewrite.md`, sha256 verified on copy). Where this plan and the
  SPEC disagree, the SPEC wins and the disagreement must be reported to the Supervisor.
  One deliberate interpretation is recorded in §7 Gate 1 (calculator test placement);
  it changes no SPEC coverage, only where a pure function lives, and is flagged there.
- This plan covers implementation only. Deployment hosting, a live PageSpeed run, and a
  live scrape of the legacy site are out of scope for local acceptance (see §7, §9).

## 1. Stack guardrails (strict, from SPEC §2)

- Go 1.23+, HTTP router: Go standard `net/http` `http.ServeMux` by default;
  `github.com/go-chi/chi/v5` only on a recorded concrete routing need. Do NOT introduce
  Webpack/Vite/React/Next.js or any other unlisted framework or JS bundler pipeline.
- Templating: standard `html/template` with a typed ViewData wrapper (preferred by SPEC
  over `a-h/templ` to minimise code-gen friction across agents).
- Frontend: HTMX 2.x vendored at `web/static/js/htmx.min.js`, Alpine.js 3.x vendored at
  `web/static/js/alpine.min.js`, for partial updates and local UI state only. No SPA.
  Exact pinned versions, source URLs and SHA-256 hashes must be recorded at Gate 0
  (see §8) and verified before the vendored files are treated as final.
- Styling: Tailwind CSS, default mode = standalone CLI (see §8). Input stylesheet is
  committed; the generated `web/static/css/app.css` is a build artifact, never hand-edited.
  Brand palette: Lotus Green `#15803d` / `#166534`, Wood/Amber `#b45309` / `#d97706`,
  Neutral Slate `#0f172a` / `#f8fafc`.
- Database: SQLite 3 via `modernc.org/sqlite` (pure Go, no CGO) with
  `PRAGMA journal_mode = WAL; synchronous = NORMAL; busy_timeout = 5000; foreign_keys = ON;`
  set on EVERY connection (DSN pragma or connect hook — listing them once is not enough,
  because these pragmas are per-connection in SQLite). Single-writer policy for SQLite:
  `MaxOpenConns(1)` or a serialised writer; record the chosen policy. Module versions are
  pinned in `go.mod`/`go.sum` and verified with `go mod verify`.
- Search: SQLite FTS5 virtual table `search_index`, `tokenize = 'unicode61 remove_diacritics 2'`,
  so Vietnamese search works with or without diacritics. Synchronisation: SQLite triggers
  (path-independent; see §3). All public search goes through an FTS query helper that
  quotes/escapes user input, enforces a length cap, and handles empty queries safely —
  parameterised SQL alone does not cover FTS `MATCH` injection.
- Migrator parsing: `github.com/PuerkitoBio/goquery` (pinned, verified as above).
- Logging: standard library `log/slog` (JSON in production, text in development).
  Never log passwords, session tokens, phone numbers, emails, or other PII/secrets.

## 2. Repository layout (SPEC §3, implement exactly)

```
cmd/server/main.go, cmd/migrator/main.go
internal/config, internal/database (migrations 001_initial_schema.sql, 002_fts_search.sql, db.go, seed.go),
internal/domain (product.go, article.go, inquiry.go, seo.go, calculator.go — pure calculator math, see §7 Gate 1),
internal/repository (category, product, article, inquiry, redirect),
internal/service (catalog, content, seo, media),
internal/handler (render, public, content, inquiry, admin),
internal/middleware (logging, recovery, auth, security, redirect)
web/embed.go, web/static/{css,js,img}, web/templates/{layouts,pages,partials,admin}
data/ (git-ignored: SQLite DB file and uploaded media), Dockerfile, Makefile, .env.example, go.mod (module sonlotus), SPEC.md, PLAN.md
```

- Create `.gitignore` in Phase 1 BEFORE any seed data is generated or any server run,
  covering at least `data/`, `*.db*` (including `-wal`/`-shm`), `bin/`, `.env`.
  Never commit a database, uploaded media, `.env`, or credentials.
  `.env.example` contains placeholders only, never a real or default admin credential.
- The template `AGENTS.md` created by `slp-init.sh` is still placeholder text. After the
  initial implementation, fill it in from the actual repo state (stack, test/lint/build
  commands, conventions, verify steps) — skill `init-deep` may be used for this step —
  and let the Supervisor review it before acceptance.
- Single binary (SPEC §1): templates, static assets AND migrations are embedded
  (`//go:embed` in `web/embed.go` and the migration embed), so the built binary must run
  from a clean directory with no repo files present. Verify this with a clean-directory
  smoke check in Phase 4.

## 3. Data, seed, and search

- Migrations implement SPEC §4 verbatim: `categories`, `products`, `articles`,
  `inquiries`, `url_redirects`, plus the FTS5 `search_index` virtual table.
- FTS synchronisation: SQLite triggers on insert/update/delete of products and articles.
  After every seed and every migration, run an FTS backfill so `search_index` cannot go
  stale when a write bypasses the repository layer. Gate 1 must include a test that an
  invalid `category_id` insert fails (proving `foreign_keys` is actually on).
- Seed data (SPEC §5), defined idempotently: if `categories` count is 0, seed in ONE
  transaction, else no-op. One shared seed function is used by both server boot and
  `migrator --seed-only` — never two seed paths. `make seed` runs `migrator --seed-only`.
- Categories (exact slugs): `son-go-he-nuoc`, `son-gia-go`, `son-kim-loai-he-nuoc`,
  `son-ngoi-be-tong`, `son-loa`, `son-composite-nhua`, `tram-tret-chong-tham`.
- Products (exact slugs, SPEC §5.2, listed verbatim to prevent transcription error):
  `son-phu-go-he-nuoc-lotus-acrylic-lacquer`,
  `son-lot-go-he-nuoc-lotus-wood-primer`,
  `bot-tret-go-he-nuoc-lotus-wood-filler`,
  `son-gia-go-lotus-sat-ma-kem-xi-mang`,
  `son-kim-loai-ma-kem-he-nuoc-lotus`,
  `son-loa-tao-gai-san-lotus-speaker-paint`,
  `son-ngoi-cao-cap-he-nuoc-lotus`,
  `keo-bong-nuoc-chong-tham-lotus-clear-sealer`.
  Seed copy is realistic Vietnamese draft copy for Sơn Lotus, not lorem ipsum and not
  polished marketing copy — Phase 1 must not balloon into copywriting. Seed data for
  categories and products must be authentic (specs, packaging, colours, coverage rates).
- Paint calculator formula (SPEC §6.1): `Total Liters = Area (m2) × Coats / Coverage Rate × 1.05`
  (5% wastage), plus an optimal packaging mix (e.g. 18L + 3.8L + 0.8L). Unit tests must
  cover zero/negative area, zero coverage rate, zero coats, and packaging-mix edge cases
  (zero area → no packaging; negative coverage/coats rejected).
- Public visibility: EVERY public read and `sitemap.xml` generation must filter
  `products.is_active` and `articles.is_published`. An inactive product must return 404
  (tested in Gate 2). Article-type route mapping is fixed: `/du-an` = `project`,
  `/ho-tro-ky-thuat` = `tech_support`, `/giai-phap` = `solution`.

## 4. Routes and behaviour (SPEC §6)

- Public routes: `/`, `/san-pham` (HTMX-filterable, `hx-target="#catalog-grid"`),
  `/danh-muc/{slug}`, `/san-pham/{slug}` (Alpine colour-swatch previewer, specs table,
  application guide, TDS/MSDS links, embedded calculator, quote modal),
  `/cong-cu/tinh-luong-son`, `POST /api/htmx/tinh-luong-son` (HTML partial),
  `/du-an`, `/ho-tro-ky-thuat`, `/giai-phap`, `/bai-viet/{slug}`,
  `/gioi-thieu`, `/lien-he` (GET form + POST inquiry), `GET /api/htmx/tim-kiem`
  (FTS5 live search partial), `/sitemap.xml`, `/robots.txt` (points to sitemap, disallows `/admin`).
- Home page sections (SPEC §6.1): hero banner, 7 category cards, "Why Water-Based Paint?"
  eco-benefits, featured products, paint calculator widget, latest projects & tech guides.
  Contact page includes factory/office address, Google Maps embed, hotline, dealer
  registration & inquiry form.
- Admin routes under `/admin/*`: login/logout, dashboard (totals + unread inquiries +
  top redirects), product CRUD, article CRUD, inquiry status management
  (`new` → `contacted` → `quoted` → `closed`), `POST /admin/upload` (multipart, saves to
  `./data/uploads/`, served publicly at `/uploads/*` — see upload rules in §6).
- Templates (SPEC §3 inventory, all required): layouts `base.html` (header, nav, footer,
  floating Zalo/hotline widget), `admin.html`; pages `home.html`, `category.html`,
  `product_detail.html`, `articles_list.html`, `article_detail.html`, `calculator.html`,
  `contact.html`, `about.html`; partials `product_grid.html`, `search_results.html`,
  `calculator_result.html`, `inquiry_success.html`; admin `login.html`, `dashboard.html`,
  `products_list.html`, `product_form.html`, `articles_list.html`, `article_form.html`,
  `inquiries_list.html`.
- Every public page passes a `PageSEO` struct to `base.html` (Title, Description,
  CanonicalURL, OGImage, OGType, JSON-LD). JSON-LD is built ONLY with `encoding/json`
  (never string concatenation) from a marshalled struct and emitted as `template.JS`;
  a test must prove a `</script>` payload cannot break out of the JSON-LD block.
  Product pages emit `schema.org/Product` JSON-LD; article pages emit `schema.org/Article`
  JSON-LD; the homepage emits `Organization` + `LocalBusiness` JSON-LD.
- HTML fields from the database (`description_html`, `application_guide_html`,
  `content_html`) are NEVER rendered raw by default. The pre-Phase 2 pin (§8) must state
  one policy and apply it everywhere: escape by default, OR sanitise on save with a stated
  allowlist. Scraped HTML follows the same policy (see §5).
- Floating quick-action widget (sticky bottom-right): Gọi Hotline, Chat Zalo, Tính Lượng Sơn.
- Inquiry form: honeypot field `website_url` must be empty, server-side validation,
  length caps on name/message, IP rate-limiting on `POST /lien-he` using the client IP
  from `RemoteAddr` (never an untrusted `X-Forwarded-For`). All SQL must be parameterised.

## 5. Migrator CLI (SPEC §7)

- `cmd/migrator` flags: `--seed-only` (migrations + seed), `--scrape-sitemap`
  (fetch `https://sonlotus.com/sitemap_index.xml` or `/sitemap.xml`, parse URLs, extract
  title/meta/H1/content/featured image, populate products, articles, `url_redirects`),
  `--dry-run` (log discovered URLs and parsed structs, no database writes).
- Polite scraping: max 2 requests/second, 10s timeout, User-Agent `SonLotus-Migrator/1.0`.
- Migrator trust rules (one-line constraints, all mandatory): fetch ONLY `sonlotus.com`
  hosts (allowlist); response-size and redirect caps; never skip TLS verification;
  slug collision = idempotent upsert-or-skip policy, stated in code (never duplicate on
  re-run); scraped HTML sanitised per the §4 policy before storage; `--dry-run` must
  create NO database file and open NO writable database.
- Safety: do NOT run a live scrape during development, tests, or verification gates.
  The migrator's scrape path is delivered as code plus a `--dry-run` demonstration only;
  any real scrape of the legacy site needs separate human approval.

## 6. Security — minimum defaults (SPEC §8 is thin; these defaults are now plan requirements)

The SPEC names session-cookie auth, `HttpOnly` + `SameSite=Lax` cookies, CSRF protection,
security headers, rate limiting and parameterised SQL, but does not specify the details.
The Lead must implement the following minimum defaults, and before Phase 3 begins must
submit a SHORT security note (a table, not a full design document) to the Supervisor for
approval. Phase 3 does not start until that note is approved. The note must cover, at
minimum, every item below; the Lead proposes the exact values (expiry duration, upload
size cap, header list) in the note and the Supervisor approves them there — the plan does
not invent those numbers now.

- Admin credentials: from `ADMIN_USER`/`ADMIN_PASSWORD` env vars only. Passwords are
  verified by hash comparison only; never compare, log, or ship a plaintext password.
  No default credential in code, `.env.example`, or config defaults. The server and the
  login handler FAIL CLOSED (refuse to start / refuse login) when admin env vars are
  missing or empty.
- Sessions: `crypto/rand` session tokens, stored hashed server-side, `HttpOnly` +
  `SameSite=Lax` cookies, an expiry, rotation on login, and invalidation on logout.
- Auth coverage: every `/admin/*` handler (pages, POSTs, upload, status changes) requires
  a valid admin session, except login GET/POST themselves. Unauthenticated `/admin/*`
  access redirects to login.
- Login hardening: generic login-failure error (no user-vs-password oracle), and login
  rate limiting that REUSES the existing rate limiter (no separate progressive-delay
  system is required).
- CSRF: required on ALL state-changing POSTs, including login, admin CRUD, upload,
  inquiry status changes, and the inquiry form. CSRF rejection is tested (Gate 3).
- Uploads: admin session + CSRF required; extension/MIME allowlist = images + PDFs only;
  size cap enforced with `MaxBytesReader` (value proposed in the security note);
  filename sanitised to a safe basename (no `..`, no path traversal); `/uploads/*`
  serves with `X-Content-Type-Options: nosniff`, no directory listing, and a safe
  `Content-Disposition` policy. Uploaded `.html`/`.svg`/`.js` must never be servable as
  executable same-origin content.
- Security headers: an explicit header set, stated in the security note and applied by
  the security middleware (including `nosniff` and a content-security posture that does
  not break HTMX/Alpine).
- Middleware order (fixed): recovery → logging → security headers → redirect →
  rate limit → admin auth → CSRF. Redirect middleware: new routes win over redirect
  lookup; a redirect lookup must not hijack a real route; `hits` increments;
  `status_code` is honoured from the database; loop guard (`old_path == new_path`, or a
  new path that is itself an old path, must not loop).

## 7. Phases and verification gates (SPEC §9, sequential, no skipping)

- Gate 0 — Environment (NEW, before Phase 1): install Go 1.23+ from the official
  distribution to a USER-LOCAL path (no system-wide install assumption — a system install
  may need permissions the Lead does not have), record the install path and `go version`
  evidence, and run a tiny `go test` smoke check. Pin the exact installed Go version in
  `go.mod` when it is created. Also at Gate 0, record the Tailwind mode (§8 default:
  standalone CLI), HTMX/Alpine pinned versions + source URLs + SHA-256, and confirm
  network access for `go mod download`. If Gate 0 fails, STOP and escalate to the
  Supervisor — do not write implementation code. Dockerfile Go-version parity is a
  Phase 4 check, not a Gate 0 blocker. If Gate 0 shows FTS5 `remove_diacritics 2` is
  unsupported by the chosen SQLite build, that finding returns to the Supervisor before
  Phase 1 proceeds — the Lead must not invent a replacement search engine.
- Phase 1 — Init, database, FTS5, seed: `.gitignore` (BEFORE any seed/run, §2),
  `go.mod` (`module sonlotus`), `Makefile`, `.env.example`, directory tree,
  `internal/config`, `internal/domain` (including pure `calculator.go`, see below),
  `internal/database`, `internal/repository`.
  Gate 1: unit tests covering product creation, category filtering, Vietnamese FTS5
  search without diacritics (e.g. `"son gia go"` finds `"Sơn Giả Gỗ Lotus"`), calculator
  math, and the FK-failure test (§3). `go test -v ./internal/...` MUST PASS 100%.
  Calculator placement (reconciled interpretation, flagged per §0): the pure calculator
  math lives in `internal/domain/calculator.go` (no I/O, no repository dependency) so
  the SPEC-named test file `internal/repository/catalog_repo_test.go` can test it
  without a layer inversion (repository tests must not import the service package, which
  would create a cycle). `internal/service` wraps the same pure function in Phase 2.
  This preserves SPEC Gate 1 coverage and file location; it adds one small pure-function
  file to `internal/domain`, which SPEC §3 lists only for structs. If the Lead finds
  this unworkable, it must ask the Supervisor before moving the test instead.
- Phase 2 — Services, public handlers, templates: `internal/service` (catalog, content,
  seo, media), `web/embed.go`, Tailwind templates (§4 inventory), public/content handlers.
  Pre-Phase 2 pins (§8) must be recorded before template work starts: router, Tailwind
  mode + input + build command, HTMX/Alpine pins, FTS triggers + backfill, HTML/JSON-LD
  policy, Article JSON-LD, template inventory.
  Gate 2: HTTP integration tests in `internal/handler/public_handler_test.go` verifying,
  for `/`, `/san-pham`, `/danh-muc/son-go-he-nuoc`,
  `/san-pham/son-gia-go-lotus-sat-ma-kem-xi-mang`, `/sitemap.xml`, and
  `POST /api/htmx/tinh-luong-son`: HTTP 200 AND `Content-Type: text/html` AND a
  page-specific marker (title/heading) — "valid HTML" means exactly that, no HTML-validator
  library is added. Plus: public filtering tests (`is_active`/`is_published`) and an
  inactive-product 404 test. `go test -v ./...` MUST PASS 100%.
- Phase 3 — Admin CMS, lead management, migrator, server wiring: admin session
  middleware, `/admin/*` CRUD + templates, `cmd/migrator`, `cmd/server/main.go`
  (routes, logging, recovery, 301 redirects, graceful shutdown — shutdown closes the
  database; an explicit WAL checkpoint is at most a code comment, not a gate).
  Phase 3 starts ONLY after the Supervisor approves the §6 security note.
  Gate 3: minimal `httptest` functional tests (NO E2E/browser suite) covering:
  unauthenticated `/admin/*` redirect, generic login failure, CSRF rejection, session
  invalidation on logout, upload rejection cases, and inquiry status transition.
  Plus: `go build -o bin/server ./cmd/server` and `go build -o bin/migrator ./cmd/migrator`
  must both build cleanly (wording corrected: "build cleanly", not "zero warnings" —
  `go vet` is Gate 4).
- Phase 4 — Production packaging and final audit: multi-stage `Dockerfile` (pinned base
  image, non-root user, `.dockerignore` excluding `data/`, `.env`, `.git`, `bin/`),
  `Makefile` (`make dev`, `make build`, `make test`, `make seed` = `migrator --seed-only`),
  filled-in `AGENTS.md`, clean-directory single-binary smoke check (§2).
  Gate 4: `go vet ./...` clean and `go test -race ./...` passes (zero data races).
- A phase is not "done" on a green test run alone: the Lead reports what changed (files),
  test/build output, anything to verify, and open questions to the Supervisor after each
  phase, and the Supervisor decides whether to proceed.
- PageSpeed proxies (local, checkable; the 95–100 number itself is post-deployment only
  and must never be claimed locally): SSR content present without JavaScript, semantic
  HTML, parseable JSON-LD, no SPA bundle, image size/format policy (see WebP below),
  no render-blocking vendored bundle beyond HTMX/Alpine. No local Lighthouse tooling is
  added.
- WebP (SPEC §1 objective, previously unassigned): assigned to the Phase 3 media
  decision. Default: `<picture>` WebP fallback for static/seed images in Phase 2
  templates; upload-time WebP conversion in `media_service` only if it stays simple.
  If conversion proves disproportionate, the Lead flags it to the Supervisor before
  Phase 3 instead of inventing a converter pipeline. WebP must not be silently dropped.

## 8. Phase checkpoints (former §8 open points, now timed — no decision may be guessed)

- Gate 0: Go version + install path + smoke test; Tailwind mode (default standalone CLI,
  input file + build command recorded; if the CLI is unavailable in this VM, the Lead
  escalates to the Supervisor BEFORE switching to pre-compiled `app.css`, which would
  then require committing `app.css` with its stated source); HTMX/Alpine pinned versions
  + source URLs + SHA-256; `go mod download` network check.
- Pre-Phase 2: router = ServeMux (chi only on recorded concrete need); FTS sync =
  triggers + backfill; HTML policy (§4) + JSON-LD via `encoding/json`; Article JSON-LD;
  template inventory (§4); seed copy accepted as draft copy.
- Pre-Phase 3: Supervisor-approved §6 security note (table, not a full design document).
- Phase 4: Dockerfile pin/parity, PageSpeed proxies (§7), WebP media decision (§7),
  clean-directory binary check.
- Vietnamese content beyond Gate 1: diacritic-insensitive search coverage beyond the
  single example query, tested with at least one additional accented/unaccented pair.

## 9. Acceptance

- Acceptance stays with the human. After Phase 4, an independent Peer review
  (read-only, with `/home/hatch/slp-kit/review-checklist.md`) must return APPROVE or
  NEEDS CHANGES with file:line evidence. If NEEDS CHANGES, the Lead fixes, the Peer
  re-reviews, and if it still fails, the Supervisor escalates to the human instead of
  looping.
- Final deliverable for human acceptance: the repo in `~/new-lotus`, the Gate 0–4
  outputs, the Peer review verdict, and a summary of what was verified vs what still
  needs a live deployment check (PageSpeed number, live scrape, production admin
  credentials).
- Probe gate: SKIPPED, confirmed by both plan reviewers. The SPEC fixes the stack,
  schema, routes, seed data and phases, so there is no genuinely unclear architectural
  choice with two viable approaches that a timeboxed probe experiment would resolve.
  The only unproven technical risks (Go availability, FTS5 `remove_diacritics 2`,
  Tailwind CLI availability) are covered by Gate 0, the Gate 1 FTS test, and the §8
  Tailwind checkpoint — not by a probe round. If Gate 0/Gate 1 shows FTS
  `remove_diacritics 2` unsupported, that finding returns to the Supervisor before
  Phase 1 proceeds.

## 10. Work allocation

- Lead: all implementation, Gate 0 then phases 1–4 in order, self-contained brief, skills
  `tdd` and `diagnosing-bugs` for Lead work, `ui-ux-pro-max` for UI work, `init-deep`
  for the AGENTS.md fill-in, in `/home/hatch/new-lotus/.agents/skills/`.
  The Lead must NOT commit or push; leave all changes in the working tree for Peer
  review and human acceptance.
- Peer (review only, after Phase 4): skills `code-review` and `ship-gate`, plus the
  review checklist at `/home/hatch/slp-kit/review-checklist.md`.
- Chloe (plan review, done): skill `security-auditor`. Emily (plan review, done):
  simplicity/feasibility lens.
- No worker may launch other agents. One writer per moving scope; concurrent writers,
  if any, use separate worktrees — the default here is sequential Lead work, so no
  worktree split is needed.

## 11. Reconciliation log (plan review gate, 2026-10-27)

- Chloe verdict: NEEDS CHANGES (correctness/security). Emily verdict: NEEDS CHANGES
  (simplicity/feasibility). Both verdicts were reconciled into this version; the
  original draft's §8 open points are now §8 timed checkpoints, and its §6 "implement
  conservatively" is now §6 minimum defaults plus a Supervisor-approved pre-Phase 3
  security note.
- Every Chloe finding H1–H8 and M1–M8 was judged REAL ISSUE by Emily and is carried
  into this plan. Every Chloe Low item L1–L3 is carried in. L4 (explicit WAL checkpoint
  on graceful shutdown) was judged OVERCAUTION by Emily and is NOT a gate: shutdown
  closes the database; a checkpoint is at most a code comment.
- Emily trims applied (not disagreements on the findings): the §6 security note is a
  short table, not a full design document inside PLAN.md; login limiting reuses the
  existing limiter (no separate progressive-delay system); Dockerfile Go-version parity
  is a Phase 4 check, not a Gate 0 blocker; Tailwind is pinned in ONE chosen mode
  (standalone CLI default), not both modes.
- Calculator placement (§7 Gate 1) is the one place this plan interprets SPEC §9 rather
  than copying it literally: pure math in `internal/domain/calculator.go`, tested from
  the SPEC-named repository test file, wrapped by the service in Phase 2. Coverage is
  unchanged; the interpretation is flagged here and in §7 for the human to see.
