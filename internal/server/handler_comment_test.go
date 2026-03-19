package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/shimasan0x00/difr/internal/comment"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testDiff contains diff data with multiple files for comment API tests.
const testDiff = `diff --git a/main.go b/main.go
index 1234567..abcdefg 100644
--- a/main.go
+++ b/main.go
@@ -1,3 +1,4 @@
 package main

+import "fmt"
 func main() {}
diff --git a/a.go b/a.go
index 1234567..abcdefg 100644
--- a/a.go
+++ b/a.go
@@ -1,2 +1,3 @@
 package main

+func a() {}
diff --git a/b.go b/b.go
index 1234567..abcdefg 100644
--- a/b.go
+++ b/b.go
@@ -1,2 +1,3 @@
 package main

+func b() {}
`

func setupCommentServer(t *testing.T) *Server {
	t.Helper()
	dir := t.TempDir()
	srv, err := New(testDiff, WithWorkDir(dir), WithNoClaude(true))
	require.NoError(t, err)
	return srv
}

// createCommentViaAPI is a test helper that creates a comment through the HTTP API.
func createCommentViaAPI(t *testing.T, s *Server, filePath string, line int, body string) comment.Comment {
	t.Helper()
	payload, err := json.Marshal(map[string]any{
		"filePath": filePath,
		"line":     line,
		"body":     body,
	})
	require.NoError(t, err)
	req := httptest.NewRequest(http.MethodPost, "/api/comments", bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code, "create comment should return 201")

	var c comment.Comment
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &c))
	return c
}

func TestCreateComment(t *testing.T) {
	s := setupCommentServer(t)

	c := createCommentViaAPI(t, s, "main.go", 10, "needs fix")

	assert.NotEmpty(t, c.ID)
	assert.Equal(t, "needs fix", c.Body)
	assert.Equal(t, "main.go", c.FilePath)
	assert.Equal(t, 10, c.Line)
}

func TestCreateComment_RejectsMalformedJSON(t *testing.T) {
	s := setupCommentServer(t)

	req := httptest.NewRequest(http.MethodPost, "/api/comments", bytes.NewBufferString(`{invalid`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestListComments_ReturnsAllCreated(t *testing.T) {
	s := setupCommentServer(t)
	createCommentViaAPI(t, s, "a.go", 1, "first")
	createCommentViaAPI(t, s, "b.go", 2, "second")

	req := httptest.NewRequest(http.MethodGet, "/api/comments", nil)
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var comments []*comment.Comment
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &comments))
	assert.Len(t, comments, 2)
}

func TestListComments_FiltersByFile(t *testing.T) {
	s := setupCommentServer(t)
	createCommentViaAPI(t, s, "a.go", 1, "first")
	createCommentViaAPI(t, s, "b.go", 2, "second")

	req := httptest.NewRequest(http.MethodGet, "/api/comments?file=a.go", nil)
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, req)

	var comments []*comment.Comment
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &comments))
	assert.Len(t, comments, 1)
	assert.Equal(t, "a.go", comments[0].FilePath)
}

func TestUpdateComment_ChangesBody(t *testing.T) {
	s := setupCommentServer(t)
	created := createCommentViaAPI(t, s, "main.go", 1, "old")

	updateBody := `{"body":"updated"}`
	req := httptest.NewRequest(http.MethodPut, "/api/comments/"+created.ID, bytes.NewBufferString(updateBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var updated comment.Comment
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &updated))
	assert.Equal(t, "updated", updated.Body)
}

