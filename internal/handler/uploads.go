package handler

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"sonlotus/internal/middleware"
	"sonlotus/internal/service"
)

// MaxUploadBytes is the approved 10 MB per-file upload cap.
const MaxUploadBytes = 10 << 20

// Upload handles POST /admin/upload. Admin session + CSRF are required
// (session via the chain; CSRF validated here because a multipart body
// must be size-capped with http.MaxBytesReader BEFORE it is parsed —
// the chain's CSRF step deliberately does not parse multipart bodies).
// Only images + PDFs are allowed; the file is saved under UploadDir
// with a sanitised basename, and a name collision gets a timestamp
// suffix instead of overwriting an existing file.
func (h *AdminHandlers) Upload(w http.ResponseWriter, r *http.Request) {
	// Cap the whole request body first (file + multipart overhead).
	r.Body = http.MaxBytesReader(w, r.Body, MaxUploadBytes+(1<<20))
	if err := r.ParseMultipartForm(MaxUploadBytes); err != nil {
		http.Error(w, "file too large", http.StatusRequestEntityTooLarge)
		return
	}
	// CSRF (multipart path): same store, same constant-time compare.
	hash, ok := middleware.SessionHashFromContext(r.Context())
	if !ok || !h.CSRF.Validate(middleware.SessionKey(hash), r.FormValue(middleware.CSRFField)) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "no file uploaded", http.StatusBadRequest)
		return
	}
	defer file.Close()
	ext := strings.ToLower(filepath.Ext(header.Filename))
	if !service.IsAllowedUploadExtension(ext) {
		http.Error(w, "file type not allowed", http.StatusBadRequest)
		return
	}
	name := service.SanitizeFilename(header.Filename)
	if err := os.MkdirAll(h.UploadDir, 0o755); err != nil {
		http.Error(w, "upload error", http.StatusInternalServerError)
		return
	}
	dest := filepath.Join(h.UploadDir, name)
	if _, err := os.Stat(dest); err == nil {
		stem := strings.TrimSuffix(name, ext)
		dest = filepath.Join(h.UploadDir, fmt.Sprintf("%s-%d%s", stem, time.Now().Unix(), ext))
		name = filepath.Base(dest)
	}
	out, err := os.Create(dest)
	if err != nil {
		http.Error(w, "upload error", http.StatusInternalServerError)
		return
	}
	// Copy with a hard byte cap: the multipart header size is
	// client-supplied and is not trusted on its own.
	written, err := io.Copy(out, io.LimitReader(file, MaxUploadBytes+1))
	_ = out.Close()
	if err != nil {
		_ = os.Remove(dest)
		http.Error(w, "upload error", http.StatusInternalServerError)
		return
	}
	if written > MaxUploadBytes {
		_ = os.Remove(dest)
		http.Error(w, "file too large", http.StatusRequestEntityTooLarge)
		return
	}
	if written == 0 {
		_ = os.Remove(dest)
		http.Error(w, "empty file", http.StatusBadRequest)
		return
	}
	_ = name
	http.Redirect(w, r, "/admin/products", http.StatusFound)
}

var uploadContentTypes = map[string]string{
	".jpg":  "image/jpeg",
	".jpeg": "image/jpeg",
	".png":  "image/png",
	".webp": "image/webp",
	".gif":  "image/gif",
	".pdf":  "application/pdf",
}

// UploadsServe serves /uploads/{file} publicly. Only allowlisted
// extensions are servable: a non-allowlisted extension (including a
// manually placed .html/.svg/.js) returns 404, so uploaded content can
// never be served as executable same-origin content. No directory
// listing: a directory path or a missing file is 404. Responses carry
// X-Content-Type-Options: nosniff and Content-Disposition: inline
// with the sanitised filename.
func (h *AdminHandlers) UploadsServe(w http.ResponseWriter, r *http.Request) {
	file := r.PathValue("file")
	if file == "" || strings.Contains(file, "/") || service.SanitizeFilename(file) != file {
		http.NotFound(w, r)
		return
	}
	ext := strings.ToLower(filepath.Ext(file))
	ct, ok := uploadContentTypes[ext]
	if !ok || !service.IsAllowedUploadExtension(ext) {
		http.NotFound(w, r)
		return
	}
	full := filepath.Join(h.UploadDir, file)
	f, err := os.Open(full)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil || info.IsDir() {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Content-Type", ct)
	w.Header().Set("Content-Disposition", `inline; filename="`+file+`"`)
	http.ServeContent(w, r, file, info.ModTime(), f)
}
