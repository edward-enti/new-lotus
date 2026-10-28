package middleware

import (
	"net"
	"net/http"
	"sync"
	"time"
)

// Limiter is a per-IP sliding-window limiter. It is the SAME mechanism
// as the Phase 2 inquiry endpoint limiter (same algorithm, same window
// semantics); the login failure rule reuses it rather than inventing a
// separate progressive-delay system (approved security note).
type Limiter struct {
	mu     sync.Mutex
	hits   map[string][]time.Time
	limit  int
	window time.Duration
}

// NewLimiter returns a limiter allowing limit events per window per IP.
func NewLimiter(limit int, window time.Duration) *Limiter {
	return &Limiter{hits: map[string][]time.Time{}, limit: limit, window: window}
}

func (l *Limiter) prune(ip string, now time.Time) []time.Time {
	cutoff := now.Add(-l.window)
	var kept []time.Time
	for _, t := range l.hits[ip] {
		if t.After(cutoff) {
			kept = append(kept, t)
		}
	}
	l.hits[ip] = kept
	return kept
}

// Allow records an event and reports whether it is within the limit.
func (l *Limiter) Allow(ip string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	kept := l.prune(ip, time.Now())
	if len(kept) >= l.limit {
		return false
	}
	l.hits[ip] = append(kept, time.Now())
	return true
}

// Blocked reports whether ip has already hit the limit, without
// recording a new event. Used for failure-counting limiters (login):
// only failures are recorded, via Record.
func (l *Limiter) Blocked(ip string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	return len(l.prune(ip, time.Now())) >= l.limit
}

// Record adds one event for ip (a failure, for login limiting).
func (l *Limiter) Record(ip string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.hits[ip] = append(l.prune(ip, time.Now()), time.Now())
}

// ClientIP returns the client IP from RemoteAddr only — never an
// untrusted X-Forwarded-For header (PLAN §4).
func ClientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

// RateLimit is the chain's rate-limit step. It applies ONLY to the
// public inquiry POST, with a coarse abuse guard; it deliberately
// does NOT apply to /admin/login, so inquiry spam from one IP can
// never lock out admin login from that IP (separate buckets by
// design, not a shared bucket). The security note's
// 5-failed-logins/15min rule is enforced separately in the login
// handler, counting failures only. Other routes pass through untouched.
func RateLimit(l *Limiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if l == nil {
				next.ServeHTTP(w, r)
				return
			}
			if r.Method == http.MethodPost && r.URL.Path == "/lien-he" {
				if !l.Allow(ClientIP(r)) {
					http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
					return
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}
