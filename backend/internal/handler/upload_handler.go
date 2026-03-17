package handler

import (
	"crypto/rand"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"classified-listings/internal/response"
)

const maxUploadBytes = 5 << 20 // 5 MB

// allowedMIME maps the MIME types we accept to their canonical file extensions.
// Detection is done on the raw bytes (magic numbers), not the filename extension.
var allowedMIME = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/webp": ".webp",
	"image/gif":  ".gif",
}

// UploadHandler handles file uploads and serves them from a local directory.
type UploadHandler struct {
	uploadsDir string
}

func NewUploadHandler(uploadsDir string) *UploadHandler {
	return &UploadHandler{uploadsDir: uploadsDir}
}

// Upload handles POST /api/upload.
// Accepts a multipart form with a single "file" field.
// Validates MIME type and size, saves the file, and returns its public URL.
func (h *UploadHandler) Upload(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadBytes)
	if err := r.ParseMultipartForm(maxUploadBytes); err != nil {
		response.WriteError(w, http.StatusBadRequest, "file too large or malformed form (max 5 MB)")
		return
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, `field "file" is required`)
		return
	}
	defer file.Close()

	// Read the first 512 bytes to detect MIME type from magic bytes, not the
	// filename extension, which can be trivially spoofed.
	header := make([]byte, 512)
	n, err := file.Read(header)
	if err != nil && err != io.EOF {
		response.WriteError(w, http.StatusBadRequest, "could not read file")
		return
	}

	mimeType := http.DetectContentType(header[:n])
	// Strip any parameters (e.g. "image/jpeg; charset=utf-8").
	if idx := strings.Index(mimeType, ";"); idx != -1 {
		mimeType = strings.TrimSpace(mimeType[:idx])
	}

	ext, ok := allowedMIME[mimeType]
	if !ok {
		response.WriteError(w, http.StatusBadRequest, "only JPEG, PNG, WebP, and GIF images are accepted")
		return
	}

	filename := randomFilename() + ext
	dst := filepath.Join(h.uploadsDir, filename)

	out, err := os.Create(dst)
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, "could not save file")
		return
	}
	defer out.Close()

	// Write the already-read header bytes, then stream the remainder.
	if _, err = out.Write(header[:n]); err != nil {
		response.WriteError(w, http.StatusInternalServerError, "could not save file")
		return
	}
	if _, err = io.Copy(out, file); err != nil {
		response.WriteError(w, http.StatusInternalServerError, "could not save file")
		return
	}

	response.WriteJSON(w, http.StatusOK, map[string]string{
		"url": "/uploads/" + filename,
	})
}

// randomFilename generates a random hex string suitable for use as a filename.
// Using crypto/rand avoids collisions without requiring an external UUID package.
func randomFilename() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return fmt.Sprintf("%x", b)
}
