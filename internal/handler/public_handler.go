package handler

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"

	"sonlotus/internal/domain"
	"sonlotus/internal/repository"
)

// HomeData is the data for home.html.
type HomeData struct {
	Categories []domain.Category
	Featured   []domain.Product
	Articles   []domain.Article
}

// Home renders the landing page (SPEC §6.1 sections).
func (h *Handlers) Home(w http.ResponseWriter, r *http.Request) {
	cats, err := h.Catalog.ActiveCategories()
	if err != nil {
		http.Error(w, "catalog error", http.StatusInternalServerError)
		return
	}
	featured, err := h.Catalog.FeaturedProducts()
	if err != nil {
		http.Error(w, "catalog error", http.StatusInternalServerError)
		return
	}
	articles, err := h.Content.LatestArticles(6)
	if err != nil {
		http.Error(w, "content error", http.StatusInternalServerError)
		return
	}
	h.Render.Render(w, "home", h.SEO.HomeSEO(), HomeData{Categories: cats, Featured: featured, Articles: articles})
}

// CatalogData is the data for category.html / catalog listings.
type CatalogData struct {
	Categories []domain.Category
	Products   []domain.Product
	Filter     repository.ProductFilter
	Title      string
}

// CatalogAll renders the full catalog with sidebar filters; HTMX
// requests receive only the product_grid partial.
func (h *Handlers) CatalogAll(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	filter := repository.ProductFilter{
		CategorySlug: q.Get("danh-muc"),
		CoatingRole:  q.Get("vai-tro"),
		FinishType:   q.Get("do-bong"),
		UsageArea:    q.Get("khu-vuc"),
	}
	prods, err := h.Catalog.ProductsByFilter(filter)
	if err != nil {
		http.Error(w, "catalog error", http.StatusInternalServerError)
		return
	}
	if isHTMX(r) {
		h.Render.RenderPartial(w, "product_grid", prods)
		return
	}
	cats, err := h.Catalog.ActiveCategories()
	if err != nil {
		http.Error(w, "catalog error", http.StatusInternalServerError)
		return
	}
	seo := h.SEO.HomeSEO()
	seo.Title = "Tất cả sản phẩm — Sơn Lotus"
	seo.Description = "Danh mục sơn hệ nước Sơn Lotus: sơn gỗ, sơn giả gỗ, sơn kim loại, sơn ngói, sơn loa, chống thấm."
	seo.CanonicalURL = h.SEO.Canonical("/san-pham")
	h.Render.Render(w, "category", seo, CatalogData{Categories: cats, Products: prods, Filter: filter, Title: "Tất cả sản phẩm"})
}

// CategoryDetail renders a category landing page with its product grid.
func (h *Handlers) CategoryDetail(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	cat, err := h.Categories.GetBySlugActive(slug)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			notFound(w)
			return
		}
		http.Error(w, "catalog error", http.StatusInternalServerError)
		return
	}
	prods, err := h.Catalog.ProductsByCategory(slug)
	if err != nil {
		http.Error(w, "catalog error", http.StatusInternalServerError)
		return
	}
	if isHTMX(r) {
		h.Render.RenderPartial(w, "product_grid", prods)
		return
	}
	cats, err := h.Catalog.ActiveCategories()
	if err != nil {
		http.Error(w, "catalog error", http.StatusInternalServerError)
		return
	}
	h.Render.Render(w, "category", h.SEO.CategorySEO(cat), CatalogData{Categories: cats, Products: prods, Filter: repository.ProductFilter{CategorySlug: slug}, Title: cat.Name})
}

// ProductData is the data for product_detail.html.
type ProductData struct {
	Product domain.Product
	Related []domain.Product
}

// ProductDetail renders a product page; inactive slugs are 404.
func (h *Handlers) ProductDetail(w http.ResponseWriter, r *http.Request) {
	p, err := h.Catalog.ProductBySlug(r.PathValue("slug"))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			notFound(w)
			return
		}
		http.Error(w, "catalog error", http.StatusInternalServerError)
		return
	}
	related, err := h.Catalog.RelatedProducts(p)
	if err != nil {
		http.Error(w, "catalog error", http.StatusInternalServerError)
		return
	}
	h.Render.Render(w, "product_detail", h.SEO.ProductSEO(p), ProductData{Product: p, Related: related})
}

// CalculatorData is the data for calculator.html and calculator_result.html.
type CalculatorData struct {
	AreaM2     float64
	Coats      int
	Coverage   float64
	Liters     float64
	Mix        domain.PackagingMix
	Calculated bool
	Error      string
	Products   []domain.Product
}

// CalculatorPage renders the dedicated calculator tool.
func (h *Handlers) CalculatorPage(w http.ResponseWriter, r *http.Request) {
	prods, err := h.Products.ListActive()
	if err != nil {
		http.Error(w, "catalog error", http.StatusInternalServerError)
		return
	}
	seo := h.SEO.HomeSEO()
	seo.Title = "Tính lượng sơn — Sơn Lotus"
	seo.Description = "Tính lượng sơn cần dùng theo diện tích, số lớp sơn và độ phủ của sản phẩm."
	seo.CanonicalURL = h.SEO.Canonical("/cong-cu/tinh-luong-son")
	h.Render.Render(w, "calculator", seo, CalculatorData{Products: prods})
}

// CalculatePaintHTMX handles POST /api/htmx/tinh-luong-son and returns
// the calculator_result partial (SPEC §6.1 formula).
func (h *Handlers) CalculatePaintHTMX(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		h.Render.RenderPartial(w, "calculator_result", CalculatorData{Error: "Dữ liệu gửi lên không hợp lệ."})
		return
	}
	area, _ := strconv.ParseFloat(r.FormValue("dien-tich"), 64)
	coats, _ := strconv.Atoi(r.FormValue("so-lop"))
	coverage, _ := strconv.ParseFloat(r.FormValue("do-phu"), 64)
	if coverage == 0 {
		// Allow choosing a product whose coverage rate fills the field.
		if slug := r.FormValue("san-pham"); slug != "" {
			if p, err := h.Catalog.ProductBySlug(slug); err == nil {
				coverage = p.CoverageRate
			}
		}
	}
	liters, mix, err := h.Catalog.Calculate(area, coats, coverage)
	if err != nil {
		h.Render.RenderPartial(w, "calculator_result", CalculatorData{AreaM2: area, Coats: coats, Coverage: coverage, Error: "Vui lòng nhập diện tích, số lớp và độ phủ hợp lệ (diện tích không âm, độ phủ lớn hơn 0)."})
		return
	}
	h.Render.RenderPartial(w, "calculator_result", CalculatorData{AreaM2: area, Coats: coats, Coverage: coverage, Liters: liters, Mix: mix, Calculated: true})
}

// AboutPage renders the company profile page.
func (h *Handlers) AboutPage(w http.ResponseWriter, r *http.Request) {
	seo := h.SEO.HomeSEO()
	seo.Title = "Giới thiệu — Sơn Lotus"
	seo.Description = "Sơn Lotus thuộc Công ty TNHH Bích Trang: sơn hệ nước Low VOC, không độc hại, không cháy, đạt chuẩn xuất khẩu."
	seo.CanonicalURL = h.SEO.Canonical("/gioi-thieu")
	h.Render.Render(w, "about", seo, nil)
}
