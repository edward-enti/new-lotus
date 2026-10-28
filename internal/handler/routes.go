package handler

import (
	"io/fs"
	"net/http"

	"sonlotus/internal/repository"
	"sonlotus/internal/service"
	"sonlotus/web"
)

func webStaticFS() fs.FS {
	sub, err := fs.Sub(web.FS, "static")
	if err != nil {
		panic(err)
	}
	return sub
}

// Deps wires repositories, services and the renderer for all handlers.
// Admin routes are NOT wired here: they arrive with Phase 3.
type Deps struct {
	Categories *repository.CategoryRepo
	Products   *repository.ProductRepo
	Articles   *repository.ArticleRepo
	Inquiries  *repository.InquiryRepo
	Redirects  *repository.RedirectRepo
	Catalog    *service.CatalogService
	Content    *service.ContentService
	SEO        *service.SEOService
	Render     *Renderer
}

// NewMux returns the public router (Go stdlib ServeMux, PHASE2-PINS).
// Static assets are served from the embedded FS under /static/.
func NewMux(d *Deps) *http.ServeMux {
	h := &Handlers{Deps: d}
	mux := http.NewServeMux()
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServerFS(webStaticFS())))
	mux.HandleFunc("GET /{$}", h.Home)
	mux.HandleFunc("GET /san-pham", h.CatalogAll)
	mux.HandleFunc("GET /danh-muc/{slug}", h.CategoryDetail)
	mux.HandleFunc("GET /san-pham/{slug}", h.ProductDetail)
	mux.HandleFunc("GET /cong-cu/tinh-luong-son", h.CalculatorPage)
	mux.HandleFunc("POST /api/htmx/tinh-luong-son", h.CalculatePaintHTMX)
	mux.HandleFunc("GET /du-an", h.ProjectsList)
	mux.HandleFunc("GET /ho-tro-ky-thuat", h.TechSupportList)
	mux.HandleFunc("GET /giai-phap", h.SolutionsList)
	mux.HandleFunc("GET /bai-viet/{slug}", h.ArticleDetail)
	mux.HandleFunc("GET /gioi-thieu", h.AboutPage)
	mux.HandleFunc("GET /lien-he", h.ContactPage)
	mux.HandleFunc("POST /lien-he", h.SubmitInquiry)
	mux.HandleFunc("GET /api/htmx/tim-kiem", h.LiveSearchHTMX)
	mux.HandleFunc("GET /sitemap.xml", h.SitemapXML)
	mux.HandleFunc("GET /robots.txt", h.RobotsTxt)
	return mux
}

// Handlers groups the public handlers sharing Deps.
type Handlers struct {
	*Deps
}

var realRouteExact = map[string]bool{
	"/": true, "/san-pham": true, "/cong-cu/tinh-luong-son": true,
	"/du-an": true, "/ho-tro-ky-thuat": true, "/giai-phap": true,
	"/gioi-thieu": true, "/lien-he": true, "/sitemap.xml": true, "/robots.txt": true,
	"/admin": true, "/admin/login": true, "/admin/products": true, "/admin/products/new": true,
	"/admin/articles": true, "/admin/articles/new": true, "/admin/inquiries": true,
}

var realRoutePrefixes = []string{
	"/danh-muc/", "/san-pham/", "/bai-viet/", "/api/", "/static/", "/uploads/",
	"/admin/products/", "/admin/articles/", "/admin/inquiries/",
}

// IsRealRoute reports whether path is a real application route, for
// the redirect middleware's "new routes win" rule. It mirrors the
// patterns registered by NewMux and RegisterAdmin; if a route is added
// to either, it must be added here too.
func IsRealRoute(path string) bool {
	if realRouteExact[path] {
		return true
	}
	for _, p := range realRoutePrefixes {
		if len(path) > len(p) && path[:len(p)] == p {
			return true
		}
	}
	return false
}
