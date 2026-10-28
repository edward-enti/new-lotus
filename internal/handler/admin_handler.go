package handler

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"sonlotus/internal/domain"
	"sonlotus/internal/middleware"
	"sonlotus/internal/repository"
)

// AdminDeps wires everything the admin handlers need. UploadDir is the
// on-disk uploads directory (under DATA_DIR, git-ignored); Production
// controls the Secure cookie attribute (APP_ENV=production).
type AdminDeps struct {
	Categories *repository.CategoryRepo
	Products   *repository.ProductRepo
	Articles   *repository.ArticleRepo
	Inquiries  *repository.InquiryRepo
	Redirects  *repository.RedirectRepo
	Auth       *middleware.AdminAuth
	CSRF       *middleware.CSRFStore
	Render     *AdminRenderer
	UploadDir  string
	Production bool
}

// AdminHandlers groups the admin handlers sharing AdminDeps.
type AdminHandlers struct {
	*AdminDeps
}

// RegisterAdmin registers all /admin/* routes (and the public
// /uploads/ serving route, which needs the same UploadDir) on mux.
// Authentication, CSRF and the rest of the fixed chain are applied by
// the server's middleware wrapper, not here, so tests can also wrap
// this mux with middleware.Chain.
func RegisterAdmin(mux *http.ServeMux, d *AdminDeps) {
	h := &AdminHandlers{AdminDeps: d}
	mux.HandleFunc("GET /admin", h.Dashboard)
	mux.HandleFunc("GET /admin/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/admin", http.StatusMovedPermanently)
	})
	mux.HandleFunc("GET /admin/login", h.LoginPage)
	mux.HandleFunc("POST /admin/login", h.LoginSubmit)
	mux.HandleFunc("POST /admin/logout", h.Logout)
	mux.HandleFunc("GET /admin/products", h.ProductsList)
	mux.HandleFunc("GET /admin/products/new", h.ProductNew)
	mux.HandleFunc("POST /admin/products", h.ProductCreate)
	mux.HandleFunc("GET /admin/products/{id}/edit", h.ProductEdit)
	mux.HandleFunc("POST /admin/products/{id}", h.ProductUpdate)
	mux.HandleFunc("GET /admin/articles", h.ArticlesList)
	mux.HandleFunc("GET /admin/articles/new", h.ArticleNew)
	mux.HandleFunc("POST /admin/articles", h.ArticleCreate)
	mux.HandleFunc("GET /admin/articles/{id}/edit", h.ArticleEdit)
	mux.HandleFunc("POST /admin/articles/{id}", h.ArticleUpdate)
	mux.HandleFunc("GET /admin/inquiries", h.InquiriesList)
	mux.HandleFunc("POST /admin/inquiries/{id}/status", h.InquiryStatus)
	mux.HandleFunc("POST /admin/upload", h.Upload)
	mux.HandleFunc("GET /uploads/{file}", h.UploadsServe)
}

// sessionCSRF returns the synchronizer token for the current admin
// session (from context, set by RequireAdmin), generating it on first use.
func (h *AdminHandlers) sessionCSRF(r *http.Request) string {
	hash, ok := middleware.SessionHashFromContext(r.Context())
	if !ok {
		return ""
	}
	tok, err := h.CSRF.TokenFor(middleware.SessionKey(hash))
	if err != nil {
		return ""
	}
	return tok
}

// LoginData is the data for login.html.
type LoginData struct {
	Error string
}

// LoginPage renders the admin login form with a pre-auth synchronizer
// token (login has no session yet, so the token is keyed by a
// login-form id cookie, stored server-side in the CSRFStore).
func (h *AdminHandlers) LoginPage(w http.ResponseWriter, r *http.Request) {
	id, err := middleware.EnsureLoginID(w, r)
	if err != nil {
		http.Error(w, "csrf error", http.StatusInternalServerError)
		return
	}
	tok, err := h.CSRF.TokenFor(middleware.LoginKey(id))
	if err != nil {
		http.Error(w, "csrf error", http.StatusInternalServerError)
		return
	}
	h.Render.RenderAdmin(w, "login", tok, false, LoginData{})
}

func (h *AdminHandlers) loginFail(w http.ResponseWriter, r *http.Request, status int, msg string) {
	id, _ := middleware.EnsureLoginID(w, r)
	tok, _ := h.CSRF.TokenFor(middleware.LoginKey(id))
	w.WriteHeader(status)
	h.Render.RenderAdmin(w, "login", tok, false, LoginData{Error: msg})
}

