package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"sonlotus/internal/repository"
)

// SessionCookie is the admin session cookie name.
const SessionCookie = "sonlotus_admin"

// SessionExpiry is the approved 12-hour admin session lifetime.
const SessionExpiry = 12 * time.Hour

type ctxKey string

const sessionHashKey ctxKey = "admin-session-hash"

// SessionHashFromContext returns the validated session hash stored by
// RequireAdmin, if any.
func SessionHashFromContext(ctx context.Context) (string, bool) {
	v, ok := ctx.Value(sessionHashKey).(string)
	return v, ok
}

// AdminAuth owns admin credential verification and session lifecycle.
// The password hash lives only in memory (hashed at server startup
// from the env password with bcrypt cost 12); no plaintext password is
// stored, logged, or compared.
type AdminAuth struct {
	User         string
	passwordHash []byte
	// dummyHash is a bcrypt hash of a random, never-known value. It
	// exists only so an unknown-user login still runs one bcrypt
	// comparison (same cost as a real password check) instead of
	// short-circuiting — otherwise timing leaks user existence.
	dummyHash []byte
	Sessions  *repository.SessionRepo
	// LoginLimiter counts FAILED logins only: 5 per 15 minutes per IP,
	// then 429 (approved security note, reusing the Limiter mechanism).
	LoginLimiter *Limiter
}

// NewAdminAuth hashes password with bcrypt cost 12 (approved note).
// Callers must fail closed before this point when user/password are
// empty; this constructor returns an error rather than a usable auth
// object with an empty credential.
func NewAdminAuth(user, password string, sessions *repository.SessionRepo) (*AdminAuth, error) {
	if user == "" || password == "" {
		return nil, errEmptyAdminCredential
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return nil, err
	}
	dummy := make([]byte, 32)
	if _, err := rand.Read(dummy); err != nil {
		return nil, err
	}
	dummyHash, err := bcrypt.GenerateFromPassword(dummy, 12)
	if err != nil {
		return nil, err
	}
	return &AdminAuth{
		User:         user,
		passwordHash: hash,
		dummyHash:    dummyHash,
		Sessions:     sessions,
		LoginLimiter: NewLimiter(5, 15*time.Minute),
	}, nil
}

type adminCredentialError string

func (e adminCredentialError) Error() string { return string(e) }

const errEmptyAdminCredential adminCredentialError = "admin credentials must not be empty"

// CheckPassword reports whether password matches the stored hash.
// It is deliberately unexported: login handling (generic error, failure
// counting) lives in the admin handler, not in callers.
func (a *AdminAuth) CheckPassword(password string) bool {
	return bcrypt.CompareHashAndPassword(a.passwordHash, []byte(password)) == nil
}

// CheckDummy runs one bcrypt comparison against the dummy hash. It
// always returns false in practice (the dummy value is random and
// never presented as a password); its purpose is timing parity for
// unknown-user logins, not authentication.
func (a *AdminAuth) CheckDummy(password string) bool {
	return bcrypt.CompareHashAndPassword(a.dummyHash, []byte(password)) == nil
}

// IsUser reports whether username matches, without revealing anything
// else; the handler always returns the same generic error for a wrong
// user and a wrong password.
func (a *AdminAuth) IsUser(user string) bool { return user == a.User }

// NewSession creates a crypto/rand 32-byte token, stores only its
// SHA-256 hash with a 12-hour expiry, and returns the raw token for
// the cookie. The raw token is never stored or logged.
func (a *AdminAuth) NewSession() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	token := hex.EncodeToString(b)
	if err := a.Sessions.Create(repository.HashToken(token), time.Now().Add(SessionExpiry)); err != nil {
		return "", err
	}
	return token, nil
}

// ValidateSession reports whether a raw token maps to a live session.
func (a *AdminAuth) ValidateSession(token string) (bool, error) {
	return a.Sessions.Valid(repository.HashToken(token))
}

// InvalidateSession deletes a session by raw token (logout).
func (a *AdminAuth) InvalidateSession(token string) error {
	return a.Sessions.Delete(repository.HashToken(token))
}

// SetSessionCookie writes the session cookie: HttpOnly, SameSite=Lax,
// Path=/admin, Secure only when APP_ENV=production (approved note).
func SetSessionCookie(w http.ResponseWriter, token string, production bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     SessionCookie,
		Value:    token,
		Path:     "/admin",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   production,
		MaxAge:   int(SessionExpiry / time.Second),
	})
}

// ClearSessionCookie expires the session cookie.
func ClearSessionCookie(w http.ResponseWriter, production bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     SessionCookie,
		Value:    "",
		Path:     "/admin",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   production,
		MaxAge:   -1,
	})
}

// RequireAdmin is the chain's auth step. Every /admin/* request needs
// a valid session, except login GET/POST themselves. Unauthenticated
// access redirects (302) to /admin/login (approved note). A validated
// session hash is stored in the request context for CSRF validation.
func RequireAdmin(auth *AdminAuth) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !strings.HasPrefix(r.URL.Path, "/admin") {
				next.ServeHTTP(w, r)
				return
			}
			if r.URL.Path == "/admin/login" {
				next.ServeHTTP(w, r)
				return
			}
			c, err := r.Cookie(SessionCookie)
			if err != nil || c.Value == "" {
				http.Redirect(w, r, "/admin/login", http.StatusFound)
				return
			}
			ok, err := auth.ValidateSession(c.Value)
			if err != nil || !ok {
				http.Redirect(w, r, "/admin/login", http.StatusFound)
				return
			}
			ctx := context.WithValue(r.Context(), sessionHashKey, repository.HashToken(c.Value))
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
