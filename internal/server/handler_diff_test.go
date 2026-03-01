package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/shimasan0x00/difr/internal/diff"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testDiffRaw = `diff --git a/main.go b/main.go
index 1234567..abcdefg 100644
--- a/main.go
+++ b/main.go
@@ -1,3 +1,3 @@
 package main

-func old() {}
+func new() {}
diff --git a/utils.go b/utils.go
new file mode 100644
index 0000000..1234567
--- /dev/null
+++ b/utils.go
@@ -0,0 +1,3 @@
+package main
+
+func helper() {}
`

func newTestServer(t *testing.T, rawDiff string) *Server {
	t.Helper()
	dir := t.TempDir()
	srv, err := New(rawDiff, WithWorkDir(dir), WithNoClaude(true))
	require.NoError(t, err)
	return srv
}

func TestHandleGetDiff_ReturnsJSON(t *testing.T) {
	srv := newTestServer(t, testDiffRaw)

	req := httptest.NewRequest(http.MethodGet, "/api/diff", nil)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))

	var resp diff.DiffResult
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Len(t, resp.Files, 2)
	assert.Equal(t, 4, resp.Stats.Additions)
	assert.Equal(t, 1, resp.Stats.Deletions)
}

func TestHandleGetDiff_EmptyDiff(t *testing.T) {
	srv := newTestServer(t, "")

	req := httptest.NewRequest(http.MethodGet, "/api/diff", nil)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp diff.DiffResult
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Empty(t, resp.Files)
}

func TestHandleGetDiffFiles_ReturnsList(t *testing.T) {
	srv := newTestServer(t, testDiffRaw)

	req := httptest.NewRequest(http.MethodGet, "/api/diff/files", nil)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var files []diff.DiffFile
	err := json.Unmarshal(rec.Body.Bytes(), &files)
	require.NoError(t, err)
	require.Len(t, files, 2)
	assert.Equal(t, "main.go", files[0].NewPath)
	assert.Equal(t, "utils.go", files[1].NewPath)
}

func TestHandleGetDiffFileByPath_ReturnsFile(t *testing.T) {
	srv := newTestServer(t, testDiffRaw)

	req := httptest.NewRequest(http.MethodGet, "/api/diff/files/main.go", nil)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var f diff.DiffFile
	err := json.Unmarshal(rec.Body.Bytes(), &f)
	require.NoError(t, err)
	assert.Equal(t, "main.go", f.NewPath)
	assert.Equal(t, diff.FileStatusModified, f.Status)
}

func TestHandleGetDiffFileByPath_NotFound(t *testing.T) {
	srv := newTestServer(t, testDiffRaw)

	req := httptest.NewRequest(http.MethodGet, "/api/diff/files/nonexistent.go", nil)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestHandleGetDiffStats_ReturnsStats(t *testing.T) {
	srv := newTestServer(t, testDiffRaw)

	req := httptest.NewRequest(http.MethodGet, "/api/diff/stats", nil)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var stats struct {
		Files int            `json:"files"`
		Stats diff.FileStats `json:"stats"`
	}
	err := json.Unmarshal(rec.Body.Bytes(), &stats)
	require.NoError(t, err)
	assert.Equal(t, 2, stats.Files)
	assert.Equal(t, 4, stats.Stats.Additions)
	assert.Equal(t, 1, stats.Stats.Deletions)
}

func TestHandleClaudeStatus_ReturnsTrueWhenRunnerAvailable(t *testing.T) {
	dir := t.TempDir()
	srv, err := New("", WithWorkDir(dir), WithNoClaude(true))
	require.NoError(t, err)
	// Manually set a runner to simulate Claude being available
	srv.claudeRunner = &mockRunner{output: ""}

	req := httptest.NewRequest(http.MethodGet, "/api/claude/status", nil)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp map[string]bool
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.True(t, resp["available"])
}

func TestHandleClaudeStatus_ReturnsFalseWhenDisabled(t *testing.T) {
	srv := newTestServer(t, testDiffRaw)

	req := httptest.NewRequest(http.MethodGet, "/api/claude/status", nil)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp map[string]bool
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.False(t, resp["available"])
}

func TestHandleHealthCheck_ReturnsOK(t *testing.T) {
	srv := newTestServer(t, "")

	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp map[string]string
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, "ok", resp["status"])
}
