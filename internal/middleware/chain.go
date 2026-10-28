package middleware

import (
	"log/slog"
	"net/http"

	"sonlotus/internal/repository"
)

// ChainConfig carries everything the fixed chain needs.
type ChainConfig struct {
	Logger      *slog.Logger
	Production  bool // APP_ENV=production: Secure cookies, HSTS over TLS
	Redirects   *repository.RedirectRepo
	IsRealRoute func(string) bool
	Auth        *AdminAuth
	CSRF        *CSRFStore
	// ChainLimiter is the coarse abuse guard for the public inquiry
	// POST only (see RateLimit); it is never applied to admin login.
	// Login failure limiting lives in the handler.
	ChainLimiter *Limiter
}

// Chain wraps next so middleware executes in the fixed PLAN §6 order:
// recovery → logging → security headers → redirect → rate limit →
// admin auth → CSRF. (Wrapping is innermost-first, so the list order
// is the execution order: Recovery is applied last and runs first.)
func Chain(next http.Handler, cfg ChainConfig) http.Handler {
	h := next
	h = CSRF(cfg.CSRF)(h)
	h = RequireAdmin(cfg.Auth)(h)
	h = RateLimit(cfg.ChainLimiter)(h)
	h = Redirect(cfg.Redirects, cfg.IsRealRoute)(h)
	h = SecurityHeaders(cfg.Production)(h)
	h = Logging(cfg.Logger)(h)
	h = Recovery(cfg.Logger)(h)
	return h
}
