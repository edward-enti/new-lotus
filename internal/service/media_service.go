package service

import (
	"path/filepath"
	"strings"
)

// MediaService holds upload filename/extension helpers for Phase 2 public
// work; the upload endpoint itself is Phase 3 (admin). These pure helpers
// are reused there, so they live here now and are tested here, not later.

// AllowedImageExtensions are the image extensions servable from /uploads.
var AllowedImageExtensions = []string{".jpg", ".jpeg", ".png", ".webp", ".gif"}

// AllowedUploadExtensions are images + PDFs only (PLAN §6).
func AllowedUploadExtensions() []string {
	return append(append([]string{}, AllowedImageExtensions...), ".pdf")
}

// SanitizeFilename reduces a user-supplied filename to a safe basename:
// no directories, no "..", only letters, digits, dot, dash, underscore.
func SanitizeFilename(name string) string {
	base := filepath.Base(name)
	var b strings.Builder
	for _, r := range base {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '.' || r == '-' || r == '_' {
			b.WriteRune(r)
		} else {
			b.WriteRune('_')
		}
	}
	out := strings.Trim(b.String(), ".")
	if out == "" {
		return "file"
	}
	return out
}

// WebPURL derives the WebP sibling URL for an image URL by swapping
// its extension (.jpg/.jpeg/.png -> .webp); other extensions and
// already-WebP URLs pass through unchanged. This is the <picture>
// fallback derivation for static/seed images (PLAN §7 WebP default).
// It does NOT generate a WebP file: upload-time WebP conversion is NOT
// implemented here — Go's standard library has no WebP encoder and a
// CGO encoder would break this project's pure-Go guardrail, so that
// conversion is flagged to the Supervisor instead of invented (see the
// Phase 3–4 report). Use webpURL in templates only for images whose
// WebP sibling is actually supplied alongside the original.
func WebPURL(url string) string {
	ext := strings.ToLower(filepath.Ext(url))
	switch ext {
	case ".jpg", ".jpeg", ".png":
		return strings.TrimSuffix(url, filepath.Ext(url)) + ".webp"
	default:
		return url
	}
}

// IsAllowedUploadExtension reports whether ext (with leading dot) is allowed.
func IsAllowedUploadExtension(ext string) bool {
	ext = strings.ToLower(ext)
	for _, ok := range AllowedUploadExtensions() {
		if ext == ok {
			return true
		}
	}
	return false
}
