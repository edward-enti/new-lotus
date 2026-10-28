-- 001_initial_schema.sql — SPEC §4.1 verbatim.
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
    msds_file_url TEXT NOT NULL DEFAULT '',        -- Safety Data Sheet PDF
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