func TestUpdateComment_Returns404ForNonexistent(t *testing.T) {
	s := setupCommentServer(t)

	req := httptest.NewRequest(http.MethodPut, "/api/comments/nonexistent", bytes.NewBufferString(`{"body":"x"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestDeleteComment_RemovesFromList(t *testing.T) {
	s := setupCommentServer(t)
	created := createCommentViaAPI(t, s, "main.go", 1, "delete me")

	// Act: Delete
	req := httptest.NewRequest(http.MethodDelete, "/api/comments/"+created.ID, nil)
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, req)
	assert.Equal(t, http.StatusNoContent, w.Code)

	// Assert: Verify deleted
	req = httptest.NewRequest(http.MethodGet, "/api/comments", nil)
	w = httptest.NewRecorder()
	s.Handler().ServeHTTP(w, req)

	var comments []*comment.Comment
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &comments))
	assert.Empty(t, comments)
}

func TestDeleteComment_Returns404ForNonexistent(t *testing.T) {
	s := setupCommentServer(t)

	req := httptest.NewRequest(http.MethodDelete, "/api/comments/nonexistent", nil)
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestUpdateComment_RejectsMalformedJSON(t *testing.T) {
	s := setupCommentServer(t)
	created := createCommentViaAPI(t, s, "main.go", 1, "original")

	req := httptest.NewRequest(http.MethodPut, "/api/comments/"+created.ID, bytes.NewBufferString(`{invalid`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUpdateComment_RejectsEmptyBody(t *testing.T) {
	s := setupCommentServer(t)
	created := createCommentViaAPI(t, s, "main.go", 1, "original")

	req := httptest.NewRequest(http.MethodPut, "/api/comments/"+created.ID, bytes.NewBufferString(`{"body":""}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateComment_RejectsEmptyFilePath(t *testing.T) {
	s := setupCommentServer(t)

	payload := `{"filePath":"","line":1,"body":"test"}`
	req := httptest.NewRequest(http.MethodPost, "/api/comments", bytes.NewBufferString(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateComment_RejectsPathTraversal(t *testing.T) {
	s := setupCommentServer(t)

	payload := `{"filePath":"../../../etc/passwd","line":1,"body":"test"}`
	req := httptest.NewRequest(http.MethodPost, "/api/comments", bytes.NewBufferString(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateComment_AcceptsZeroLineAsFileComment(t *testing.T) {
	s := setupCommentServer(t)

	c := createCommentViaAPI(t, s, "main.go", 0, "file-level comment")

	assert.NotEmpty(t, c.ID)
	assert.Equal(t, 0, c.Line)
	assert.Equal(t, "file-level comment", c.Body)
}

func TestCreateComment_RejectsNegativeLine(t *testing.T) {
	s := setupCommentServer(t)

	payload := `{"filePath":"main.go","line":-1,"body":"test"}`
	req := httptest.NewRequest(http.MethodPost, "/api/comments", bytes.NewBufferString(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestExportComments_ReturnsMarkdownByDefault(t *testing.T) {
	s := setupCommentServer(t)
	createCommentViaAPI(t, s, "main.go", 10, "fix this")

	req := httptest.NewRequest(http.MethodGet, "/api/comments/export", nil)
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "text/markdown; charset=utf-8", w.Header().Get("Content-Type"))
	assert.Contains(t, w.Body.String(), "fix this")
}

func TestExportComments_ReturnsJSON(t *testing.T) {
	s := setupCommentServer(t)
	createCommentViaAPI(t, s, "main.go", 10, "fix this")

	req := httptest.NewRequest(http.MethodGet, "/api/comments/export?format=json", nil)
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var comments []comment.Comment
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &comments))
	require.Len(t, comments, 1)
	assert.Equal(t, "fix this", comments[0].Body)
}

func TestCreateComment_RejectsFilePathNotInDiff(t *testing.T) {
	dir := t.TempDir()
	rawDiff := "diff --git a/hello.go b/hello.go\nindex 1234567..abcdefg 100644\n--- a/hello.go\n+++ b/hello.go\n@@ -1,3 +1,4 @@\n package main\n \n+import \"fmt\"\n func main() {}\n"
	s, err := New(rawDiff, WithWorkDir(dir), WithNoClaude(true))
	require.NoError(t, err)

	payload := `{"filePath":"notindiff.go","line":1,"body":"test"}`
	req := httptest.NewRequest(http.MethodPost, "/api/comments", bytes.NewBufferString(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "filePath is not in the current diff")
}

func TestCreateComment_RejectsMissingContentType(t *testing.T) {
	s := setupCommentServer(t)

	payload := `{"filePath":"main.go","line":1,"body":"test"}`
	req := httptest.NewRequest(http.MethodPost, "/api/comments", bytes.NewBufferString(payload))
	// No Content-Type header
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnsupportedMediaType, w.Code)
}

func TestCreateComment_RejectsWrongContentType(t *testing.T) {
	s := setupCommentServer(t)

	payload := `{"filePath":"main.go","line":1,"body":"test"}`
	req := httptest.NewRequest(http.MethodPost, "/api/comments", bytes.NewBufferString(payload))
	req.Header.Set("Content-Type", "text/plain")
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnsupportedMediaType, w.Code)
}

func TestUpdateComment_RejectsMissingContentType(t *testing.T) {
	s := setupCommentServer(t)
	created := createCommentViaAPI(t, s, "main.go", 1, "original")

	req := httptest.NewRequest(http.MethodPut, "/api/comments/"+created.ID, bytes.NewBufferString(`{"body":"updated"}`))
	// No Content-Type header
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnsupportedMediaType, w.Code)
}

func TestListComments_Pagination(t *testing.T) {
	s := setupCommentServer(t)
	// Create 5 comments
	for i := 1; i <= 5; i++ {
		createCommentViaAPI(t, s, "main.go", i, "comment")
	}

	// Limit only
	req := httptest.NewRequest(http.MethodGet, "/api/comments?limit=2", nil)
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	var comments []*comment.Comment
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &comments))
	assert.Len(t, comments, 2)

	// Limit + offset
	req = httptest.NewRequest(http.MethodGet, "/api/comments?limit=2&offset=3", nil)
	w = httptest.NewRecorder()
	s.Handler().ServeHTTP(w, req)
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &comments))
	assert.Len(t, comments, 2, "should return remaining 2 comments at offset 3")

	// Offset beyond total
	req = httptest.NewRequest(http.MethodGet, "/api/comments?limit=10&offset=100", nil)
	w = httptest.NewRecorder()
	s.Handler().ServeHTTP(w, req)
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &comments))
	assert.Empty(t, comments)

	// Invalid limit
	req = httptest.NewRequest(http.MethodGet, "/api/comments?limit=0", nil)
	w = httptest.NewRecorder()
	s.Handler().ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// Invalid offset
	req = httptest.NewRequest(http.MethodGet, "/api/comments?limit=2&offset=-1", nil)
	w = httptest.NewRecorder()
	s.Handler().ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDeleteAllComments_RemovesAll(t *testing.T) {
	s := setupCommentServer(t)
	createCommentViaAPI(t, s, "main.go", 1, "first")
	createCommentViaAPI(t, s, "a.go", 2, "second")

	req := httptest.NewRequest(http.MethodDelete, "/api/comments", nil)
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, req)
	assert.Equal(t, http.StatusNoContent, w.Code)

	// Verify all comments deleted
	req = httptest.NewRequest(http.MethodGet, "/api/comments", nil)
	w = httptest.NewRecorder()
	s.Handler().ServeHTTP(w, req)

	var comments []*comment.Comment
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &comments))
	assert.Empty(t, comments)
}

