package middleware

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"net/http"
)

// SecurityHeaders applies the explicit header set from the approved
// security note. The CSP trade-off is stated, not hidden: 'unsafe-eval'
// is required because Alpine.js evaluates expressions, and
// 'unsafe-inline' styles because Alpine :style sets inline styles.
// An Alpine CSP build is a Phase 4 consideration, not a Phase 3 change.
// HSTS is set only when served over HTTPS in production (r.TLS != nil
// and production mode); it is never forced on plain-HTTP local dev.
func SecurityHeaders(production bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			h := w.Header()
			h.Set("X-Content-Type-Options", "nosniff")
			h.Set("Referrer-Policy", "strict-origin-when-cross-origin")
			h.Set("X-Frame-Options", "DENY")
			h.Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
			h.Set("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-eval'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; frame-ancestors 'none'")
			if production && r.TLS != nil {
				h.Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
			}
			next.ServeHTTP(w, r)
		})
	}
}

// randomToken returns a 32-byte crypto/rand token as hex (64 chars).
func randomToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// equalTokens compares two tokens in constant time (length mismatch
// is a rejection, not a timing oracle shortcut in the compare itself).
func equalTokens(a, b string) bool {
	if len(a) != len(b) || a == "" || b == "" {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}
