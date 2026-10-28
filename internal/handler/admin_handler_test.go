package handler_test

import (
	"bytes"
	"database/sql"
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"sonlotus/internal/database"
	"sonlotus/internal/domain"
	"sonlotus/internal/handler"
	"sonlotus/internal/middleware"
	"sonlotus/internal/repository"
	"sonlotus/internal/service"
)

type adminTestEnv struct {
	handler   http.Handler
	db        *sql.DB
	csrf      *middleware.CSRFStore
	inqs      *repository.InquiryRepo
	uploadDir string
}

func newAdminTestEnv(t *testing.T) *adminTestEnv {
	return newAdminTestEnvWithLimiter(t, 1000)
}

func newAdminTestEnvWithLimiter(t *testing.T, chainLimit int) *adminTestEnv {
	t.Helper()
	path := filepath.Join(t.TempDir(), "admin.db")
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
	adminRend, err := handler.NewAdminRenderer()
	if err != nil {
		t.Fatalf("admin renderer: %v", err)
	}
	auth, err := middleware.NewAdminAuth("admin", "secret123", &repository.SessionRepo{DB: db})
	if err != nil {
		t.Fatalf("auth: %v", err)
	}
	csrf := middleware.NewCSRFStore()
	mux := handler.NewMux(&handler.Deps{
		Categories: cats, Products: prods, Articles: arts, Inquiries: inqs, Redirects: reds,
		Catalog: &service.CatalogService{Categories: cats, Products: prods},
		Content: &service.ContentService{Articles: arts, Products: prods},
		SEO:     &service.SEOService{},
		Render:  rend,
	})
	uploadDir := filepath.Join(t.TempDir(), "uploads")
	handler.RegisterAdmin(mux, &handler.AdminDeps{
		Categories: cats, Products: prods, Articles: arts, Inquiries: inqs, Redirects: reds,
		Auth: auth, CSRF: csrf, Render: adminRend, UploadDir: uploadDir,
	})
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	chain := middleware.Chain(mux, middleware.ChainConfig{
		Logger: logger, Redirects: reds, IsRealRoute: handler.IsRealRoute,
		Auth: auth, CSRF: csrf, ChainLimiter: middleware.NewLimiter(chainLimit, 15*time.Minute),
	})
	return &adminTestEnv{handler: chain, db: db, csrf: csrf, inqs: inqs, uploadDir: uploadDir}
}

func (e *adminTestEnv) do(req *http.Request) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	e.handler.ServeHTTP(rec, req)
	return rec
}

var csrfRe = regexp.MustCompile(`name="csrf_token" value="([^"]+)"`)

func extractCSRF(t *testing.T, body string) string {
	t.Helper()
	m := csrfRe.FindStringSubmatch(body)
	if m == nil {
		t.Fatalf("no csrf_token in body: %.200s", body)
	}
	return m[1]
}

func cookieValue(rec *httptest.ResponseRecorder, name string) string {
	for _, c := range rec.Result().Cookies() {
		if c.Name == name {
			return c.Value
		}
	}
	return ""
}

func (e *adminTestEnv) loginOK(t *testing.T) string {
	t.Helper()
	get := httptest.NewRequest(http.MethodGet, "/admin/login", nil)
	rec := e.do(get)
	loginID := cookieValue(rec, "csrf_login_id")
	token := extractCSRF(t, rec.Body.String())
	form := url.Values{"username": {"admin"}, "password": {"secret123"}, "csrf_token": {token}}
	req := httptest.NewRequest(http.MethodPost, "/admin/login", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: "csrf_login_id", Value: loginID})
	req.RemoteAddr = "127.0.0.1:1234"
	rec = e.do(req)
	if rec.Code != http.StatusFound {
		t.Fatalf("login status = %d, want 302, body: %.300s", rec.Code, rec.Body.String())
	}
	sess := cookieValue(rec, middleware.SessionCookie)
	if sess == "" {
		t.Fatal("login did not set a session cookie")
	}
	return sess
}