// LoginSubmit authenticates an admin. Failure counting is
// failure-only: a blocked IP gets 429 before any password check, and
// only a failed attempt is recorded. The error is generic for both an
// unknown user and a wrong password (no user-vs-password oracle).
func (h *AdminHandlers) LoginSubmit(w http.ResponseWriter, r *http.Request) {
	// Defence in depth: the chain's CSRF step already validated this
	// token, but a directly-wired mux (tests, future callers) gets the
	// same check here.
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	if !h.CSRF.Validate(middleware.LoginKey(middleware.LoginIDFromRequest(r)), r.FormValue(middleware.CSRFField)) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	ip := middleware.ClientIP(r)
	if h.Auth.LoginLimiter.Blocked(ip) {
		h.loginFail(w, r, http.StatusTooManyRequests, "Quá nhiều lần đăng nhập thất bại, vui lòng thử lại sau 15 phút.")
		return
	}
	user := strings.TrimSpace(r.FormValue("username"))
	password := r.FormValue("password")
	// Timing parity: ALWAYS run exactly one bcrypt comparison before
	// returning the generic error — the real hash for a known user,
	// the dummy hash for an unknown user — so timing cannot leak
	// which case failed. No timing test is added (it would be flaky);
	// the guarantee is structural, plus the generic-error test.
	userOK := h.Auth.IsUser(user)
	var passOK bool
	if userOK {
		passOK = h.Auth.CheckPassword(password)
	} else {
		passOK = h.Auth.CheckDummy(password)
	}
	if !userOK || !passOK {
		h.Auth.LoginLimiter.Record(ip)
		// Generic failure only — the same message for unknown user and
		// wrong password (approved security note).
		h.loginFail(w, r, http.StatusUnauthorized, "Tên đăng nhập hoặc mật khẩu không đúng")
		return
	}
	// Rotation on login: an existing session cookie's row (and its
	// CSRF token) is deleted before the new session is created.
	if c, err := r.Cookie(middleware.SessionCookie); err == nil && c.Value != "" {
		_ = h.Auth.InvalidateSession(c.Value)
		h.CSRF.Delete(middleware.SessionKey(repository.HashToken(c.Value)))
	}
	token, err := h.Auth.NewSession()
	if err != nil {
		http.Error(w, "session error", http.StatusInternalServerError)
		return
	}
	middleware.SetSessionCookie(w, token, h.Production)
	http.Redirect(w, r, "/admin", http.StatusFound)
}

// Logout invalidates the session (row deleted) and its CSRF token,
// clears the cookie, and redirects to login. CSRF for this POST is
// validated by the chain in production; this handler itself does not
// validate a CSRF token, it only deletes the session and cookie.
func (h *AdminHandlers) Logout(w http.ResponseWriter, r *http.Request) {
	if c, err := r.Cookie(middleware.SessionCookie); err == nil && c.Value != "" {
		_ = h.Auth.InvalidateSession(c.Value)
		h.CSRF.Delete(middleware.SessionKey(repository.HashToken(c.Value)))
	}
	middleware.ClearSessionCookie(w, h.Production)
	http.Redirect(w, r, "/admin/login", http.StatusFound)
}

// DashboardData is the data for dashboard.html.
type DashboardData struct {
	ProductCount int
	ArticleCount int
	UnreadCount  int
	TopRedirects []repository.Redirect
}

// Dashboard renders totals, unread inquiries and top redirects.
func (h *AdminHandlers) Dashboard(w http.ResponseWriter, r *http.Request) {
	prods, err := h.Products.ListAll()
	if err != nil {
		http.Error(w, "catalog error", http.StatusInternalServerError)
		return
	}
	arts, err := h.Articles.ListAll()
	if err != nil {
		http.Error(w, "content error", http.StatusInternalServerError)
		return
	}
	unread, err := h.Inquiries.CountByStatus("new")
	if err != nil {
		http.Error(w, "inquiry error", http.StatusInternalServerError)
		return
	}
	top, err := h.Redirects.Top(10)
	if err != nil {
		http.Error(w, "redirect error", http.StatusInternalServerError)
		return
	}
	h.Render.RenderAdmin(w, "dashboard", h.sessionCSRF(r), true, DashboardData{
		ProductCount: len(prods), ArticleCount: len(arts), UnreadCount: unread, TopRedirects: top,
	})
}

// ProductsListData is the data for products_list.html.
type ProductsListData struct {
	Products []domain.Product
}

