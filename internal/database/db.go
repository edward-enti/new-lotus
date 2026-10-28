// Package database owns the SQLite connection, pragmas, migrations,
// FTS backfill and the shared seed entry point.
package database

import (
	"database/sql"
	"embed"
	"fmt"
	"strings"

	_ "modernc.org/sqlite"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// migrationFiles run in filename order.
var migrationFiles = []string{
	"migrations/001_initial_schema.sql",
	"migrations/002_fts_search.sql",
	"migrations/003_admin_sessions.sql",
}

// Open opens a SQLite database at path with the SPEC §2 pragmas applied
// on EVERY connection via DSN pragmas (these pragmas are per-connection
// in SQLite, so listing them once elsewhere is not enough).
// Single-writer policy (PLAN §1): MaxOpenConns(1), serialising all writes.
func Open(path string) (*sql.DB, error) {
	dsn := fmt.Sprintf("file:%s?_pragma=journal_mode(WAL)&_pragma=synchronous(NORMAL)&_pragma=busy_timeout(5000)&_pragma=foreign_keys(ON)", path)
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}

// splitSQL splits a migration script on semicolons that are outside
// single-quoted strings, outside -- line comments, and outside
// BEGIN...END trigger bodies (whose inner statements end with ';' too).
// It is deliberately small: migration files are ours, not arbitrary SQL.
func splitSQL(script string) []string {
	var out []string
	var cur strings.Builder
	inStr := false
	depth := 0
	i := 0
	isWord := func(c byte) bool {
		return c == '_' || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')
	}
	for i < len(script) {
		c := script[i]
		if inStr {
			cur.WriteByte(c)
			if c == '\'' {
				// '' is an escaped quote inside a string literal.
				if i+1 < len(script) && script[i+1] == '\'' {
					cur.WriteByte(script[i+1])
					i += 2
					continue
				}
				inStr = false
			}
			i++
			continue
		}
		if c == '\'' {
			inStr = true
			cur.WriteByte(c)
			i++
			continue
		}
		if c == '-' && i+1 < len(script) && script[i+1] == '-' {
			// Copy the comment through to end of line.
			for i < len(script) && script[i] != '\n' {
				cur.WriteByte(script[i])
				i++
			}
			continue
		}
		if isWord(c) {
			start := i
			for i < len(script) && isWord(script[i]) {
				i++
			}
			word := script[start:i]
			cur.WriteString(word)
			switch strings.ToUpper(word) {
			case "BEGIN":
				depth++
			case "END":
				if depth > 0 {
					depth--
				}
			}
			continue
		}
		if c == ';' && depth == 0 {
			stmt := strings.TrimSpace(cur.String())
			if stmt != "" {
				out = append(out, stmt)
			}
			cur.Reset()
			i++
			continue
		}
		cur.WriteByte(c)
		i++
	}
	if stmt := strings.TrimSpace(cur.String()); stmt != "" {
		out = append(out, stmt)
	}
	return out
}

// Migrate runs the embedded migrations in order. Files are idempotent
// (IF NOT EXISTS), so re-running Migrate is safe.
func Migrate(db *sql.DB) error {
	for _, name := range migrationFiles {
		raw, err := migrationsFS.ReadFile(name)
		if err != nil {
			return fmt.Errorf("database: read migration %s: %w", name, err)
		}
		for _, stmt := range splitSQL(string(raw)) {
			if _, err := db.Exec(stmt); err != nil {
				return fmt.Errorf("database: migration %s statement %q: %w", name, stmt, err)
			}
		}
	}
	return nil
}

// BackfillFTS rebuilds search_index from products and articles. Run after
// every seed and every migration so the index cannot go stale when a
// write bypasses the repository layer (PLAN §3).
func BackfillFTS(db *sql.DB) error {
	if _, err := db.Exec(`DELETE FROM search_index`); err != nil {
		return err
	}
	if _, err := db.Exec(`INSERT INTO search_index(entity_type, entity_id, title, slug, thumbnail_url, summary, body_text)
		SELECT 'product', id, name, slug, thumbnail_url, short_summary, description_html FROM products`); err != nil {
		return err
	}
	if _, err := db.Exec(`INSERT INTO search_index(entity_type, entity_id, title, slug, thumbnail_url, summary, body_text)
		SELECT 'article', id, title, slug, thumbnail_url, excerpt, content_html FROM articles`); err != nil {
		return err
	}
	return nil
}

// FTSQuery converts raw user input into a safe FTS5 MATCH expression:
// each whitespace-separated token is cleaned of FTS operators and wrapped
// in double quotes, so tokens are ANDed as single-token phrases.
// Parameterised SQL alone does not cover FTS MATCH injection (PLAN §1).
// An empty or all-operator input returns "" (caller must not query then).
func FTSQuery(q string) string {
	q = strings.TrimSpace(q)
	if len(q) > 200 {
		q = q[:200]
	}
	var phrases []string
	for _, tok := range strings.Fields(q) {
		clean := strings.Map(func(r rune) rune {
			switch r {
			case '"', '*', '(', ')', ':', '^', '{', '}', '[', ']':
				return -1
			}
			return r
		}, tok)
		clean = strings.Trim(clean, "'")
		if clean == "" {
			continue
		}
		// Reject FTS boolean keywords as bare tokens; they are treated
		// as ordinary words only when quoted, which we already do, so
		// no special-casing is needed beyond the quoting above.
		phrases = append(phrases, `"`+clean+`"`)
	}
	return strings.Join(phrases, " ")
}
