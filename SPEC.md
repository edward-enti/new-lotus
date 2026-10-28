# Technical Specification: SonLotus Web Platform Rewrite (`sonlotus.com`)

## 1. Project Overview & Objectives
Rewrite the official web platform for **Sơn Lotus (`https://sonlotus.com`)** — a Vietnamese brand by Công ty TNHH Bích Trang specializing in eco-friendly, water-based paints (Sơn hệ nước: Sơn Giả Gỗ, Sơn Gỗ, Sơn Kim Loại, Sơn Ngói/Bê Tông, Sơn Loa, Sơn Composite, Trám Trét & Chống Thấm).

### Primary Objectives
1. **Performance & SEO:** Achieve 95–100 Google PageSpeed Insights scores across Mobile and Desktop via Server-Side Rendering (SSR), semantic HTML5, JSON-LD structured data, and automated WebP image optimization.
2. **Zero-Dependency Deployment:** Package templates, static assets, and database migrations into a single Go binary backed by SQLite in WAL mode.
3. **Interactive UX without SPA Complexity:** Utilize HTMX and Alpine.js for instant category filtering, live product search, color swatch previews, and a Paint Coverage Calculator.
4. **Seamless Content Migration:** Include a built-in CLI scraper/migrator (`cmd/migrator`) to extract existing categories, products, images, and technical articles from the legacy `sonlotus.com` site and populate the new database while preserving SEO URLs (or generating 301 redirect maps).

---

## 2. Technology Stack & Guardrails (Strict)
AI Coding Agents MUST adhere strictly to the following technology choices. Do NOT introduce unlisted frameworks or JavaScript bundler pipelines (no Webpack/Vite/React/Next.js).

- **Language:** Go 1.23+
- **HTTP Router:** Go 1.22+ standard `net/http` (`http.ServeMux`) or `github.com/go-chi/chi/v5`.
- **HTML Templating:** Standard `html/template` with clean layout/partial composition (or `github.com/a-h/templ` if configured in Makefile). To minimize external code-gen friction across agents, prefer standard `html/template` with a typed ViewData wrapper unless instructed otherwise.
- **Frontend Interactivity:**
  - **HTMX 2.x** (vendored in `web/static/js/htmx.min.js`) for partial page updates, filtering, search, and form submissions.
  - **Alpine.js 3.x** (vendored in `web/static/js/alpine.min.js`) for local UI state (mobile drawer, image gallery lightbox, color swatch switcher, calculator tabs).
- **Styling:** **Tailwind CSS** (via standalone CLI or pre-compiled stylesheet in `web/static/css/app.css`). Brand palette:
  - Primary Lotus Green: `#15803d` / `#166534`
  - Accent Wood/Amber: `#b45309` / `#d97706`
  - Neutral Slate: `#0f172a` / `#f8fafc`
- **Database:** **SQLite 3** (`modernc.org/sqlite` — pure Go, no CGO required) configured with:
  - `PRAGMA journal_mode = WAL;`
  - `PRAGMA synchronous = NORMAL;`
  - `PRAGMA busy_timeout = 5000;`
  - `PRAGMA foreign_keys = ON;`
- **Search Engine:** SQLite FTS5 (Full-Text Search) virtual tables for instant Vietnamese product and article search.
- **HTML Parsing (Migrator):** `github.com/PuerkitoBio/goquery` for scraping legacy `sonlotus.com` content.
- **Logging:** Standard library `log/slog` (structured JSON in production, text in development).

---

## 3. Repository & Directory Structure

