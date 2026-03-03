package server

import (
	"bytes"
	"net/http"
	"os"
	"path/filepath"
	"unicode/utf8"

	"github.com/go-chi/chi/v5"
)

const maxFileSize = 5 * 1024 * 1024 // 5MB

type fileContentResponse struct {
	Path        string `json:"path"`
	Content     string `json:"content"`
	IsBinary    bool   `json:"isBinary,omitempty"`
	IsTruncated bool   `json:"isTruncated,omitempty"`
	Size        int64  `json:"size"`
}

func (s *Server) handleGetFileContent(w http.ResponseWriter, r *http.Request) {
	relPath := chi.URLParam(r, "*")
	if relPath == "" {
		writeError(w, http.StatusBadRequest, "path required")
		return
	}

	// Security: reject path traversal
	if !filepath.IsLocal(relPath) {
		writeError(w, http.StatusBadRequest, "invalid path")
		return
	}

	// Security: whitelist check — only git-tracked files
	if _, ok := s.trackedIndex[relPath]; !ok {
		writeError(w, http.StatusNotFound, "file not found")
		return
	}

	absPath := filepath.Join(s.workDir, relPath)

	// Security: resolve symlinks and verify the resolved path stays within workDir
	resolved, err := filepath.EvalSymlinks(absPath)
	if err != nil {
		writeError(w, http.StatusNotFound, "file not found")
		return
	}
	if !isSubPath(s.resolvedWorkDir, resolved) {
		writeError(w, http.StatusBadRequest, "invalid path")
		return
	}

	info, err := os.Stat(resolved)
	if err != nil {
		writeError(w, http.StatusNotFound, "file not found")
		return
	}

	size := info.Size()

	// Truncated: file exceeds size limit
	if size > maxFileSize {
		writeJSON(w, http.StatusOK, fileContentResponse{
			Path:        relPath,
			Content:     "",
			IsTruncated: true,
			Size:        size,
		})
		return
	}

	data, err := os.ReadFile(resolved)
	if err != nil {
		writeError(w, http.StatusNotFound, "file not found")
		return
	}

	// Binary detection: null bytes or invalid UTF-8
	if isBinaryContent(data) {
		writeJSON(w, http.StatusOK, fileContentResponse{
			Path:     relPath,
			Content:  "",
			IsBinary: true,
			Size:     size,
		})
		return
	}

	writeJSON(w, http.StatusOK, fileContentResponse{
		Path:    relPath,
		Content: string(data),
		Size:    size,
	})
}

// isBinaryContent detects binary content by checking for null bytes or invalid UTF-8.
func isBinaryContent(data []byte) bool {
	if bytes.ContainsRune(data, 0) {
		return true
	}
	return !utf8.Valid(data)
}

// isSubPath checks if child is inside parent directory.
func isSubPath(parent, child string) bool {
	rel, err := filepath.Rel(parent, child)
	if err != nil {
		return false
	}
	return filepath.IsLocal(rel)
}
