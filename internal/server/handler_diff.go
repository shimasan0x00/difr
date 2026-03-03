package server

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/shimasan0x00/difr/internal/diff"
)

func (s *Server) handleGetDiff(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, s.diffResult)
}

func (s *Server) handleGetDiffFiles(w http.ResponseWriter, r *http.Request) {
	files := s.diffResult.Files
	if files == nil {
		files = []diff.DiffFile{}
	}
	writeJSON(w, http.StatusOK, files)
}

func (s *Server) handleGetDiffFileByPath(w http.ResponseWriter, r *http.Request) {
	path := chi.URLParam(r, "*")
	if path == "" {
		writeError(w, http.StatusBadRequest, "path required")
		return
	}

	if f, ok := s.fileIndex[path]; ok {
		writeJSON(w, http.StatusOK, f)
		return
	}

	writeError(w, http.StatusNotFound, "file not found")
}

func (s *Server) handleGetDiffStats(w http.ResponseWriter, r *http.Request) {
	resp := struct {
		Files int            `json:"files"`
		Stats diff.FileStats `json:"stats"`
	}{
		Files: len(s.diffResult.Files),
		Stats: s.diffResult.Stats,
	}
	writeJSON(w, http.StatusOK, resp)
}

func (s *Server) handleGetDiffMeta(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, s.diffMeta)
}

func (s *Server) handleGetTrackedFiles(w http.ResponseWriter, r *http.Request) {
	files := s.trackedFiles
	if files == nil {
		files = []string{}
	}
	writeJSON(w, http.StatusOK, map[string][]string{"files": files})
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	data, err := json.Marshal(v)
	if err != nil {
		slog.Error("writeJSON encode error", "err", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"internal server error"}`))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if _, err := w.Write(data); err != nil {
		slog.Error("writeJSON write error", "err", err)
	}
}