func TestCreateComment_WithCategoryAndSeverity(t *testing.T) {
	s := setupCommentServer(t)

	payload := `{"filePath":"main.go","line":10,"body":"fix","reviewCategory":"MUST","severity":"Critical"}`
	req := httptest.NewRequest(http.MethodPost, "/api/comments", bytes.NewBufferString(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	var c comment.Comment
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &c))
	assert.Equal(t, "MUST", c.ReviewCategory)
	assert.Equal(t, "Critical", c.Severity)
}

func TestCreateComment_RejectsInvalidCategory(t *testing.T) {
	s := setupCommentServer(t)

	payload := `{"filePath":"main.go","line":1,"body":"test","reviewCategory":"INVALID"}`
	req := httptest.NewRequest(http.MethodPost, "/api/comments", bytes.NewBufferString(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "invalid reviewCategory")
}

func TestCreateComment_RejectsInvalidSeverity(t *testing.T) {
	s := setupCommentServer(t)

	payload := `{"filePath":"main.go","line":1,"body":"test","severity":"INVALID"}`
	req := httptest.NewRequest(http.MethodPost, "/api/comments", bytes.NewBufferString(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "invalid severity")
}

func TestUpdateComment_WithCategoryAndSeverity(t *testing.T) {
	s := setupCommentServer(t)
	created := createCommentViaAPI(t, s, "main.go", 1, "old")

	payload := `{"body":"updated","reviewCategory":"IMO","severity":"High"}`
	req := httptest.NewRequest(http.MethodPut, "/api/comments/"+created.ID, bytes.NewBufferString(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var updated comment.Comment
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &updated))
	assert.Equal(t, "updated", updated.Body)
	assert.Equal(t, "IMO", updated.ReviewCategory)
	assert.Equal(t, "High", updated.Severity)
}

func TestUpdateComment_RejectsInvalidCategory(t *testing.T) {
	s := setupCommentServer(t)
	created := createCommentViaAPI(t, s, "main.go", 1, "old")

	payload := `{"body":"updated","reviewCategory":"BAD"}`
	req := httptest.NewRequest(http.MethodPut, "/api/comments/"+created.ID, bytes.NewBufferString(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestExportComments_ReturnsCSV(t *testing.T) {
	s := setupCommentServer(t)

	// Create comment with category/severity via full payload
	payload := `{"filePath":"main.go","line":10,"body":"fix this","reviewCategory":"MUST","severity":"Critical"}`
	req := httptest.NewRequest(http.MethodPost, "/api/comments", bytes.NewBufferString(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)

	req = httptest.NewRequest(http.MethodGet, "/api/comments/export?format=csv", nil)
	w = httptest.NewRecorder()
	s.Handler().ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "text/csv; charset=utf-8", w.Header().Get("Content-Type"))
	assert.Equal(t, `attachment; filename="comments.csv"`, w.Header().Get("Content-Disposition"))
	assert.Contains(t, w.Body.String(), "filepath,review_category,severity,comment")
	assert.Contains(t, w.Body.String(), "main.go,MUST,Critical,fix this")
}

func TestExportComments_IncludesContentDisposition(t *testing.T) {
	s := setupCommentServer(t)

	// Markdown export
	req := httptest.NewRequest(http.MethodGet, "/api/comments/export", nil)
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, req)
	assert.Equal(t, `attachment; filename="comments.md"`, w.Header().Get("Content-Disposition"))

	// JSON export
	req = httptest.NewRequest(http.MethodGet, "/api/comments/export?format=json", nil)
	w = httptest.NewRecorder()
	s.Handler().ServeHTTP(w, req)
	assert.Equal(t, `attachment; filename="comments.json"`, w.Header().Get("Content-Disposition"))

	// Excel export
	req = httptest.NewRequest(http.MethodGet, "/api/comments/export?format=xlsx", nil)
	w = httptest.NewRecorder()
	s.Handler().ServeHTTP(w, req)
	assert.Equal(t, `attachment; filename="comments.xlsx"`, w.Header().Get("Content-Disposition"))
}

func TestExportComments_ReturnsXlsx(t *testing.T) {
	s := setupCommentServer(t)

	// Create a comment
	payload := `{"filePath":"main.go","line":10,"body":"fix this","reviewCategory":"MUST","severity":"Critical"}`
	req := httptest.NewRequest(http.MethodPost, "/api/comments", bytes.NewBufferString(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)

	// Export as xlsx
	req = httptest.NewRequest(http.MethodGet, "/api/comments/export?format=xlsx", nil)
	w = httptest.NewRecorder()
	s.Handler().ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", w.Header().Get("Content-Type"))
	assert.Equal(t, `attachment; filename="comments.xlsx"`, w.Header().Get("Content-Disposition"))

	// Verify xlsx content is valid and contains data
	assert.True(t, len(w.Body.Bytes()) > 0)
	// xlsx files start with PK (zip signature)
	assert.Equal(t, byte('P'), w.Body.Bytes()[0])
	assert.Equal(t, byte('K'), w.Body.Bytes()[1])
}
