package handler

import (
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"sonlotus/internal/domain"
	"sonlotus/internal/middleware"
)

// ContactData is the data for contact.html.
type ContactData struct {
	Success   bool
	Error     string
	CSRFToken string
}

// contactCSRFToken returns the double-submit token to render into the
// form: the cookie value on POST (the browser sent it), or "" on a
// fresh GET (ContactPage sets the cookie and passes the token).
func contactCSRFToken(r *http.Request) string {
	if c, err := r.Cookie("csrf_token"); err == nil {
		return c.Value
	}
	return ""
}

// ipLimiter is a tiny per-IP rate limiter for POST /lien-he (SPEC §8).
// It uses the client IP from RemoteAddr only — never an untrusted
// X-Forwarded-For header (PLAN §4). A full middleware rate limiter
// arrives with Phase 3; this endpoint-scoped limiter is not a second
// system, it is the same rule applied at the endpoint until then.
type ipLimiter struct {
	mu     sync.Mutex
	hits   map[string][]time.Time
	limit  int
	window time.Duration
}

func newIPLimiter(limit int, window time.Duration) *ipLimiter {
	return &ipLimiter{hits: map[string][]time.Time{}, limit: limit, window: window}
}

func (l *ipLimiter) allow(ip string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := time.Now()
	cutoff := now.Add(-l.window)
	var kept []time.Time
	for _, t := range l.hits[ip] {
		if t.After(cutoff) {
			kept = append(kept, t)
		}
	}
	if len(kept) >= l.limit {
		l.hits[ip] = kept
		return false
	}
	l.hits[ip] = append(kept, now)
	return true
}

var contactLimiter = newIPLimiter(10, time.Minute)

func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

// ContactPage renders the contact / inquiry form page. The form uses
// the double-submit CSRF token (cookie + hidden field, no session):
// the cookie is set here and the same value is rendered into the form.
func (h *Handlers) ContactPage(w http.ResponseWriter, r *http.Request) {
	seo := h.SEO.HomeSEO()
	seo.Title = "Liên hệ — Sơn Lotus"
	seo.Description = "Liên hệ Sơn Lotus: tư vấn kỹ thuật, báo giá, đăng ký đại lý."
	seo.CanonicalURL = h.SEO.Canonical("/lien-he")
	tok, err := middleware.EnsureDoubleSubmit(w, r)
	if err != nil {
		http.Error(w, "csrf error", http.StatusInternalServerError)
		return
	}
	h.Render.Render(w, "contact", seo, ContactData{CSRFToken: tok})
}

// SubmitInquiry validates and saves an inquiry (SPEC §6.1).
// Honeypot field website_url must be empty; name/message have length caps.
func (h *Handlers) SubmitInquiry(w http.ResponseWriter, r *http.Request) {
	seo := h.SEO.HomeSEO()
	seo.Title = "Liên hệ — Sơn Lotus"
	seo.CanonicalURL = h.SEO.Canonical("/lien-he")
	fail := func(msg string) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		h.Render.Render(w, "contact", seo, ContactData{Error: msg, CSRFToken: contactCSRFToken(r)})
	}
	if !contactLimiter.allow(clientIP(r)) {
		w.WriteHeader(http.StatusTooManyRequests)
		h.Render.Render(w, "contact", seo, ContactData{Error: "Bạn gửi quá nhiều yêu cầu, vui lòng thử lại sau ít phút.", CSRFToken: contactCSRFToken(r)})
		return
	}
	if err := r.ParseForm(); err != nil {
		fail("Dữ liệu gửi lên không hợp lệ.")
		return
	}
	// CSRF (double-submit): the chain already validated this, but a
	// directly-wired mux gets the same check here before any inquiry
	// is created.
	if !middleware.ValidateDoubleSubmit(r) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	// Honeypot: bots fill this hidden field; humans leave it empty.
	if strings.TrimSpace(r.FormValue("website_url")) != "" {
		fail("Yêu cầu không hợp lệ.")
		return
	}
	name := strings.TrimSpace(r.FormValue("ten"))
	phone := strings.TrimSpace(r.FormValue("dien-thoai"))
	message := strings.TrimSpace(r.FormValue("noi-dung"))
	if name == "" || phone == "" || message == "" {
		fail("Vui lòng nhập đầy đủ họ tên, số điện thoại và nội dung.")
		return
	}
	if len(name) > 100 || len(phone) > 30 || len(message) > 2000 {
		fail("Nội dung vượt quá giới hạn cho phép.")
		return
	}
	area, _ := strconv.ParseFloat(r.FormValue("dien-tich"), 64)
	inq := domain.Inquiry{
		CustomerName:  name,
		Phone:         phone,
		Email:         strings.TrimSpace(r.FormValue("email")),
		CompanyName:   strings.TrimSpace(r.FormValue("cong-ty")),
		InquiryType:   "quote",
		SurfaceAreaM2: area,
		Message:       message,
		IPAddress:     clientIP(r),
	}
	if t := strings.TrimSpace(r.FormValue("loai-yeu-cau")); t == "tech_support" || t == "dealer" || t == "general" {
		inq.InquiryType = t
	}
	if _, err := h.Inquiries.Create(inq); err != nil {
		http.Error(w, "inquiry error", http.StatusInternalServerError)
		return
	}
	if isHTMX(r) {
		h.Render.RenderPartial(w, "inquiry_success", nil)
		return
	}
	h.Render.Render(w, "contact", seo, ContactData{Success: true, CSRFToken: contactCSRFToken(r)})
}