```text
sonlotus-web/
├── cmd/
│   ├── server/
│   │   └── main.go               # Main HTTP server entrypoint & graceful shutdown
│   └── migrator/
│       └── main.go               # CLI tool to seed DB & scrape legacy sonlotus.com
├── internal/
│   ├── config/
│   │   └── config.go             # Environment variable loader with sensible defaults
│   ├── database/
│   │   ├── migrations/           # Embedded SQL migration files
│   │   │   ├── 001_initial_schema.sql
│   │   │   └── 002_fts_search.sql
│   │   ├── db.go                 # SQLite connection, PRAGMA setup & migration runner
│   │   └── seed.go               # Initial seed data for Sơn Lotus categories & products
│   ├── domain/
│   │   ├── product.go            # Product, Category, ColorSwatch, Spec structs
│   │   ├── article.go            # Article (Project, TechSupport, News, Solution) structs
│   │   ├── inquiry.go            # Customer Lead / Quote request structs
│   │   └── seo.go                # Meta tags, OpenGraph, JSON-LD structs
│   ├── repository/
│   │   ├── category_repo.go
│   │   ├── product_repo.go
│   │   ├── article_repo.go
│   │   ├── inquiry_repo.go
│   │   └── redirect_repo.go
│   ├── service/
│   │   ├── catalog_service.go    # Product filtering, related products, paint calculator
│   │   ├── content_service.go    # Articles, projects, solutions, FTS5 search
│   │   ├── seo_service.go        # Sitemap XML generator, JSON-LD builder
│   │   └── media_service.go      # File upload & image handling
│   ├── handler/
│   │   ├── render.go             # Template cache, layout renderer, HTMX detection helper
│   │   ├── public_handler.go     # Home, Category, Product, Calculator, Static pages
│   │   ├── content_handler.go    # Projects, Tech Support, Search, Sitemap, Robots
│   │   ├── inquiry_handler.go    # Contact & Quote request POST handlers
│   │   └── admin_handler.go      # Admin Auth & CRUD handlers
│   └── middleware/
│       ├── logging.go            # Request logging with slog
│       ├── recovery.go           # Panic recovery
│       ├── auth.go               # Admin session authentication
│       ├── security.go           # Security headers, rate limiting, CSRF protection
│       └── redirect.go           # Legacy URL 301 redirect middleware
├── web/
│   ├── embed.go                  # //go:embed static templates
│   ├── static/
│   │   ├── css/
│   │   ├── js/
│   │   └── img/
│   └── templates/
│       ├── layouts/
│       │   ├── base.html         # Public master layout (Header, Nav, Footer, Floating Zalo/Hotline)
│       │   └── admin.html        # Admin dashboard layout
│       ├── pages/
│       │   ├── home.html
│       │   ├── category.html
│       │   ├── product_detail.html
│       │   ├── articles_list.html
│       │   ├── article_detail.html
│       │   ├── calculator.html
│       │   ├── contact.html
│       │   └── about.html
│       ├── partials/
│       │   ├── product_grid.html
│       │   ├── search_results.html
│       │   ├── calculator_result.html
│       │   └── inquiry_success.html
│       └── admin/
│           ├── login.html
│           ├── dashboard.html
│           ├── products_list.html
│           ├── product_form.html
│           ├── articles_list.html
│           ├── article_form.html
│           └── inquiries_list.html
├── data/                         # Ignored by git: stores SQLite DB file & uploaded media
├── Dockerfile
├── Makefile
├── .env.example
├── go.mod
└── SPEC.md
```

---

## 4. Database Schema Specification

### 4.1. `001_initial_schema.sql`

