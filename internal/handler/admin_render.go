package handler

import (
	"encoding/json"
	"html/template"
	"net/http"

	"sonlotus/web"
)

// toJSON marshals a value for pre-filling a JSON textarea in an edit
// form. Without it, editing a product would silently wipe its
// features/packaging/specs/colors/images (empty textareas parse as
// empty values on save).
func toJSON(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		return ""
	}
	return string(b)
}

// AdminView is the typed wrapper every admin template receives.
// CSRFToken is the synchronizer token for the current session (or the
// login form's pre-auth token); layout and forms both use .CSRFToken,
// so content templates must use $.CSRFToken inside range blocks.
type AdminView struct {
	CSRFToken string
	ShowNav   bool
	Data      any
}

// AdminRenderer parses the admin layout and admin pages from the
// embedded FS. Like the public renderer, each page gets its own
// template set because every page defines a "content" block.
type AdminRenderer struct {
	pages map[string]*template.Template
}

var adminPageFiles = map[string]string{
	"login":          "templates/admin/login.html",
	"dashboard":      "templates/admin/dashboard.html",
	"products_list":  "templates/admin/products_list.html",
	"product_form":   "templates/admin/product_form.html",
	"articles_list":  "templates/admin/articles_list.html",
	"article_form":   "templates/admin/article_form.html",
	"inquiries_list": "templates/admin/inquiries_list.html",
}

// NewAdminRenderer parses layouts/admin.html, admin pages and partials.
func NewAdminRenderer() (*AdminRenderer, error) {
	funcs := template.FuncMap{"toJSON": toJSON}
	pages := map[string]*template.Template{}
	for name, file := range adminPageFiles {
		tmpl, err := template.New("").Funcs(funcs).ParseFS(web.FS,
			"templates/layouts/admin.html", file, "templates/partials/*.html")
		if err != nil {
			return nil, err
		}
		pages[name] = tmpl
	}
	return &AdminRenderer{pages: pages}, nil
}

// RenderAdmin writes a full admin page through admin.html.
func (r *AdminRenderer) RenderAdmin(w http.ResponseWriter, page, csrfToken string, showNav bool, data any) {
	tmpl, ok := r.pages[page]
	if !ok {
		http.Error(w, "unknown page", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.ExecuteTemplate(w, "admin", AdminView{CSRFToken: csrfToken, ShowNav: showNav, Data: data}); err != nil {
		http.Error(w, "render error", http.StatusInternalServerError)
	}
}
