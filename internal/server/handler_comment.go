package server

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"path/filepath"

	"github.com/go-chi/chi/v5"
	"github.com/shimasan0x00/diffff/internal/comment"
)

type createCommentRequest struct {
	FilePath string `json:"filePath"`
	Line     int    `json:"line"`
	Body     string `json:"body"`
}

type updateCommentRequest struct {
	Body string `json:"body"`
}

const maxCommentBodySize = 1 << 20 // 1MB

func (s *Server) handleCreateComment(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxCommentBodySize)
	var req createCommentRequest
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
	if req.Body == "" {
		writeError(w, http.StatusBadRequest, "body is required")
		return
	}
	if req.Line < 1 {
		writeError(w, http.StatusBadRequest, "line must be a positive integer")
		return
	}
	if _, ok := s.fileIndex[req.FilePath]; !ok {
		writeError(w, http.StatusBadRequest, "filePath is not in the current diff")
		return
	}

	c, err := s.commentStore.Create(&comment.Comment{
		FilePath: req.FilePath,
		Line:     req.Line,
		Body:     req.Body,
	})
	if err != nil {
		slog.Error("create comment error", "err", err)
		writeError(w, http.StatusInternalServerError, "failed to create comment")
		return
	}

	writeJSON(w, http.StatusCreated, c)
}

func (s *Server) handleListComments(w http.ResponseWriter, r *http.Request) {
	filePath := r.URL.Query().Get("file")
	comments := s.commentStore.List(filePath)
	if comments == nil {
		comments = []*comment.Comment{}
	}
	writeJSON(w, http.StatusOK, comments)
}

func (s *Server) handleUpdateComment(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	r.Body = http.MaxBytesReader(w, r.Body, maxCommentBodySize)
	var req updateCommentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Body == "" {
		writeError(w, http.StatusBadRequest, "body is required")
		return
	}

	c, err := s.commentStore.Update(id, req.Body)
	if err != nil {
		if errors.Is(err, comment.ErrNotFound) {
			writeError(w, http.StatusNotFound, "not found")
			return
		}
		slog.Error("update comment error", "err", err)
		writeError(w, http.StatusInternalServerError, "failed to update comment")
		return
	}

	writeJSON(w, http.StatusOK, c)
}

func (s *Server) handleDeleteComment(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	err := s.commentStore.Delete(id)
	if err != nil {
		if errors.Is(err, comment.ErrNotFound) {
			writeError(w, http.StatusNotFound, "not found")
			return
		}
		slog.Error("delete comment error", "err", err)
		writeError(w, http.StatusInternalServerError, "failed to delete comment")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleExportComments(w http.ResponseWriter, r *http.Request) {
	format := r.URL.Query().Get("format")
	comments := s.commentStore.List("")

	switch format {
	case "json":
		jsonStr, err := comment.ExportJSON(comments)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to export comments")
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Content-Disposition", `attachment; filename="comments.json"`)
		if _, err := w.Write([]byte(jsonStr)); err != nil {
			slog.Error("export write error", "err", err)
		}
	default:
		w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
		w.Header().Set("Content-Disposition", `attachment; filename="comments.md"`)
		if _, err := w.Write([]byte(comment.ExportMarkdown(comments))); err != nil {
			slog.Error("export write error", "err", err)
		}
	}
}
