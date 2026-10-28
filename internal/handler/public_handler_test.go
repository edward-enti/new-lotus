package handler_test

import (
	"database/sql"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"strings"
	"testing"

	"sonlotus/internal/database"
	"sonlotus/internal/handler"
	"sonlotus/internal/repository"
	"sonlotus/internal/service"
)

// newTestMux builds a fully wired public mux over a fresh seeded DB.
func newTestMux(t *testing.T) (*http.ServeMux, *sql.DB) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "test.db")
	db, err := database.Open(path)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	if err := database.Migrate(db); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	if err := database.Seed(db); err != nil {
		t.Fatalf("seed: %v", err)
	}
	cats := &repository.CategoryRepo{DB: db}
	prods := &repository.ProductRepo{DB: db}
	arts := &repository.ArticleRepo{DB: db}
	inqs := &repository.InquiryRepo{DB: db}
	reds := &repository.RedirectRepo{DB: db}
	rend, err := handler.NewRenderer()
	if err != nil {
		t.Fatalf("renderer: %v", err)
	}
	deps := &handler.Deps{
		Categories: cats, Products: prods, Articles: arts, Inquiries: inqs, Redirects: reds,
		Catalog: &service.CatalogService{Categories: cats, Products: prods},
		Content: &service.ContentService{Articles: arts, Products: prods},
		SEO:     &service.SEOService{},
		Render:  rend,
	}
	return handler.NewMux(deps), db
}

func get(t *testing.T, mux *http.ServeMux, path string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	return rec
}

func assertHTML(t *testing.T, rec *httptest.ResponseRecorder, marker string) {
	t.Helper()
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200, body: %s", rec.Code, rec.Body.String())
	}
	ct := rec.Header().Get("Content-Type")
	if !strings.HasPrefix(ct, "text/html") {
		t.Fatalf("Content-Type = %q, want text/html", ct)
	}
	if !strings.Contains(rec.Body.String(), marker) {
		t.Fatalf("body does not contain marker %q", marker)
	}
}

func TestHome(t *testing.T) {
	mux, _ := newTestMux(t)
	assertHTML(t, get(t, mux, "/"), "Sơn Gỗ Hệ Nước")
}

func TestCatalog(t *testing.T) {
	mux, _ := newTestMux(t)
	assertHTML(t, get(t, mux, "/san-pham"), "Sơn Giả Gỗ Lotus")
}

func TestCategory(t *testing.T) {
	mux, _ := newTestMux(t)
	assertHTML(t, get(t, mux, "/danh-muc/son-go-he-nuoc"), "Sơn Phủ Gỗ Hệ Nước Lotus Acrylic Lacquer")
}

func TestProductDetail(t *testing.T) {
	mux, _ := newTestMux(t)
	rec := get(t, mux, "/san-pham/son-gia-go-lotus-sat-ma-kem-xi-mang")
	assertHTML(t, rec, "Sơn Giả Gỗ Lotus")
	if !strings.Contains(rec.Body.String(), `"@type":"Product"`) {
		t.Fatal("product page body does not contain Product JSON-LD")
	}
}

func TestCalculatorPage(t *testing.T) {
	mux, _ := newTestMux(t)
	assertHTML(t, get(t, mux, "/cong-cu/tinh-luong-son"), "Tính lượng sơn")
}

func TestCalculatePaintHTMX(t *testing.T) {
	mux, _ := newTestMux(t)
	form := url.Values{"dien-tich": {"100"}, "so-lop": {"2"}, "do-phu": {"10"}}
	req := httptest.NewRequest(http.MethodPost, "/api/htmx/tinh-luong-son", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("HX-Request", "true")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	assertHTML(t, rec, "21.00")
}

func TestSitemap(t *testing.T) {
	mux, _ := newTestMux(t)
	rec := get(t, mux, "/sitemap.xml")
	if rec.Code != http.StatusOK {
		t.Fatalf("sitemap status = %d, want 200", rec.Code)
	}
	// Sitemap is XML, not HTML: asserting text/html here would be a bug.
	// (PLAN Gate 2's "valid HTML" definition applies to HTML pages; the
	// sitemap entry is verified by status + XML content-type + marker.)
	ct := rec.Header().Get("Content-Type")
	if !strings.HasPrefix(ct, "application/xml") {
		t.Fatalf("sitemap Content-Type = %q, want application/xml", ct)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "<urlset") || !strings.Contains(body, "/san-pham/son-gia-go-lotus-sat-ma-kem-xi-mang") {
		t.Fatalf("sitemap body missing urlset/product URL: %s", body)
	}
}

func TestRobots(t *testing.T) {
	mux, _ := newTestMux(t)
	rec := get(t, mux, "/robots.txt")
	if rec.Code != http.StatusOK {
		t.Fatalf("robots status = %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "Disallow: /admin") || !strings.Contains(rec.Body.String(), "Sitemap: /sitemap.xml") {
		t.Fatalf("robots body = %q", rec.Body.String())
	}
}

func TestLiveSearchHTMX(t *testing.T) {
	mux, _ := newTestMux(t)
	rec := get(t, mux, "/api/htmx/tim-kiem?q=son+gia+go")
	assertHTML(t, rec, "Sơn Giả Gỗ Lotus")
}

func TestInactiveProduct404AndSitemapFiltering(t *testing.T) {
	mux, db := newTestMux(t)
	if _, err := db.Exec(`UPDATE products SET is_active = 0 WHERE slug = 'son-loa-tao-gai-san-lotus-speaker-paint'`); err != nil {
		t.Fatal(err)
	}
	rec := get(t, mux, "/san-pham/son-loa-tao-gai-san-lotus-speaker-paint")
	if rec.Code != http.StatusNotFound {
		t.Fatalf("inactive product status = %d, want 404", rec.Code)
	}
	sitemap := get(t, mux, "/sitemap.xml")
	body, _ := io.ReadAll(sitemap.Body)
	if strings.Contains(string(body), "son-loa-tao-gai-san-lotus-speaker-paint") {
		t.Fatal("sitemap contains inactive product slug")
	}
}

func TestMissingArticle404(t *testing.T) {
	mux, _ := newTestMux(t)
	rec := get(t, mux, "/bai-viet/khong-ton-tai")
	if rec.Code != http.StatusNotFound {
		t.Fatalf("missing article status = %d, want 404", rec.Code)
	}
}
