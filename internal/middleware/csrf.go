package middleware

import (
	"net/http"
	"strings"
	"sync"
)

// CSRFStore holds synchronizer tokens server-side, keyed by a session
// identifier. Admin/session forms use this store (approved security
// note); the token is compared in constant time. The store holds no
// passwords and no session tokens themselves — keys are session hashes
// (admin) or login-form ids (pre-auth login form).
type CSRFStore struct {
	mu     sync.Mutex
	tokens map[string]string
}

// NewCSRFStore returns an empty synchronizer-token store.
func NewCSRFStore() *CSRFStore {
	return &CSRFStore{tokens: map[string]string{}}
}

// TokenFor returns the stored token for key, generating and storing
// one on first use.
func (s *CSRFStore) TokenFor(key string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if tok, ok := s.tokens[key]; ok {
		return tok, nil
	}
	tok, err := randomToken()
	if err != nil {
		return "", err
	}
	s.tokens[key] = tok
	return tok, nil
}

// Validate reports whether submitted matches the stored token for key.
func (s *CSRFStore) Validate(key, submitted string) bool {
	s.mu.Lock()
	tok, ok := s.tokens[key]
	s.mu.Unlock()
	if !ok {
		return false
	}
	return equalTokens(tok, submitted)
}

// Delete drops the token for key (logout / session rotation).
func (s *CSRFStore) Delete(key string) {
	s.mu.Lock()
	delete(s.tokens, key)
	s.mu.Unlock()
}

// SessionKey converts a session hash into a CSRF store key.
func SessionKey(sessionHash string) string { return "session:" + sessionHash }

// LoginKey converts a login-form id into a CSRF store key.
func LoginKey(loginID string) string { return "login:" + loginID }

const (
	// CSRFField is the hidden form field name for all CSRF tokens.
	CSRFField = "csrf_token"
	// loginIDCookie carries the login form's synchronizer key id.
	loginIDCookie = "csrf_login_id"
	// doubleSubmitCookie carries the public inquiry double-submit token.
	doubleSubmitCookie = "csrf_token"
)

// EnsureLoginID returns the login-form id cookie value, setting a fresh
// one when absent. The id itself is not a secret credential; the token
// it keys lives server-side in the CSRFStore.
func EnsureLoginID(w http.ResponseWriter, r *http.Request) (string, error) {
	if c, err := r.Cookie(loginIDCookie); err == nil && c.Value != "" {
		return c.Value, nil
	}
	id, err := randomToken()
	if err != nil {
		return "", err
	}
	http.SetCookie(w, &http.Cookie{
		Name: loginIDCookie, Value: id, Path: "/admin",
		HttpOnly: true, SameSite: http.SameSiteLaxMode,
	})
	return id, nil
}

// LoginIDFromRequest returns the login-form id from the request cookie.
func LoginIDFromRequest(r *http.Request) string {
	c, err := r.Cookie(loginIDCookie)
	if err != nil {
		return ""
	}
	return c.Value
}

// EnsureDoubleSubmit returns the public-form double-submit token,
// setting the cookie when absent. The public inquiry form has no
// session, so the token is a double-submit cookie + hidden field pair
// (approved security note), compared for equality on POST.
func EnsureDoubleSubmit(w http.ResponseWriter, r *http.Request) (string, error) {
	if c, err := r.Cookie(doubleSubmitCookie); err == nil && c.Value != "" {
		return c.Value, nil
	}
	tok, err := randomToken()
	if err != nil {
		return "", err
	}
	http.SetCookie(w, &http.Cookie{
		Name: doubleSubmitCookie, Value: tok, Path: "/",
		HttpOnly: true, SameSite: http.SameSiteLaxMode,
	})
	return tok, nil
}

// ValidateDoubleSubmit compares the double-submit cookie with the
// hidden field value, in constant time.
func ValidateDoubleSubmit(r *http.Request) bool {
	c, err := r.Cookie(doubleSubmitCookie)
	if err != nil {
		return false
	}
	return equalTokens(c.Value, r.FormValue(CSRFField))
}

// CSRF is the chain's CSRF step (after auth, so an admin session hash
// is available in context for session-bound forms). It validates:
//   - POST /admin/login: login synchronizer token (pre-auth store key).
//   - POST /admin/* (urlencoded): session synchronizer token.
//   - POST /lien-he: public double-submit token.
//
// Multipart admin POSTs (the upload endpoint) are NOT parsed here:
// parsing a multipart body here would read the file before the upload
// handler can cap it with http.MaxBytesReader. Those handlers validate
// with the same CSRFStore after capping the reader; see admin_handler.
// Non-state-changing POSTs (the calculator HTMX endpoint, which writes
// nothing) are exempt by design — the note's CSRF list covers login,
// admin CRUD, upload, inquiry status, and the inquiry form.
func CSRF(store *CSRFStore) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				next.ServeHTTP(w, r)
				return
			}
			path := r.URL.Path
			switch {
			case path == "/admin/login":
				if err := r.ParseForm(); err != nil {
					http.Error(w, "Forbidden", http.StatusForbidden)
					return
				}
				if !store.Validate(LoginKey(LoginIDFromRequest(r)), r.FormValue(CSRFField)) {
					http.Error(w, "Forbidden", http.StatusForbidden)
					return
				}
			case strings.HasPrefix(path, "/admin/"):
				if strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data") {
					// Validated by the upload handler after MaxBytesReader.
					break
				}
				if err := r.ParseForm(); err != nil {
					http.Error(w, "Forbidden", http.StatusForbidden)
					return
				}
				hash, ok := SessionHashFromContext(r.Context())
				if !ok || !store.Validate(SessionKey(hash), r.FormValue(CSRFField)) {
					http.Error(w, "Forbidden", http.StatusForbidden)
					return
				}
			case path == "/lien-he":
				if err := r.ParseForm(); err != nil {
					http.Error(w, "Forbidden", http.StatusForbidden)
					return
				}
				if !ValidateDoubleSubmit(r) {
					http.Error(w, "Forbidden", http.StatusForbidden)
					return
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}
