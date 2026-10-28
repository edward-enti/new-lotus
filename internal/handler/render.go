// Package handler serves public HTTP routes. HTML policy (PHASE2-PINS):
// escape by default — html/template auto-escapes, and database HTML fields
// are never rendered as template.HTML anywhere in this package.
package handler

import (
	"html/template"
	"net/http"

	"sonlotus/internal/domain"
	"sonlotus/internal/service"
	"sonlotus/web"
)

// ViewData is the typed wrapper every template receives (SPEC §2).
type ViewData struct {
	SEO  domain.PageSEO
	Data any
}

// Renderer parses the embedded templates. Each page gets its own template
// set (base + that one page + partials) because every page defines a
// "content" block: parsing all pages into one set would let one page's
// "content" definition silently replace the others.
type Renderer struct {
	pages    map[string]*template.Template
	partials *template.Template
}

// pageFiles maps a render page name to its template file.
var pageFiles = map[string]string{
	"home":           "templates/pages/home.html",
	"category":       "templates/pages/category.html",
	"product_detail": "templates/pages/product_detail.html",
	"articles_list":  "templates/pages/articles_list.html",
	"article_detail": "templates/pages/article_detail.html",
	"calculator":     "templates/pages/calculator.html",
	"contact":        "templates/pages/contact.html",
	"about":          "templates/pages/about.html",
}

// NewRenderer parses layouts, pages and partials from the embedded FS.
// webpURL is the <picture> WebP fallback derivation (media_service):
// templates use it for image URLs whose WebP sibling is supplied.
func NewRenderer() (*Renderer, error) {
	funcs := template.FuncMap{"webpURL": service.WebPURL}
	partials, err := template.New("").Funcs(funcs).ParseFS(web.FS, "templates/partials/*.html")
	if err != nil {
		return nil, err
	}
	pages := map[string]*template.Template{}
	for name, file := range pageFiles {
		tmpl, err := template.New("").Funcs(funcs).ParseFS(web.FS,
			"templates/layouts/base.html", file, "templates/partials/*.html")
		if err != nil {
			return nil, err
		}
		pages[name] = tmpl
	}
	return &Renderer{pages: pages, partials: partials}, nil
}

// Render writes a full page through base.html.
func (r *Renderer) Render(w http.ResponseWriter, page string, seo domain.PageSEO, data any) {
	tmpl, ok := r.pages[page]
	if !ok {
		http.Error(w, "unknown page", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.ExecuteTemplate(w, "base", ViewData{SEO: seo, Data: data}); err != nil {
		http.Error(w, "render error", http.StatusInternalServerError)
	}
}

// RenderPartial writes one named partial template (HTMX responses).
func (r *Renderer) RenderPartial(w http.ResponseWriter, name string, data any) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := r.partials.ExecuteTemplate(w, name, data); err != nil {
		http.Error(w, "render error", http.StatusInternalServerError)
	}
}

// isHTMX reports whether the request came from HTMX.
func isHTMX(r *http.Request) bool {
	return r.Header.Get("HX-Request") == "true"
}

// notFound writes a plain 404. Public 404s must not leak whether a slug
// exists as an inactive record: inactive and missing are identical here.
func notFound(w http.ResponseWriter) {
	http.NotFound(w, nil)
}
