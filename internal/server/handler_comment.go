package server

import (
	"encoding/json"
	"errors"
	"log/slog"
	"mime"
	"net/http"
	"path/filepath"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/shimasan0x00/difr/internal/comment"
)

// requireJSONContentType validates that the request has a JSON content type.
func requireJSONContentType(w http.ResponseWriter, r *http.Request) bool {
	ct := r.Header.Get("Content-Type")
	if ct == "" {
		writeError(w, http.StatusUnsupportedMediaType, "Content-Type header is required")
		return false
	}
	mediaType, _, err := mime.ParseMediaType(ct)
	if err != nil || mediaType != "application/json" {
		writeError(w, http.StatusUnsupportedMediaType, "Content-Type must be application/json")
		return false
	}
	return true
}

type createCommentRequest struct {
	FilePath       string `json:"filePath"`
	Line           int    `json:"line"`
	Body           string `json:"body"`
	ReviewCategory string `json:"reviewCategory"`
	Severity       string `json:"severity"`
}

type updateCommentRequest struct {
	Body           string `json:"body"`
	ReviewCategory string `json:"reviewCategory"`
	Severity       string `json:"severity"`
}

const maxCommentBodySize = 1 << 20 // 1MB

func (s *Server) handleCreateComment(w http.ResponseWriter, r *http.Request) {
	if !requireJSONContentType(w, r) {
		return
	}
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
	if req.Line < 0 {
		writeError(w, http.StatusBadRequest, "line must be a non-negative integer")
		return
	}
	if _, ok := s.fileIndex[req.FilePath]; !ok {
		writeError(w, http.StatusBadRequest, "filePath is not in the current diff")
		return
	}
	if !comment.ValidateCategory(req.ReviewCategory) {
		writeError(w, http.StatusBadRequest, "invalid reviewCategory")
		return
	}
	if !comment.ValidateSeverity(req.Severity) {
		writeError(w, http.StatusBadRequest, "invalid severity")
		return
	}

	c, err := s.commentStore.Create(&comment.Comment{
		FilePath:       req.FilePath,
		Line:           req.Line,
		Body:           req.Body,
		ReviewCategory: req.ReviewCategory,
		Severity:       req.Severity,
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

	// Optional pagination: ?limit=N&offset=N
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err != nil || limit < 1 {
			writeError(w, http.StatusBadRequest, "limit must be a positive integer")
			return
		}
		offset := 0
		if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
			offset, err = strconv.Atoi(offsetStr)
			if err != nil || offset < 0 {
				writeError(w, http.StatusBadRequest, "offset must be a non-negative integer")
				return
			}
		}
		if offset >= len(comments) {
			comments = []*comment.Comment{}
		} else {
			end := offset + limit
			if end > len(comments) {
				end = len(comments)
			}
			comments = comments[offset:end]
		}
	}

	writeJSON(w, http.StatusOK, comments)
}

func (s *Server) handleUpdateComment(w http.ResponseWriter, r *http.Request) {
	if !requireJSONContentType(w, r) {
		return
	}
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
	if !comment.ValidateCategory(req.ReviewCategory) {
		writeError(w, http.StatusBadRequest, "invalid reviewCategory")
		return
	}
	if !comment.ValidateSeverity(req.Severity) {
		writeError(w, http.StatusBadRequest, "invalid severity")
		return
	}

	c, err := s.commentStore.Update(id, comment.UpdateFields{
		Body:           req.Body,
		ReviewCategory: req.ReviewCategory,
		Severity:       req.Severity,
	})
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

func (s *Server) handleDeleteAllComments(w http.ResponseWriter, r *http.Request) {
	if err := s.commentStore.DeleteAll(); err != nil {
		slog.Error("delete all comments error", "err", err)
		writeError(w, http.StatusInternalServerError, "failed to delete all comments")
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
	case "csv":
		w.Header().Set("Content-Type", "text/csv; charset=utf-8")
		w.Header().Set("Content-Disposition", `attachment; filename="comments.csv"`)
		if _, err := w.Write([]byte(comment.ExportCSV(comments))); err != nil {
			slog.Error("export write error", "err", err)
		}
	case "xlsx":
		data, err := comment.ExportExcel(comments)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to export comments")
			return
		}
		w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
		w.Header().Set("Content-Disposition", `attachment; filename="comments.xlsx"`)
		if _, err := w.Write(data); err != nil {
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