```sql
-- Categories (Danh mục sản phẩm chính)
CREATE TABLE IF NOT EXISTS categories (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    slug TEXT NOT NULL UNIQUE,
    short_description TEXT NOT NULL DEFAULT '',
    description_html TEXT NOT NULL DEFAULT '',
    surface_type TEXT NOT NULL DEFAULT 'general', -- 'wood', 'faux_wood', 'metal', 'concrete_tile', 'speaker', 'composite', 'putty_waterproof'
    icon_name TEXT NOT NULL DEFAULT 'paint-roller',
    thumbnail_url TEXT NOT NULL DEFAULT '',
    sort_order INTEGER NOT NULL DEFAULT 0,
    is_active BOOLEAN NOT NULL DEFAULT 1,
    seo_title TEXT NOT NULL DEFAULT '',
    seo_description TEXT NOT NULL DEFAULT '',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Products (Sản phẩm sơn hệ nước)
CREATE TABLE IF NOT EXISTS products (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    category_id INTEGER NOT NULL REFERENCES categories(id) ON DELETE RESTRICT,
    sku TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL,
    slug TEXT NOT NULL UNIQUE,
    coating_role TEXT NOT NULL DEFAULT 'topcoat', -- 'primer' (Sơn lót), 'topcoat' (Sơn phủ), 'putty' (Trám trét), 'sealer' (Chống thấm), 'stain' (Màu tinh)
    finish_type TEXT NOT NULL DEFAULT 'gloss',    -- 'gloss' (Bóng), 'matte' (Mờ), 'satin' (Bán bóng), 'flat' (Không bóng)
    usage_area TEXT NOT NULL DEFAULT 'both',      -- 'indoor' (Nội thất), 'outdoor' (Ngoại thất), 'both' (Trong & Ngoài trời)
    short_summary TEXT NOT NULL DEFAULT '',
    description_html TEXT NOT NULL DEFAULT '',
    application_guide_html TEXT NOT NULL DEFAULT '',
    features_json TEXT NOT NULL DEFAULT '[]',     -- JSON array of strings: ["Hệ nước 100% không mùi", "Kháng tia UV", "Chống rỉ sét"]
    specs_json TEXT NOT NULL DEFAULT '{}',        -- JSON object: {"coverage_m2_per_liter": 10.0, "recommended_coats": 2, "dry_time_touch_mins": 30, "dry_time_recoat_hours": 2, "dilution": "Nước sạch 5-10%"}
    packaging_json TEXT NOT NULL DEFAULT '[]',    -- JSON array: ["Lon 0.8L", "Lon 3.8L", "Thùng 18L / 20Kg"]
    colors_json TEXT NOT NULL DEFAULT '[]',       -- JSON array: [{"code": "LW-01", "name": "Cánh Gián", "hex": "#6E2C00", "swatch_url": ""}]
    images_json TEXT NOT NULL DEFAULT '[]',       -- JSON array of image URLs
    thumbnail_url TEXT NOT NULL DEFAULT '',
    tds_file_url TEXT NOT NULL DEFAULT '',        -- Technical Data Sheet PDF
    msds_file_url TEXT NOT NULL DEFAULT '',       -- Safety Data Sheet PDF
    coverage_rate REAL NOT NULL DEFAULT 8.0,      -- Average m2/liter or m2/kg per coat (used by Paint Calculator)
    is_featured BOOLEAN NOT NULL DEFAULT 0,
    is_active BOOLEAN NOT NULL DEFAULT 1,
    seo_title TEXT NOT NULL DEFAULT '',
    seo_description TEXT NOT NULL DEFAULT '',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_products_category ON products(category_id);
CREATE INDEX IF NOT EXISTS idx_products_role ON products(coating_role);
CREATE INDEX IF NOT EXISTS idx_products_featured ON products(is_featured, is_active);

-- Articles (Dự án, Hỗ trợ kỹ thuật, Giải pháp, Tin tức)
CREATE TABLE IF NOT EXISTS articles (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    type TEXT NOT NULL CHECK(type IN ('project', 'tech_support', 'solution', 'news')),
    title TEXT NOT NULL,
    slug TEXT NOT NULL UNIQUE,
    excerpt TEXT NOT NULL DEFAULT '',
    content_html TEXT NOT NULL DEFAULT '',
    thumbnail_url TEXT NOT NULL DEFAULT '',
    author_name TEXT NOT NULL DEFAULT 'Kỹ Thuật Sơn Lotus',
    related_category_id INTEGER REFERENCES categories(id) ON DELETE SET NULL,
    is_featured BOOLEAN NOT NULL DEFAULT 0,
    is_published BOOLEAN NOT NULL DEFAULT 1,
    published_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    seo_title TEXT NOT NULL DEFAULT '',
    seo_description TEXT NOT NULL DEFAULT '',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_articles_type_pub ON articles(type, is_published, published_at DESC);

-- Inquiries (Yêu cầu báo giá, tư vấn kỹ thuật, đăng ký đại lý)
CREATE TABLE IF NOT EXISTS inquiries (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    customer_name TEXT NOT NULL,
    phone TEXT NOT NULL,
    email TEXT NOT NULL DEFAULT '',
    company_name TEXT NOT NULL DEFAULT '',
    inquiry_type TEXT NOT NULL DEFAULT 'quote' CHECK(inquiry_type IN ('quote', 'tech_support', 'dealer', 'general')),
    product_id INTEGER REFERENCES products(id) ON DELETE SET NULL,
    surface_area_m2 REAL DEFAULT 0,
    message TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'new' CHECK(status IN ('new', 'contacted', 'quoted', 'closed')),
    admin_notes TEXT NOT NULL DEFAULT '',
    ip_address TEXT NOT NULL DEFAULT '',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Legacy URL Redirects (Bảo toàn SEO từ web cũ)
CREATE TABLE IF NOT EXISTS url_redirects (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    old_path TEXT NOT NULL UNIQUE, -- e.g., '/product/son-lot-go-he-nuoc/'
    new_path TEXT NOT NULL,        -- e.g., '/san-pham/son-lot-go-he-nuoc-lotus'
    status_code INTEGER NOT NULL DEFAULT 301,
    hits INTEGER NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
```

