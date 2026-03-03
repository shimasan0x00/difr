package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupReviewedServer(t *testing.T) *Server {
	t.Helper()
	dir := t.TempDir()
	srv, err := New(testDiff, WithWorkDir(dir), WithNoClaude(true))
	require.NoError(t, err)
	return srv
}

func TestListReviewedFiles_EmptyByDefault(t *testing.T) {
	s := setupReviewedServer(t)

	req := httptest.NewRequest(http.MethodGet, "/api/reviewed-files", nil)
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp struct {
		Files []string `json:"files"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Empty(t, resp.Files)
}

func TestToggleReviewedFile_AddsFile(t *testing.T) {
	s := setupReviewedServer(t)

	payload := `{"filePath":"main.go"}`
	req := httptest.NewRequest(http.MethodPost, "/api/reviewed-files", bytes.NewBufferString(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp struct {
		Files    []string `json:"files"`
		Reviewed bool     `json:"reviewed"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.True(t, resp.Reviewed)
	assert.Contains(t, resp.Files, "main.go")
}

func TestToggleReviewedFile_RemovesFile(t *testing.T) {
	s := setupReviewedServer(t)

	// Add first
	payload := `{"filePath":"main.go"}`
	req := httptest.NewRequest(http.MethodPost, "/api/reviewed-files", bytes.NewBufferString(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	// Toggle again to remove
	req = httptest.NewRequest(http.MethodPost, "/api/reviewed-files", bytes.NewBufferString(payload))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	s.Handler().ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp struct {
		Files    []string `json:"files"`
		Reviewed bool     `json:"reviewed"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.False(t, resp.Reviewed)
	assert.NotContains(t, resp.Files, "main.go")
}

func TestToggleReviewedFile_RejectsEmptyFilePath(t *testing.T) {
	s := setupReviewedServer(t)

	payload := `{"filePath":""}`
	req := httptest.NewRequest(http.MethodPost, "/api/reviewed-files", bytes.NewBufferString(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestToggleReviewedFile_RejectsPathTraversal(t *testing.T) {
	s := setupReviewedServer(t)

	payload := `{"filePath":"../../../etc/passwd"}`
	req := httptest.NewRequest(http.MethodPost, "/api/reviewed-files", bytes.NewBufferString(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestClearReviewedFiles(t *testing.T) {
	s := setupReviewedServer(t)

	// Add some files
	for _, f := range []string{"main.go", "a.go"} {
		payload, _ := json.Marshal(map[string]string{"filePath": f})
		req := httptest.NewRequest(http.MethodPost, "/api/reviewed-files", bytes.NewBuffer(payload))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		s.Handler().ServeHTTP(w, req)
		require.Equal(t, http.StatusOK, w.Code)
	}

	// Clear all
	req := httptest.NewRequest(http.MethodDelete, "/api/reviewed-files", nil)
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, req)
	assert.Equal(t, http.StatusNoContent, w.Code)

	// Verify empty
	req = httptest.NewRequest(http.MethodGet, "/api/reviewed-files", nil)
	w = httptest.NewRecorder()
	s.Handler().ServeHTTP(w, req)

	var resp struct {
		Files []string `json:"files"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Empty(t, resp.Files)
}
