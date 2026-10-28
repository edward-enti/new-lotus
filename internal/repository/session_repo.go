package repository

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"time"
)

// SessionRepo stores admin sessions server-side. Only the SHA-256 of a
// token is ever stored (approved security note); the raw token exists
// only in the client's cookie and in memory during login validation.
type SessionRepo struct{ DB *sql.DB }

// HashToken returns the hex SHA-256 of a raw session token.
func HashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

// Create inserts a session row with the given expiry.
func (r *SessionRepo) Create(tokenHash string, expiresAt time.Time) error {
	_, err := r.DB.Exec(`INSERT INTO admin_sessions(token_hash, expires_at) VALUES (?, ?)`, tokenHash, expiresAt.UTC())
	return err
}

// Valid reports whether a token hash exists and has not expired.
// Expired rows are deleted opportunistically by the caller (see
// DeleteExpired); a single expired row here is simply not valid.
func (r *SessionRepo) Valid(tokenHash string) (bool, error) {
	var expiresAt time.Time
	err := r.DB.QueryRow(`SELECT expires_at FROM admin_sessions WHERE token_hash = ?`, tokenHash).Scan(&expiresAt)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return time.Now().Before(expiresAt), nil
}

// Delete removes one session row (logout / rotation).
func (r *SessionRepo) Delete(tokenHash string) error {
	_, err := r.DB.Exec(`DELETE FROM admin_sessions WHERE token_hash = ?`, tokenHash)
	return err
}

// DeleteExpired removes all expired session rows.
func (r *SessionRepo) DeleteExpired() error {
	_, err := r.DB.Exec(`DELETE FROM admin_sessions WHERE expires_at <= ?`, time.Now().UTC())
	return err
}