func (e *adminTestEnv) sessionCSRF(t *testing.T, sess string) string {
	t.Helper()
	tok, err := e.csrf.TokenFor(middleware.SessionKey(repository.HashToken(sess)))
	if err != nil {
		t.Fatalf("session csrf: %v", err)
	}
	return tok
}

func (e *adminTestEnv) authed(t *testing.T, method, path, sess string, form url.Values) *httptest.ResponseRecorder {
	t.Helper()
	var body io.Reader
	if form != nil {
		body = strings.NewReader(form.Encode())
	}
	req := httptest.NewRequest(method, path, body)
	if form != nil {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}
	req.AddCookie(&http.Cookie{Name: middleware.SessionCookie, Value: sess})
	req.RemoteAddr = "127.0.0.1:1234"
	return e.do(req)
}

func TestAdminUnauthenticatedRedirect(t *testing.T) {
	e := newAdminTestEnv(t)
	for _, path := range []string{"/admin", "/admin/products", "/admin/inquiries", "/admin/articles"} {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		rec := e.do(req)
		if rec.Code != http.StatusFound || rec.Header().Get("Location") != "/admin/login" {
			t.Errorf("GET %s: status=%d location=%q, want 302 to /admin/login", path, rec.Code, rec.Header().Get("Location"))
		}
	}
}