### 4.2. `002_fts_search.sql`

```sql
-- Full-Text Search virtual table for instant catalog & technical article search
CREATE VIRTUAL TABLE IF NOT EXISTS search_index USING fts5(
    entity_type, -- 'product' or 'article'
    entity_id UNINDEXED,
    title,
    slug UNINDEXED,
    thumbnail_url UNINDEXED,
    summary,
    body_text,
    tokenize = 'unicode61 remove_diacritics 2'
);
```
> **Note for Agents:** `remove_diacritics 2` in SQLite FTS5 allows Vietnamese users to search either with accents (`"sơn giả gỗ"`) or without accents (`"son gia go"`) and still match accurately! Keep `search_index` synchronized via SQLite triggers or repository save hooks.

---

## 5. Seed Data Requirements (`internal/database/seed.go`)

The application MUST boot with comprehensive, authentic seed data representing **Sơn Lotus** if the database is empty:

### 5.1. Core Categories (7 Categories)
1. **Sơn Gỗ Hệ Nước** (`son-go-he-nuoc`) — Giải pháp sơn gỗ nội ngoại thất cao cấp, đạt chuẩn xuất khẩu Châu Âu/Mỹ, không độc hại, giữ vân gỗ tự nhiên hoặc phủ màu hiện đại.
2. **Sơn Giả Gỗ** (`son-gia-go`) — Sơn tạo vân gỗ trên sắt mạ kẽm, bê tông, tấm xi măng sợi (Smartboard/Cemboard), nhôm đúc.
3. **Sơn Kim Loại Hệ Nước** (`son-kim-loai-he-nuoc`) — Sơn chống rỉ và sơn phủ trực tiếp lên sắt mạ kẽm, inox, nhôm, đồng, độ bám dính cực cao.
4. **Sơn Ngói & Bê Tông** (`son-ngoi-be-tong`) — Sơn ngói, sơn giả đá, sơn sân tennis, sàn bê tông chống thấm, kháng kiềm và chống rêu mốc.
5. **Sơn Loa (Speaker Paint)** (`son-loa`) — Sơn thùng loa chuyên dụng tạo gai sần (texture), chống trầy xước, chống thấm nước cho loa sân khấu & hội trường.
6. **Sơn Composite & Nhựa** (`son-composite-nhua`) — Sơn chuyên dụng cho vật liệu FRP, nhựa ABS, PVC, chậu kiểng composite.
7. **Trám Trét & Chống Thấm** (`tram-tret-chong-tham`) — Keo bóng nước, keo chống thấm, bột trét gỗ (Wood Filler) hệ nước.

### 5.2. Representative Seed Products (At least 8 detailed products)
1. **Sơn Phủ Gỗ Hệ Nước Lotus Acrylic Lacquer** (`son-phu-go-he-nuoc-lotus-acrylic-lacquer`)
2. **Sơn Lót Gỗ Hệ Nước Lotus Wood Primer** (`son-lot-go-he-nuoc-lotus-wood-primer`)
3. **Bột Trét Gỗ Hệ Nước Lotus Wood Filler** (`bot-tret-go-he-nuoc-lotus-wood-filler`)
4. **Sơn Giả Gỗ Lotus Trên Sắt Mạ Kẽm & Xi Măng** (`son-gia-go-lotus-sat-ma-kem-xi-mang`)
5. **Sơn Kim Loại Mạ Kẽm Hệ Nước Lotus 2-in-1** (`son-kim-loai-ma-kem-he-nuoc-lotus`)
6. **Sơn Loa Tạo Gai Sần Lotus Speaker Paint** (`son-loa-tao-gai-san-lotus-speaker-paint`)
7. **Sơn Ngói Cao Cấp Hệ Nước Lotus Roof Paint** (`son-ngoi-cao-cap-he-nuoc-lotus`)
8. **Keo Bóng Nước Chống Thấm Lotus Clear Sealer** (`keo-bong-nuoc-chong-tham-lotus-clear-sealer`)

---

## 6. Routing & Functional Specification

