package server

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"path/filepath"
)

type toggleReviewedRequest struct {
	FilePath string `json:"filePath"`
}

func (s *Server) handleListReviewedFiles(w http.ResponseWriter, r *http.Request) {
	files := s.reviewedStore.List()
	writeJSON(w, http.StatusOK, map[string][]string{"files": files})
}

func (s *Server) handleToggleReviewedFile(w http.ResponseWriter, r *http.Request) {
	if !requireJSONContentType(w, r) {
		return
	}

	var req toggleReviewedRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.FilePath == "" {
		writeError(w, http.StatusBadRequest, "filePath is required")
		return
	}
	if !filepath.IsLocal(req.FilePath) {
		writeError(w, http.StatusBadRequest, "filePath must be a relative path within the project")
		return
	}

	var isReviewed bool
	if s.reviewedStore.Has(req.FilePath) {
		if err := s.reviewedStore.Remove(req.FilePath); err != nil {
			slog.Error("remove reviewed file error", "err", err)
			writeError(w, http.StatusInternalServerError, "failed to update reviewed status")
			return
		}
		isReviewed = false
	} else {
		if err := s.reviewedStore.Add(req.FilePath); err != nil {
			slog.Error("add reviewed file error", "err", err)
			writeError(w, http.StatusInternalServerError, "failed to update reviewed status")
			return
		}
		isReviewed = true
	}

	files := s.reviewedStore.List()
	writeJSON(w, http.StatusOK, map[string]any{"files": files, "reviewed": isReviewed})
}

func (s *Server) handleClearReviewedFiles(w http.ResponseWriter, r *http.Request) {
	if err := s.reviewedStore.Clear(); err != nil {
		slog.Error("clear reviewed files error", "err", err)
		writeError(w, http.StatusInternalServerError, "failed to clear reviewed files")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