// ProductsList renders all products (active and inactive) for admin.
func (h *AdminHandlers) ProductsList(w http.ResponseWriter, r *http.Request) {
	prods, err := h.Products.ListAll()
	if err != nil {
		http.Error(w, "catalog error", http.StatusInternalServerError)
		return
	}
	h.Render.RenderAdmin(w, "products_list", h.sessionCSRF(r), true, ProductsListData{Products: prods})
}

// ProductFormData is the data for product_form.html.
type ProductFormData struct {
	Product    domain.Product
	Categories []domain.Category
	Error      string
	IsNew      bool
	Action     string
}

func (h *AdminHandlers) productForm(w http.ResponseWriter, r *http.Request, p domain.Product, isNew bool, action, errMsg string) {
	cats, err := h.Categories.ListAll()
	if err != nil {
		http.Error(w, "catalog error", http.StatusInternalServerError)
		return
	}
	if errMsg != "" {
		w.WriteHeader(http.StatusUnprocessableEntity)
	}
	h.Render.RenderAdmin(w, "product_form", h.sessionCSRF(r), true, ProductFormData{
		Product: p, Categories: cats, Error: errMsg, IsNew: isNew, Action: action,
	})
}

// ProductNew renders a blank product form (empty JSON slices, so the
// edit form's toJSON textareas render [] rather than null).
func (h *AdminHandlers) ProductNew(w http.ResponseWriter, r *http.Request) {
	h.productForm(w, r, domain.Product{
		IsActive: true, Features: []string{}, Packaging: []string{},
		Colors: []domain.ColorSwatch{}, Images: []string{},
	}, true, "/admin/products", "")
}

// parseProductForm reads and validates the product form. JSON fields
// (features, packaging, specs, colors, images) are textarea JSON; an
// invalid JSON value is a validation error, not a silent default.
func parseProductForm(r *http.Request) (domain.Product, error) {
	var p domain.Product
	if err := r.ParseForm(); err != nil {
		return p, err
	}
	p.Name = strings.TrimSpace(r.FormValue("name"))
	p.Slug = strings.TrimSpace(r.FormValue("slug"))
	p.SKU = strings.TrimSpace(r.FormValue("sku"))
	if p.Name == "" || p.Slug == "" || p.SKU == "" {
		return p, errors.New("vui lòng nhập tên, slug và SKU")
	}
	catID, err := strconv.ParseInt(strings.TrimSpace(r.FormValue("category_id")), 10, 64)
	if err != nil || catID <= 0 {
		return p, errors.New("danh mục không hợp lệ")
	}
	p.CategoryID = catID
	p.CoatingRole = strings.TrimSpace(r.FormValue("coating_role"))
	p.FinishType = strings.TrimSpace(r.FormValue("finish_type"))
	p.UsageArea = strings.TrimSpace(r.FormValue("usage_area"))
	p.ShortSummary = strings.TrimSpace(r.FormValue("short_summary"))
	p.DescriptionHTML = r.FormValue("description_html")
	p.ApplicationGuideHTML = r.FormValue("application_guide_html")
	p.ThumbnailURL = strings.TrimSpace(r.FormValue("thumbnail_url"))
	p.TDSFileURL = strings.TrimSpace(r.FormValue("tds_file_url"))
	p.MSDSFileURL = strings.TrimSpace(r.FormValue("msds_file_url"))
	p.SEOTitle = strings.TrimSpace(r.FormValue("seo_title"))
	p.SEODescription = strings.TrimSpace(r.FormValue("seo_description"))
	p.IsFeatured = r.FormValue("is_featured") == "on"
	p.IsActive = r.FormValue("is_active") == "on"
	if v := strings.TrimSpace(r.FormValue("coverage_rate")); v != "" {
		cov, err := strconv.ParseFloat(v, 64)
		if err != nil || cov < 0 {
			return p, errors.New("độ phủ không hợp lệ")
		}
		p.CoverageRate = cov
	}
	// JSON textarea fields, each optional (empty = empty value).
	if v := strings.TrimSpace(r.FormValue("features_json")); v != "" {
		if err := json.Unmarshal([]byte(v), &p.Features); err != nil {
			return p, errors.New("features_json không phải JSON hợp lệ")
		}
	}
	if v := strings.TrimSpace(r.FormValue("packaging_json")); v != "" {
		if err := json.Unmarshal([]byte(v), &p.Packaging); err != nil {
			return p, errors.New("packaging_json không phải JSON hợp lệ")
		}
	}
	if v := strings.TrimSpace(r.FormValue("specs_json")); v != "" {
		if err := json.Unmarshal([]byte(v), &p.Specs); err != nil {
			return p, errors.New("specs_json không phải JSON hợp lệ")
		}
	}
	if v := strings.TrimSpace(r.FormValue("colors_json")); v != "" {
		if err := json.Unmarshal([]byte(v), &p.Colors); err != nil {
			return p, errors.New("colors_json không phải JSON hợp lệ")
		}
	}
	if v := strings.TrimSpace(r.FormValue("images_json")); v != "" {
		if err := json.Unmarshal([]byte(v), &p.Images); err != nil {
			return p, errors.New("images_json không phải JSON hợp lệ")
		}
	}
	if p.Features == nil {
		p.Features = []string{}
	}
	if p.Packaging == nil {
		p.Packaging = []string{}
	}
	if p.Colors == nil {
		p.Colors = []domain.ColorSwatch{}
	}
	if p.Images == nil {
		p.Images = []string{}
	}
	return p, nil
}