### 6.1. Public Website Routes
| Method | Route | Handler | Description |
| :--- | :--- | :--- | :--- |
| `GET` | `/` | `Home` | Hero banner, 7 Category cards, "Why Water-Based Paint?" eco-benefits, Featured Products, Paint Calculator Widget, Latest Projects & Tech Guides. |
| `GET` | `/san-pham` | `CatalogAll` | Full product catalog with sidebar filters (Category, Coating Role, Finish, Usage Area). Supports HTMX partial updates (`hx-target="#catalog-grid"`). |
| `GET` | `/danh-muc/{slug}` | `CategoryDetail` | Category landing page with SEO intro, application instructions, and HTMX-filterable product grid. |
| `GET` | `/san-pham/{slug}` | `ProductDetail` | Product gallery, interactive Alpine.js Color Swatch previewer, packaging badges, technical specs table, step-by-step application guide, TDS/MSDS download links, embedded Paint Coverage Calculator, and "Nhận Báo Giá" modal. |
| `GET` | `/cong-cu/tinh-luong-son` | `PaintCalculatorPage` | Dedicated interactive tool to calculate paint quantity ($L$ / $Kg$) based on Surface Area ($m^2$), Coats, and Product type. |
| `POST` | `/api/htmx/tinh-luong-son` | `CalculatePaintHTMX` | Calculates total paint needed: $$\text{Total Liters} = \frac{\text{Area } (m^2) \times \text{Coats}}{\text{Coverage Rate}} \times 1.05 \text{ (5\% wastage)}$$ and recommends optimal packaging mix (e.g., Thùng 18L + Lon 3.8L + Lon 0.8L). Returns HTML partial. |
| `GET` | `/du-an` | `ProjectsList` | Showcase of completed projects using Sơn Lotus. |
| `GET` | `/ho-tro-ky-thuat` | `TechSupportList` | Technical manuals, surface preparation guides, troubleshooting. |
| `GET` | `/giai-phap` | `SolutionsList` | Industry solutions (Nội thất xuất khẩu, Đồ chơi gỗ trẻ em chuẩn EN71, Sơn không mùi cho Resort/Bệnh viện). |
| `GET` | `/bai-viet/{slug}` | `ArticleDetail` | Full article/project view with Table of Contents, related products sidebar, and `Article` JSON-LD. |
| `GET` | `/gioi-thieu` | `AboutPage` | Company profile (Công ty TNHH Bích Trang), R&D technology, environmental safety standards (Low VOC, non-toxic, non-flammable). |
| `GET` | `/lien-he` | `ContactPage` | Factory/Office address, Google Maps embed, hotline, dealer registration & inquiry form. |
| `POST` | `/lien-he` | `SubmitInquiry` | Validates form (honeypot field `website_url` must be empty), saves to `inquiries`, returns HTMX confirmation banner. |
| `GET` | `/api/htmx/tim-kiem` | `LiveSearchHTMX` | Queries `search_index` via FTS5 for `?q=...` and returns dropdown results partial (Products + Tech Articles). |
| `GET` | `/sitemap.xml` | `SitemapXML` | Dynamically renders standard XML sitemap of all active pages, categories, products, and articles. |
| `GET` | `/robots.txt` | `RobotsTxt` | Standard robots file pointing to `/sitemap.xml` and disallowing `/admin`. |

### 6.2. Admin Portal Routes (`/admin/*`)
Protected by session cookie authentication. Default admin credentials loaded from environment variables (`ADMIN_USER`, `ADMIN_PASSWORD`).
- `GET /admin/login` & `POST /admin/login` & `POST /admin/logout`
- `GET /admin` — Dashboard showing total products, articles, unread customer inquiries, and top 301 redirects.
- `GET /admin/products`, `GET /admin/products/new`, `POST /admin/products`, `GET /admin/products/{id}/edit`, `POST /admin/products/{id}` — Full CRUD for products.
- `GET /admin/articles`, `GET /admin/articles/new`, `POST /admin/articles`, `GET /admin/articles/{id}/edit`, `POST /admin/articles/{id}` — Full CRUD for articles/projects.
- `GET /admin/inquiries`, `POST /admin/inquiries/{id}/status` — Manage customer leads & update status (`new` -> `contacted` -> `quoted` -> `closed`).
- `POST /admin/upload` — Multipart file upload endpoint saving images/PDFs to `./data/uploads/`. Served publicly at `/uploads/*`.

---

## 7. Legacy Site Migrator Tool (`cmd/migrator/main.go`)

