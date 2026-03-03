package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestServerWithFiles(t *testing.T, files map[string]string) *Server {
	t.Helper()
	dir := t.TempDir()

	tracked := make([]string, 0, len(files))
	for relPath, content := range files {
		absPath := filepath.Join(dir, relPath)
		require.NoError(t, os.MkdirAll(filepath.Dir(absPath), 0o755))
		require.NoError(t, os.WriteFile(absPath, []byte(content), 0o644))
		tracked = append(tracked, relPath)
	}

	srv, err := New("", WithWorkDir(dir), WithNoClaude(true), WithTrackedFiles(tracked))
	require.NoError(t, err)
	return srv
}

func TestHandleGetFileContent_ReturnsTrackedFile(t *testing.T) {
	srv := newTestServerWithFiles(t, map[string]string{
		"main.go": "package main\n",
	})

	req := httptest.NewRequest(http.MethodGet, "/api/files/main.go", nil)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp fileContentResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, "main.go", resp.Path)
	assert.Equal(t, "package main\n", resp.Content)
	assert.False(t, resp.IsBinary)
	assert.False(t, resp.IsTruncated)
	assert.Equal(t, int64(13), resp.Size)
}

func TestHandleGetFileContent_NestedPath(t *testing.T) {
	srv := newTestServerWithFiles(t, map[string]string{
		"src/pkg/main.go": "package pkg\n",
	})

	req := httptest.NewRequest(http.MethodGet, "/api/files/src/pkg/main.go", nil)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp fileContentResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, "src/pkg/main.go", resp.Path)
	assert.Equal(t, "package pkg\n", resp.Content)
}

func TestHandleGetFileContent_UntrackedFile(t *testing.T) {
	dir := t.TempDir()
	// File exists on disk but is not tracked
	require.NoError(t, os.WriteFile(filepath.Join(dir, "secret.txt"), []byte("secret"), 0o644))

	srv, err := New("", WithWorkDir(dir), WithNoClaude(true), WithTrackedFiles([]string{"main.go"}))
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/api/files/secret.txt", nil)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestHandleGetFileContent_NonexistentFile(t *testing.T) {
	srv := newTestServerWithFiles(t, map[string]string{
		"main.go": "package main\n",
	})

	req := httptest.NewRequest(http.MethodGet, "/api/files/nonexistent.go", nil)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestHandleGetFileContent_EmptyPath(t *testing.T) {
	srv := newTestServerWithFiles(t, map[string]string{
		"main.go": "package main\n",
	})

	req := httptest.NewRequest(http.MethodGet, "/api/files/", nil)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestHandleGetFileContent_PathTraversal(t *testing.T) {
	srv := newTestServerWithFiles(t, map[string]string{
		"main.go": "package main\n",
	})

	tests := []struct {
		path       string
		wantStatus int
	}{
		{"/api/files/../../../etc/passwd", http.StatusBadRequest},
		{"/api/files/src/../../etc/passwd", http.StatusBadRequest},
		// URL-encoded traversal is rejected by trackedIndex whitelist (404)
		{"/api/files/..%2F..%2Fetc%2Fpasswd", http.StatusNotFound},
	}
	for _, tt := range tests {
		req := httptest.NewRequest(http.MethodGet, tt.path, nil)
		rec := httptest.NewRecorder()
		srv.Handler().ServeHTTP(rec, req)

		assert.Equal(t, tt.wantStatus, rec.Code, "path: %s", tt.path)
	}
}

func TestHandleGetFileContent_BinaryFile(t *testing.T) {
	// Create binary content with null bytes
	binaryContent := []byte{0x89, 0x50, 0x4E, 0x47, 0x00, 0x00, 0x00}

	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "image.png"), binaryContent, 0o644))

	srv, err := New("", WithWorkDir(dir), WithNoClaude(true), WithTrackedFiles([]string{"image.png"}))
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/api/files/image.png", nil)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp fileContentResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.True(t, resp.IsBinary)
	assert.Empty(t, resp.Content)
	assert.Equal(t, int64(7), resp.Size)
}

func TestHandleGetFileContent_LargeFile(t *testing.T) {
	// Create a file larger than 5MB
	largeContent := strings.Repeat("x", 6*1024*1024)

	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "large.txt"), []byte(largeContent), 0o644))

	srv, err := New("", WithWorkDir(dir), WithNoClaude(true), WithTrackedFiles([]string{"large.txt"}))
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/api/files/large.txt", nil)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp fileContentResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.True(t, resp.IsTruncated)
	assert.Empty(t, resp.Content)
	assert.Equal(t, int64(6*1024*1024), resp.Size)
}