// ProductCreate creates a product from the admin form.
func (h *AdminHandlers) ProductCreate(w http.ResponseWriter, r *http.Request) {
	p, err := parseProductForm(r)
	if err != nil {
		h.productForm(w, r, p, true, "/admin/products", err.Error())
		return
	}
	if _, err := h.Products.Create(p); err != nil {
		h.productForm(w, r, p, true, "/admin/products", "Không tạo được sản phẩm: "+err.Error())
		return
	}
	http.Redirect(w, r, "/admin/products", http.StatusFound)
}

func (h *AdminHandlers) productID(r *http.Request) (int64, error) {
	return strconv.ParseInt(r.PathValue("id"), 10, 64)
}

// ProductEdit renders the edit form for one product.
func (h *AdminHandlers) ProductEdit(w http.ResponseWriter, r *http.Request) {
	id, err := h.productID(r)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	p, err := h.Products.GetByID(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.NotFound(w, r)
			return
		}
		http.Error(w, "catalog error", http.StatusInternalServerError)
		return
	}
	h.productForm(w, r, p, false, "/admin/products/"+strconv.FormatInt(id, 10), "")
}

// ProductUpdate saves edits to one product.
func (h *AdminHandlers) ProductUpdate(w http.ResponseWriter, r *http.Request) {
	id, err := h.productID(r)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	p, err := parseProductForm(r)
	if err != nil {
		p.ID = id
		h.productForm(w, r, p, false, "/admin/products/"+strconv.FormatInt(id, 10), err.Error())
		return
	}
	p.ID = id
	if err := h.Products.Update(p); err != nil {
		h.productForm(w, r, p, false, "/admin/products/"+strconv.FormatInt(id, 10), "Không lưu được sản phẩm: "+err.Error())
		return
	}
	http.Redirect(w, r, "/admin/products", http.StatusFound)
}

// ArticlesListData is the data for articles_list.html.
type ArticlesListData struct {
	Articles []domain.Article
}

// ArticlesList renders all articles (published and unpublished).
func (h *AdminHandlers) ArticlesList(w http.ResponseWriter, r *http.Request) {
	arts, err := h.Articles.ListAll()
	if err != nil {
		http.Error(w, "content error", http.StatusInternalServerError)
		return
	}
	h.Render.RenderAdmin(w, "articles_list", h.sessionCSRF(r), true, ArticlesListData{Articles: arts})
}

// ArticleFormData is the data for article_form.html.
// RelatedCategoryID is the dereferenced related category (0 = none),
// because templates cannot compare a *int64 with an int64 directly.
type ArticleFormData struct {
	Article           domain.Article
	Categories        []domain.Category
	RelatedCategoryID int64
	Error             string
	IsNew             bool
	Action            string
}

func (h *AdminHandlers) articleForm(w http.ResponseWriter, r *http.Request, a domain.Article, isNew bool, action, errMsg string) {
	cats, err := h.Categories.ListAll()
	if err != nil {
		http.Error(w, "catalog error", http.StatusInternalServerError)
		return
	}
	var relID int64
	if a.RelatedCategoryID != nil {
		relID = *a.RelatedCategoryID
	}
	if errMsg != "" {
		w.WriteHeader(http.StatusUnprocessableEntity)
	}
	h.Render.RenderAdmin(w, "article_form", h.sessionCSRF(r), true, ArticleFormData{
		Article: a, Categories: cats, RelatedCategoryID: relID, Error: errMsg, IsNew: isNew, Action: action,
	})
}

// ArticleNew renders a blank article form.
func (h *AdminHandlers) ArticleNew(w http.ResponseWriter, r *http.Request) {
	h.articleForm(w, r, domain.Article{IsPublished: true, AuthorName: "Kỹ Thuật Sơn Lotus"}, true, "/admin/articles", "")
}