func TestAdminLoginGenericFailure(t *testing.T) {
	e := newAdminTestEnv(t)
	get := httptest.NewRequest(http.MethodGet, "/admin/login", nil)
	rec := e.do(get)
	loginID := cookieValue(rec, "csrf_login_id")
	token := extractCSRF(t, rec.Body.String())
	for _, user := range []string{"admin", "nosuchuser"} {
		form := url.Values{"username": {user}, "password": {"wrong"}, "csrf_token": {token}}
		req := httptest.NewRequest(http.MethodPost, "/admin/login", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.AddCookie(&http.Cookie{Name: "csrf_login_id", Value: loginID})
		req.RemoteAddr = "127.0.0.1:1234"
		rec = e.do(req)
		if rec.Code != http.StatusUnauthorized {
			t.Errorf("login %s: status=%d, want 401", user, rec.Code)
		}
		if !strings.Contains(rec.Body.String(), "Tên đăng nhập hoặc mật khẩu không đúng") {
			t.Errorf("login %s: body lacks generic error", user)
		}
	}
}

func TestAdminLoginSuccessLogoutInvalidation(t *testing.T) {
	e := newAdminTestEnv(t)
	sess := e.loginOK(t)
	if rec := e.authed(t, http.MethodGet, "/admin", sess, nil); rec.Code != http.StatusOK {
		t.Fatalf("GET /admin after login: %d, want 200", rec.Code)
	}
	form := url.Values{"csrf_token": {e.sessionCSRF(t, sess)}}
	rec := e.authed(t, http.MethodPost, "/admin/logout", sess, form)
	if rec.Code != http.StatusFound {
		t.Fatalf("logout: %d, want 302", rec.Code)
	}
	if rec := e.authed(t, http.MethodGet, "/admin", sess, nil); rec.Code != http.StatusFound {
		t.Fatalf("GET /admin after logout: %d, want 302 to login", rec.Code)
	}
}

func TestAdminCSRFRejection(t *testing.T) {
	e := newAdminTestEnv(t)
	sess := e.loginOK(t)
	form := url.Values{"name": {"X"}, "slug": {"x"}, "sku": {"X"}, "category_id": {"1"}, "csrf_token": {"bogus"}}
	if rec := e.authed(t, http.MethodPost, "/admin/products", sess, form); rec.Code != http.StatusForbidden {
		t.Errorf("product create with bad CSRF: %d, want 403", rec.Code)
	}
}

func TestAdminUploadRejection(t *testing.T) {
	e := newAdminTestEnv(t)
	sess := e.loginOK(t)
	upload := func(filename string, data []byte) *httptest.ResponseRecorder {
		var buf bytes.Buffer
		w := multipart.NewWriter(&buf)
		_ = w.WriteField("csrf_token", e.sessionCSRF(t, sess))
		fw, _ := w.CreateFormFile("file", filename)
		_, _ = fw.Write(data)
		_ = w.Close()
		req := httptest.NewRequest(http.MethodPost, "/admin/upload", &buf)
		req.Header.Set("Content-Type", w.FormDataContentType())
		req.AddCookie(&http.Cookie{Name: middleware.SessionCookie, Value: sess})
		req.RemoteAddr = "127.0.0.1:1234"
		return e.do(req)
	}
	if rec := upload("evil.html", []byte("<script>alert(1)</script>")); rec.Code != http.StatusBadRequest {
		t.Errorf("upload .html: %d, want 400", rec.Code)
	}
	if rec := upload("notes.txt", []byte("hello")); rec.Code != http.StatusBadRequest {
		t.Errorf("upload .txt: %d, want 400", rec.Code)
	}
	png := []byte{0x89, 'P', 'N', 'G', 0x0d, 0x0a, 0x1a, 0x0a, 1, 2, 3}
	if rec := upload("ok.png", png); rec.Code != http.StatusFound {
		t.Errorf("upload .png: %d, want 302", rec.Code)
	}
}

func TestAdminInquiryStatusTransition(t *testing.T) {
	e := newAdminTestEnv(t)
	id, err := e.inqs.Create(domain.Inquiry{CustomerName: "Test", Phone: "0900000000", Message: "hello", InquiryType: "quote"})
	if err != nil {
		t.Fatalf("create inquiry: %v", err)
	}
	sess := e.loginOK(t)
	form := url.Values{"status": {"contacted"}, "csrf_token": {e.sessionCSRF(t, sess)}}
	rec := e.authed(t, http.MethodPost, fmt.Sprintf("/admin/inquiries/%d/status", id), sess, form)
	if rec.Code != http.StatusFound {
		t.Fatalf("status transition: %d, want 302", rec.Code)
	}
	list, err := e.inqs.List()
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	for _, inq := range list {
		if inq.ID == id && inq.Status != "contacted" {
			t.Errorf("inquiry status = %q, want contacted", inq.Status)
		}
	}
}

func TestAdminLoginRateLimit(t *testing.T) {
	e := newAdminTestEnv(t)
	get := httptest.NewRequest(http.MethodGet, "/admin/login", nil)
	rec := e.do(get)
	loginID := cookieValue(rec, "csrf_login_id")
	token := extractCSRF(t, rec.Body.String())
	try := func() int {
		form := url.Values{"username": {"admin"}, "password": {"wrong"}, "csrf_token": {token}}
		req := httptest.NewRequest(http.MethodPost, "/admin/login", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.AddCookie(&http.Cookie{Name: "csrf_login_id", Value: loginID})
		req.RemoteAddr = "127.0.0.1:1234"
		return e.do(req).Code
	}
	for i := 0; i < 5; i++ {
		if code := try(); code != http.StatusUnauthorized {
			t.Fatalf("failed attempt %d: %d, want 401", i+1, code)
		}
	}
	if code := try(); code != http.StatusTooManyRequests {
		t.Errorf("6th failed attempt: %d, want 429", code)
	}
}

// --- Fix-round tests ---

func TestAdminUnauthenticatedPOSTsRedirect(t *testing.T) {
	e := newAdminTestEnv(t)
	paths := []string{
		"/admin/products", "/admin/products/1", "/admin/articles", "/admin/articles/1",
		"/admin/inquiries/1/status", "/admin/upload", "/admin/logout",
	}
	for _, path := range paths {
		form := url.Values{"csrf_token": {"bogus"}}
		req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.RemoteAddr = "127.0.0.1:1234"
		rec := e.do(req)
		if rec.Code != http.StatusFound || rec.Header().Get("Location") != "/admin/login" {
			t.Errorf("POST %s unauth: status=%d location=%q, want 302 to /admin/login", path, rec.Code, rec.Header().Get("Location"))
		}
	}
}

func TestAdminCSRFRejectionLoginUploadStatus(t *testing.T) {
	e := newAdminTestEnv(t)
	// Login with a bogus token must be 403 (chain CSRF, before auth).
	get := httptest.NewRequest(http.MethodGet, "/admin/login", nil)
	rec := e.do(get)
	loginID := cookieValue(rec, "csrf_login_id")
	form := url.Values{"username": {"admin"}, "password": {"secret123"}, "csrf_token": {"bogus"}}
	req := httptest.NewRequest(http.MethodPost, "/admin/login", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: "csrf_login_id", Value: loginID})
	req.RemoteAddr = "127.0.0.1:1234"
	if rec := e.do(req); rec.Code != http.StatusForbidden {
		t.Errorf("login bad CSRF: %d, want 403", rec.Code)
	}
	sess := e.loginOK(t)
	// Upload with a bogus token must be 403.
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	_ = w.WriteField("csrf_token", "bogus")
	fw, _ := w.CreateFormFile("file", "ok.png")
	_, _ = fw.Write([]byte{0x89, 'P', 'N', 'G'})
	_ = w.Close()
	req = httptest.NewRequest(http.MethodPost, "/admin/upload", &buf)
	req.Header.Set("Content-Type", w.FormDataContentType())
	req.AddCookie(&http.Cookie{Name: middleware.SessionCookie, Value: sess})
	req.RemoteAddr = "127.0.0.1:1234"
	if rec := e.do(req); rec.Code != http.StatusForbidden {
		t.Errorf("upload bad CSRF: %d, want 403", rec.Code)
	}
	// Inquiry status with a bogus token must be 403.
	id, err := e.inqs.Create(domain.Inquiry{CustomerName: "T", Phone: "0900000000", Message: "hi", InquiryType: "quote"})
	if err != nil {
		t.Fatal(err)
	}
	form = url.Values{"status": {"contacted"}, "csrf_token": {"bogus"}}
	if rec := e.authed(t, http.MethodPost, fmt.Sprintf("/admin/inquiries/%d/status", id), sess, form); rec.Code != http.StatusForbidden {
		t.Errorf("status bad CSRF: %d, want 403", rec.Code)
	}
}

func TestAdminUploadOversizeRejected(t *testing.T) {
	e := newAdminTestEnv(t)
	sess := e.loginOK(t)
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	_ = w.WriteField("csrf_token", e.sessionCSRF(t, sess))
	fw, _ := w.CreateFormFile("file", "big.png")
	_, _ = fw.Write(bytes.Repeat([]byte{'x'}, handler.MaxUploadBytes+1))
	_ = w.Close()
	req := httptest.NewRequest(http.MethodPost, "/admin/upload", &buf)
	req.Header.Set("Content-Type", w.FormDataContentType())
	req.AddCookie(&http.Cookie{Name: middleware.SessionCookie, Value: sess})
	req.RemoteAddr = "127.0.0.1:1234"
	if rec := e.do(req); rec.Code != http.StatusRequestEntityTooLarge {
		t.Errorf("oversize upload: %d, want 413", rec.Code)
	}
}

func TestUploadsServeTraversalAndNonAllowlisted(t *testing.T) {
	e := newAdminTestEnv(t)
	sess := e.loginOK(t)
	// Place a file via the real upload path, then try traversal and
	// non-allowlisted serving. Upload dir is reached through the env's
	// handler only via mux, so use direct file placement in the test
	// temp upload area via a successful upload first.
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	_ = w.WriteField("csrf_token", e.sessionCSRF(t, sess))
	fw, _ := w.CreateFormFile("file", "ok.png")
	_, _ = fw.Write([]byte{0x89, 'P', 'N', 'G', 1, 2, 3})
	_ = w.Close()
	req := httptest.NewRequest(http.MethodPost, "/admin/upload", &buf)
	req.Header.Set("Content-Type", w.FormDataContentType())
	req.AddCookie(&http.Cookie{Name: middleware.SessionCookie, Value: sess})
	req.RemoteAddr = "127.0.0.1:1234"
	if rec := e.do(req); rec.Code != http.StatusFound {
		t.Fatalf("setup upload: %d, want 302", rec.Code)
	}
	get := func(path string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		return e.do(req)
	}
	if rec := get("/uploads/ok.png"); rec.Code != http.StatusOK {
		t.Errorf("serve ok.png: %d, want 200", rec.Code)
	}
	// Non-allowlisted extension must be 404 even when the file
	// exists on disk (placed directly, since upload rejects .html).
	if err := os.MkdirAll(e.uploadDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(e.uploadDir, "evil.html"), []byte("<script>evil</script>"), 0o644); err != nil {
		t.Fatal(err)
	}
	if rec := get("/uploads/evil.html"); rec.Code != http.StatusNotFound {
		t.Errorf("serve existing .html: %d, want 404", rec.Code)
	}
	// Traversal must not serve a file outside the upload dir.
	secret := "TOPSECRET-TRAVERSAL"
	if err := os.WriteFile(filepath.Join(filepath.Dir(e.uploadDir), "secret.txt"), []byte(secret), 0o644); err != nil {
		t.Fatal(err)
	}
	rec := get("/uploads/..%2Fsecret.txt")
	if rec.Code == http.StatusOK || strings.Contains(rec.Body.String(), secret) {
		t.Errorf("traversal serve: status=%d, want non-200 without secret content", rec.Code)
	}
}

func TestRedirectChainAndLoop(t *testing.T) {
	e := newAdminTestEnv(t)
	add := func(old, new string, code int) {
		if _, err := e.db.Exec(`INSERT INTO url_redirects (old_path, new_path, status_code) VALUES (?, ?, ?)`, old, new, code); err != nil {
			t.Fatal(err)
		}
	}
	add("/old-a", "/old-b", 301)
	add("/old-b", "/final-c", 301)
	add("/loop-a", "/loop-b", 301)
	add("/loop-b", "/loop-a", 301)
	add("/self", "/self", 301)
	get := func(path string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		return e.do(req)
	}
	if rec := get("/old-a"); rec.Code != http.StatusMovedPermanently || rec.Header().Get("Location") != "/final-c" {
		t.Errorf("chain A->B->C: status=%d location=%q, want 301 to /final-c", rec.Code, rec.Header().Get("Location"))
	}
	for _, path := range []string{"/loop-a", "/self"} {
		rec := get(path)
		if rec.Code == http.StatusMovedPermanently || rec.Code == http.StatusFound {
			t.Errorf("loop %s: issued redirect to %q, want no redirect", path, rec.Header().Get("Location"))
		}
	}
}

func TestInquirySpamDoesNotLockOutAdminLoginProductionLimit(t *testing.T) {
	// Production chain-limiter value: 100 POSTs / 15 min / IP.
	e := newAdminTestEnvWithLimiter(t, 100)
	postInquiry := func() int {
		form := url.Values{"csrf_token": {"bogus"}}
		req := httptest.NewRequest(http.MethodPost, "/lien-he", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.RemoteAddr = "127.0.0.1:1234"
		return e.do(req).Code
	}
	for i := 0; i < 100; i++ {
		if code := postInquiry(); code == http.StatusTooManyRequests {
			t.Fatalf("inquiry %d: got 429 early, want limiter at 100", i+1)
		}
	}
	if code := postInquiry(); code != http.StatusTooManyRequests {
		t.Errorf("101st inquiry: %d, want 429 (limiter active at production value)", code)
	}
	// Admin login from the same IP must still succeed: the coarse
	// inquiry bucket must not block login.
	sess := e.loginOK(t)
	if sess == "" {
		t.Fatal("login after inquiry spam did not yield a session")
	}
}
