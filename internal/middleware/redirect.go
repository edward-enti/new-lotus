package middleware

import (
	"database/sql"
	"errors"
	"net/http"
	"strings"

	"sonlotus/internal/repository"
)

// maxRedirectHops caps server-side redirect-chain resolution (PLAN §6
// loop guard). A chain longer than this, or one that revisits a path,
// is not followed: the request passes through instead of issuing a
// looping redirect.
const maxRedirectHops = 5

// Redirect is the chain's redirect step. New routes win over redirect
// lookup: IsRealRoute reports whether a path is a real application
// route, and such paths pass through untouched even if a redirect row
// exists for them. Unmatched paths are looked up in url_redirects and
// resolved server-side: the chain is followed (visited-set, depth cap)
// to its final target and a SINGLE redirect to that target is issued
// with the first row's status_code (defaulting to 301 for values
// outside 301/302/307/308). hits increments for each row followed.
// Loop guard: a redirect whose new path revisits any visited path
// (including direct A->A), or a chain exceeding maxRedirectHops, is
// not followed — the request passes through untouched.
func Redirect(repo *repository.RedirectRepo, isRealRoute func(string) bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodGet || isRealRoute == nil || isRealRoute(r.URL.Path) {
				next.ServeHTTP(w, r)
				return
			}
			visited := map[string]bool{strings.TrimSuffix(r.URL.Path, "/"): true}
			current := r.URL.Path
			var code int
			hops := 0
			for {
				rd, err := repo.GetByOldPath(current)
				if err != nil {
					if errors.Is(err, sql.ErrNoRows) {
						if hops == 0 {
							next.ServeHTTP(w, r)
							return
						}
						// Chain broke at a non-redirect path: it is
						// the final target.
						break
					}
					next.ServeHTTP(w, r)
					return
				}
				target := strings.TrimSuffix(rd.NewPath, "/")
				if visited[target] {
					// Loop (direct A->A or longer A->B->A cycle).
					next.ServeHTTP(w, r)
					return
				}
				visited[target] = true
				hops++
				if hops == 1 {
					code = rd.StatusCode
				}
				_ = repo.IncrementHits(rd.ID)
				if hops > maxRedirectHops {
					// Cap exceeded: do not issue a redirect that
					// may still be mid-chain/looping.
					next.ServeHTTP(w, r)
					return
				}
				// Loop again: if the target is itself a redirect
				// row it resolves further; otherwise the top of
				// the loop treats it as the final target.
				current = rd.NewPath
			}
			if code != http.StatusMovedPermanently && code != http.StatusFound &&
				code != http.StatusTemporaryRedirect && code != http.StatusPermanentRedirect {
				code = http.StatusMovedPermanently
			}
			http.Redirect(w, r, current, code)
		})
	}
}