func parseArticleForm(r *http.Request) (domain.Article, error) {
	var a domain.Article
	if err := r.ParseForm(); err != nil {
		return a, err
	}
	a.Type = strings.TrimSpace(r.FormValue("type"))
	a.Title = strings.TrimSpace(r.FormValue("title"))
	a.Slug = strings.TrimSpace(r.FormValue("slug"))
	if a.Type == "" || a.Title == "" || a.Slug == "" {
		return a, errors.New("vui lòng nhập loại, tiêu đề và slug")
	}
	switch a.Type {
	case domain.ArticleProject, domain.ArticleTechSupport, domain.ArticleSolution, domain.ArticleNews:
	default:
		return a, errors.New("loại bài viết không hợp lệ")
	}
	a.Excerpt = strings.TrimSpace(r.FormValue("excerpt"))
	a.ContentHTML = r.FormValue("content_html")
	a.ThumbnailURL = strings.TrimSpace(r.FormValue("thumbnail_url"))
	a.AuthorName = strings.TrimSpace(r.FormValue("author_name"))
	a.SEOTitle = strings.TrimSpace(r.FormValue("seo_title"))
	a.SEODescription = strings.TrimSpace(r.FormValue("seo_description"))
	a.IsFeatured = r.FormValue("is_featured") == "on"
	a.IsPublished = r.FormValue("is_published") == "on"
	if v := strings.TrimSpace(r.FormValue("related_category_id")); v != "" {
		catID, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return a, errors.New("danh mục liên quan không hợp lệ")
		}
		a.RelatedCategoryID = &catID
	}
	return a, nil
}

// ArticleCreate creates an article from the admin form.
func (h *AdminHandlers) ArticleCreate(w http.ResponseWriter, r *http.Request) {
	a, err := parseArticleForm(r)
	if err != nil {
		h.articleForm(w, r, a, true, "/admin/articles", err.Error())
		return
	}
	if _, err := h.Articles.Create(a); err != nil {
		h.articleForm(w, r, a, true, "/admin/articles", "Không tạo được bài viết: "+err.Error())
		return
	}
	http.Redirect(w, r, "/admin/articles", http.StatusFound)
}

// ArticleEdit renders the edit form for one article.
func (h *AdminHandlers) ArticleEdit(w http.ResponseWriter, r *http.Request) {
	id, err := h.productID(r)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	a, err := h.Articles.GetByID(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.NotFound(w, r)
			return
		}
		http.Error(w, "content error", http.StatusInternalServerError)
		return
	}
	h.articleForm(w, r, a, false, "/admin/articles/"+strconv.FormatInt(id, 10), "")
}

// ArticleUpdate saves edits to one article.
func (h *AdminHandlers) ArticleUpdate(w http.ResponseWriter, r *http.Request) {
	id, err := h.productID(r)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	a, err := parseArticleForm(r)
	if err != nil {
		a.ID = id
		h.articleForm(w, r, a, false, "/admin/articles/"+strconv.FormatInt(id, 10), err.Error())
		return
	}
	a.ID = id
	if err := h.Articles.Update(a); err != nil {
		h.articleForm(w, r, a, false, "/admin/articles/"+strconv.FormatInt(id, 10), "Không lưu được bài viết: "+err.Error())
		return
	}
	http.Redirect(w, r, "/admin/articles", http.StatusFound)
}

// InquiriesListData is the data for inquiries_list.html.
type InquiriesListData struct {
	Inquiries []domain.Inquiry
}

// InquiriesList renders all inquiries, newest first.
func (h *AdminHandlers) InquiriesList(w http.ResponseWriter, r *http.Request) {
	inqs, err := h.Inquiries.List()
	if err != nil {
		http.Error(w, "inquiry error", http.StatusInternalServerError)
		return
	}
	h.Render.RenderAdmin(w, "inquiries_list", h.sessionCSRF(r), true, InquiriesListData{Inquiries: inqs})
}

// InquiryStatus updates one inquiry's status (new/contacted/quoted/closed).
// Only membership in the allowed set is enforced; no sequential-only
// transition rule (see repository.ValidStatus).
func (h *AdminHandlers) InquiryStatus(w http.ResponseWriter, r *http.Request) {
	id, err := h.productID(r)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	status := strings.TrimSpace(r.FormValue("status"))
	if !repository.ValidStatus(status) {
		http.Error(w, "invalid status", http.StatusBadRequest)
		return
	}
	if err := h.Inquiries.UpdateStatus(id, status); err != nil {
		http.Error(w, "inquiry error", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/admin/inquiries", http.StatusFound)
}