Build a standalone CLI utility that allows operators or AI Agents to import content directly from `https://sonlotus.com`:
- **Flags:**
  - `--seed-only`: Runs default database migrations and inserts built-in seed data.
  - `--scrape-sitemap`: Fetches `https://sonlotus.com/sitemap_index.xml` (or `/sitemap.xml`), parses product and post URLs, extracts Title, Meta Description, H1, Main Content HTML, Featured Images, and populates `products`, `articles`, and `url_redirects` tables.
  - `--dry-run`: Logs discovered URLs and parsed structs without modifying the database.
- **Polite Scraping:** Respect rate limits (max 2 requests/second, 10s timeout, custom User-Agent `SonLotus-Migrator/1.0`).

---

## 8. Non-Functional Requirements & UI/UX Details

1. **Floating Quick-Action Bar (Mobile & Desktop):**
   - Include a sticky bottom-right widget for instant conversion: **Gọi Hotline**, **Chat Zalo Tư Vấn Kỹ Thuật**, and **Tính Lượng Sơn**.
2. **SEO & Structured Data (`internal/service/seo_service.go`):**
   - Every public page must pass a `PageSEO` struct into `base.html` containing `Title`, `Description`, `CanonicalURL`, `OGImage`, `OGType`, and `JSONLD` (`template.JS`).
   - Product pages must generate valid `schema.org/Product` JSON-LD.
   - Homepage must generate `schema.org/Organization` and `schema.org/LocalBusiness` JSON-LD.
3. **Security:**
   - Parameterized SQL queries everywhere.
   - Honeypot anti-spam field + IP rate-limiting on `POST /lien-he`.
   - `HttpOnly`, `SameSite=Lax` cookies for Admin sessions.

---

## 9. Agent Execution Plan & Verification Protocol

Agents must implement the project in **4 sequential phases**. At the end of each phase, run the verification command before proceeding.

### Phase 1: Project Initialization, Database, FTS5 & Seed Data
- Initialize `go.mod` (`module sonlotus`), `Makefile`, `.env.example`, and directory tree.
- Implement `internal/config`, `internal/domain`, `internal/database` (migrations + SQLite WAL + FTS5 synchronization + comprehensive Sơn Lotus seed data).
- Implement `internal/repository` for Categories, Products, Articles, Inquiries, and Redirects.
- **Verification Gate 1:**
  - Write unit tests in `internal/repository/catalog_repo_test.go` testing product creation, category filtering, Paint Calculator math, and Vietnamese FTS5 search without diacritics (e.g., querying `"son gia go"` returns `"Sơn Giả Gỗ Lotus"`).
  - Run `go test -v ./internal/...` — MUST PASS 100%.

### Phase 2: Core Services, Public HTTP Handlers & Templates
- Implement `internal/service` (CatalogService, PaintCalculator logic, SEOService).
- Create `web/embed.go` and build responsive TailwindCSS templates in `web/templates/` (`base.html`, `home.html`, `category.html`, `product_detail.html`, `calculator.html`, `articles_list.html`, `article_detail.html`, `contact.html`, `about.html`, and HTMX partials).
- Implement public & content HTTP handlers in `internal/handler/`.
- **Verification Gate 2:**
  - Write HTTP integration tests in `internal/handler/public_handler_test.go` verifying HTTP 200 responses and valid HTML rendering for `/`, `/san-pham`, `/danh-muc/son-go-he-nuoc`, `/san-pham/son-gia-go-lotus-sat-ma-kem-xi-mang`, `/sitemap.xml`, and `POST /api/htmx/tinh-luong-son`.
  - Run `go test -v ./...` — MUST PASS 100%.

### Phase 3: Admin CMS, Lead Management & Legacy Migrator CLI
- Implement Admin session middleware and `/admin/*` CRUD handlers + templates.
- Implement `cmd/migrator/main.go` with `--seed-only` and `--scrape-sitemap` modes.
- Implement `cmd/server/main.go` wiring all routes, middleware (logging, recovery, 301 redirects), and graceful shutdown.
- **Verification Gate 3:**
  - Run `go build -o bin/server ./cmd/server` and `go build -o bin/migrator ./cmd/migrator`. Both binaries must compile cleanly with zero warnings.

### Phase 4: Production Packaging & Final Audit
- Create multi-stage `Dockerfile` and comprehensive `Makefile` (`make dev`, `make build`, `make test`, `make seed`).
- Run `go vet ./...` and `go test -race ./...` to guarantee thread-safety and zero data races.