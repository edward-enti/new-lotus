// Package config loads runtime settings from environment variables.
// Admin credentials intentionally have NO defaults: a missing or empty
// ADMIN_USER / ADMIN_PASSWORD must fail closed in the server and login
// handler (Phase 3 security note). This package only carries the values.
package config

import (
	"os"
	"strings"
)

type Config struct {
	Addr          string
	DBPath        string
	DataDir       string
	AppEnv        string
	AdminUser     string
	AdminPassword string
	// SiteURL is the public base URL (SITE_URL) used for absolute
	// sitemap <loc>, canonical URLs, JSON-LD urls and the robots.txt
	// Sitemap line. Empty means local dev: those values stay relative
	// (the pre-existing fallback); the server does not fail closed on
	// a missing SITE_URL.
	SiteURL string
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

// Load returns a Config with sensible defaults for everything except
// admin credentials, which stay empty unless explicitly set.
func Load() Config {
	return Config{
		Addr:          getenv("PORT", ":8080"),
		DBPath:        getenv("DB_PATH", "data/sonlotus.db"),
		DataDir:       getenv("DATA_DIR", "data"),
		AppEnv:        getenv("APP_ENV", "development"),
		AdminUser:     os.Getenv("ADMIN_USER"),
		AdminPassword: os.Getenv("ADMIN_PASSWORD"),
		SiteURL:       strings.TrimSuffix(os.Getenv("SITE_URL"), "/"),
	}
}

// AdminConfigured reports whether both admin env vars are non-empty.
func (c Config) AdminConfigured() bool {
	return c.AdminUser != "" && c.AdminPassword != ""
}
