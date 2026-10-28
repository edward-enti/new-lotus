# Phase 2 pins (recorded before template work, PLAN §8)

- Router: Go standard `net/http` `http.ServeMux` (Go 1.22+ patterns).
  No chi: no concrete routing need was recorded.
- Tailwind: standalone CLI v3.4.17
  (https://github.com/tailwindlabs/tailwindcss/releases/download/v3.4.17/tailwindcss-linux-x64,
  SHA-256 7d24f7fa191d2193b78cd5f5a42a6093e14409521908529f42d80b11fde1f1d4).
  Input: `web/static/css/input.css`. Config: `tailwind.config.js`.
  Build: `tailwindcss -c tailwind.config.js -i web/static/css/input.css -o web/static/css/app.css --minify`.
  Generated `web/static/css/app.css` is a build artifact, never hand-edited.
- HTMX 2.0.7: https://cdn.jsdelivr.net/npm/htmx.org@2.0.7/dist/htmx.min.js
  SHA-256 60231ae6ba9db3825eb15a261122d5f55921c4d53b66bf637dc18b4ee27c79f9
  vendored at `web/static/js/htmx.min.js`.
- Alpine.js 3.14.8: https://cdn.jsdelivr.net/npm/alpinejs@3.14.8/dist/cdn.min.js
  SHA-256 b600e363d99d95444db54acbfb2deffec9ae792aa99a09229bcda078e5b55643
  vendored at `web/static/js/alpine.min.js`.
- FTS sync: SQLite triggers (migration 002) + `BackfillFTS` after seed/migration.
- HTML policy: ESCAPE BY DEFAULT. `html/template` auto-escaping everywhere;
  database HTML fields (`description_html`, `application_guide_html`,
  `content_html`) are NEVER rendered as `template.HTML`.
- JSON-LD: ONLY via `encoding/json` from a marshalled struct, emitted as
  `template.JS`. A service test proves a `</script>` payload cannot break
  out (encoding/json escapes `<` as `\u003c`).
- Article JSON-LD: `schema.org/Article` on article pages; Product pages emit
  `schema.org/Product`; home emits `Organization` + `LocalBusiness` via `@graph`.
- Template inventory (Phase 2, public only): layouts `base.html`; pages
  `home.html`, `category.html`, `product_detail.html`, `articles_list.html`,
  `article_detail.html`, `calculator.html`, `contact.html`, `about.html`;
  partials `product_grid.html`, `search_results.html`, `calculator_result.html`,
  `inquiry_success.html`. Admin templates are Phase 3, not Phase 2.
- Public filtering: EVERY public read and sitemap generation filters
  `products.is_active` and `articles.is_published`; an inactive product
  returns 404 (tested in Gate 2).
- UI skill note: the ui-ux-pro-max stack search for "paint ecommerce
  product catalog" and "product catalog" (`--stack html-tailwind`) returned
  0 matches, so templates follow the skill's priority-table defaults
  (accessibility, mobile-first, semantic HTML) plus the SPEC brand palette
  and page sections — not a database-matched design system.
