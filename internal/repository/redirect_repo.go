package repository

import "database/sql"

// Redirect is a legacy URL redirect row (SPEC §4.1). It lives in the
// repository package; the domain package has no redirect struct because
// SPEC §3 lists domain structs only for product/article/inquiry/seo.
type Redirect struct {
	ID         int64
	OldPath    string
	NewPath    string
	StatusCode int
	Hits       int
}

// RedirectRepo reads legacy redirects.
type RedirectRepo struct{ DB *sql.DB }

// GetByOldPath returns a redirect, or sql.ErrNoRows if none exists.
func (r *RedirectRepo) GetByOldPath(oldPath string) (Redirect, error) {
	var rd Redirect
	err := r.DB.QueryRow(`SELECT id, old_path, new_path, status_code, hits FROM url_redirects WHERE old_path = ?`, oldPath).
		Scan(&rd.ID, &rd.OldPath, &rd.NewPath, &rd.StatusCode, &rd.Hits)
	return rd, err
}

// IncrementHits bumps the hit counter for a redirect id.
func (r *RedirectRepo) IncrementHits(id int64) error {
	_, err := r.DB.Exec(`UPDATE url_redirects SET hits = hits + 1 WHERE id = ?`, id)
	return err
}

// Top returns redirects ordered by hits, most-hit first.
func (r *RedirectRepo) Top(limit int) ([]Redirect, error) {
	rows, err := r.DB.Query(`SELECT id, old_path, new_path, status_code, hits FROM url_redirects ORDER BY hits DESC, id ASC LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Redirect
	for rows.Next() {
		var rd Redirect
		if err := rows.Scan(&rd.ID, &rd.OldPath, &rd.NewPath, &rd.StatusCode, &rd.Hits); err != nil {
			return nil, err
		}
		out = append(out, rd)
	}
	return out, rows.Err()
}
