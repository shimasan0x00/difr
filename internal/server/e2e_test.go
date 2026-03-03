package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/shimasan0x00/difr/internal/comment"
	"github.com/shimasan0x00/difr/internal/diff"
	"github.com/shimasan0x00/difr/internal/git"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- E2E Test Helpers ---

// startE2EServer starts a real HTTP server and returns both URL and Server.
func startE2EServer(t *testing.T, rawDiff string, opts ...Option) (baseURL string, srv *Server) {
	t.Helper()
	dir := t.TempDir()
	allOpts := append([]Option{WithWorkDir(dir), WithNoClaude(true)}, opts...)
	srv, err := New(rawDiff, allOpts...)
	require.NoError(t, err)

	ts := httptest.NewServer(srv.Handler())
	t.Cleanup(ts.Close)
	return ts.URL, srv
}

// setupE2EGitRepo creates a temporary git repository with changes for E2E tests.
func setupE2EGitRepo(t *testing.T) (repoDir string) {
	t.Helper()
	dir := t.TempDir()

	cmds := [][]string{
		{"git", "init"},
		{"git", "config", "user.email", "test@test.com"},
		{"git", "config", "user.name", "Test"},
		{"git", "checkout", "-b", "test-branch"},
	}
	for _, args := range cmds {
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Dir = dir
		out, err := cmd.CombinedOutput()
		require.NoError(t, err, "cmd %v failed: %s", args, string(out))
	}

	return dir
}

// e2eAddFileAndCommit creates a file and commits it in the E2E test repo.
func e2eAddFileAndCommit(t *testing.T, dir, filename, content, message string) {
	t.Helper()
	fpath := filepath.Join(dir, filename)
	require.NoError(t, os.MkdirAll(filepath.Dir(fpath), 0o755))
	require.NoError(t, os.WriteFile(fpath, []byte(content), 0o644))

	cmds := [][]string{
		{"git", "add", filename},
		{"git", "commit", "-m", message},
	}
	for _, args := range cmds {
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Dir = dir
		out, err := cmd.CombinedOutput()
		require.NoError(t, err, "cmd %v failed: %s", args, string(out))
	}
}

// httpGet performs a GET request to the E2E server.
func httpGet(t *testing.T, url string) (statusCode int, body []byte) {
	t.Helper()
	resp, err := http.Get(url)
	require.NoError(t, err)
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	return resp.StatusCode, data
}

// httpPost performs a POST request with JSON body.
func httpPost(t *testing.T, url string, jsonBody string) (statusCode int, body []byte) {
	t.Helper()
	resp, err := http.Post(url, "application/json", strings.NewReader(jsonBody))
	require.NoError(t, err)
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	return resp.StatusCode, data
}

// httpPut performs a PUT request with JSON body.
func httpPut(t *testing.T, url string, jsonBody string) (statusCode int, body []byte) {
	t.Helper()
	req, err := http.NewRequest(http.MethodPut, url, strings.NewReader(jsonBody))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	return resp.StatusCode, data
}

// httpDelete performs a DELETE request.
func httpDelete(t *testing.T, url string) (statusCode int, body []byte) {
	t.Helper()
	req, err := http.NewRequest(http.MethodDelete, url, nil)
	require.NoError(t, err)
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	return resp.StatusCode, data
}

// --- A. Diff Workflow Tests ---

func TestE2E_DiffWorkflow_GitRepoThroughAPI(t *testing.T) {
	// Arrange: Create a real git repo with changes
	repoDir := setupE2EGitRepo(t)
	e2eAddFileAndCommit(t, repoDir, "hello.go", "package main\n", "initial")
	e2eAddFileAndCommit(t, repoDir, "hello.go", "package main\n\nimport \"fmt\"\n\nfunc hello() {\n\tfmt.Println(\"hello\")\n}\n", "add hello func")

	// Get raw diff from real git repo
	client := git.NewClient(repoDir)
	rawDiff, err := client.GetDiff(context.Background(), diff.DiffRequest{Mode: diff.DiffModeLatestCommit})
	require.NoError(t, err)
	require.NotEmpty(t, rawDiff, "git diff should produce output")

	// Start E2E server with real diff
	baseURL, _ := startE2EServer(t, rawDiff)

	// Act & Assert: GET /api/diff
	status, body := httpGet(t, baseURL+"/api/diff")
	assert.Equal(t, http.StatusOK, status)
	var diffResult diff.DiffResult
	require.NoError(t, json.Unmarshal(body, &diffResult))
	assert.NotEmpty(t, diffResult.Files, "should have at least one file")
	assert.Greater(t, diffResult.Stats.Additions, 0, "should have additions")

	// Act & Assert: GET /api/diff/files
	status, body = httpGet(t, baseURL+"/api/diff/files")
	assert.Equal(t, http.StatusOK, status)
	var files []diff.DiffFile
	require.NoError(t, json.Unmarshal(body, &files))
	assert.Len(t, files, 1, "should have exactly one changed file")
	assert.Equal(t, "hello.go", files[0].NewPath)

	// Act & Assert: GET /api/diff/files/hello.go
	status, body = httpGet(t, baseURL+"/api/diff/files/hello.go")
	assert.Equal(t, http.StatusOK, status)
	var file diff.DiffFile
	require.NoError(t, json.Unmarshal(body, &file))
	assert.Equal(t, "hello.go", file.NewPath)

	// Act & Assert: GET /api/diff/stats
	status, body = httpGet(t, baseURL+"/api/diff/stats")
	assert.Equal(t, http.StatusOK, status)
	var stats struct {
		Files int            `json:"files"`
		Stats diff.FileStats `json:"stats"`
	}
	require.NoError(t, json.Unmarshal(body, &stats))
	assert.Equal(t, 1, stats.Files)
	assert.Equal(t, diffResult.Stats.Additions, stats.Stats.Additions, "stats should match diff result")
	assert.Equal(t, diffResult.Stats.Deletions, stats.Stats.Deletions, "stats should match diff result")
}

func TestE2E_DiffWorkflow_MultiFileRename(t *testing.T) {
	// Arrange: Multi-file diff with add, delete, and modify
	rawDiff := `diff --git a/main.go b/main.go
index 1234567..abcdefg 100644
--- a/main.go
+++ b/main.go
@@ -1,3 +1,4 @@
 package main

+import "fmt"
 func main() {}
diff --git a/utils.go b/utils.go
new file mode 100644
index 0000000..1234567
--- /dev/null
+++ b/utils.go
@@ -0,0 +1,3 @@
+package main
+
+func helper() {}
diff --git a/old.go b/old.go
deleted file mode 100644
index 1234567..0000000
--- a/old.go
+++ /dev/null
@@ -1,3 +0,0 @@
-package main
-
-func deprecated() {}
`
	baseURL, _ := startE2EServer(t, rawDiff)

	// Verify file list
	status, body := httpGet(t, baseURL+"/api/diff/files")
	assert.Equal(t, http.StatusOK, status)
	var files []diff.DiffFile
	require.NoError(t, json.Unmarshal(body, &files))
	assert.Len(t, files, 3, "should have 3 files")

	// Verify stats consistency
	status, body = httpGet(t, baseURL+"/api/diff/stats")
	assert.Equal(t, http.StatusOK, status)
	var stats struct {
		Files int            `json:"files"`
		Stats diff.FileStats `json:"stats"`
	}
	require.NoError(t, json.Unmarshal(body, &stats))
	assert.Equal(t, 3, stats.Files)

	// Sum up individual file stats and verify they match aggregate stats
	totalAdds := 0
	totalDels := 0
	for _, f := range files {
		totalAdds += f.Stats.Additions
		totalDels += f.Stats.Deletions
	}
	assert.Equal(t, totalAdds, stats.Stats.Additions, "aggregate additions should match sum of file additions")
	assert.Equal(t, totalDels, stats.Stats.Deletions, "aggregate deletions should match sum of file deletions")

	// Verify individual file access
	for _, f := range files {
		path := f.NewPath
		if path == "" {
			path = f.OldPath
		}
		status, body = httpGet(t, baseURL+"/api/diff/files/"+path)
		assert.Equal(t, http.StatusOK, status, "should be able to access file: %s", path)

		var fetched diff.DiffFile
		require.NoError(t, json.Unmarshal(body, &fetched))
		assert.Equal(t, f.Status, fetched.Status, "status should match for file: %s", path)
	}
}

// --- B. Comment Lifecycle Tests ---

func TestE2E_CommentLifecycle_CreateUpdateDeleteExport(t *testing.T) {
	rawDiff := `diff --git a/main.go b/main.go
index 1234567..abcdefg 100644
--- a/main.go
+++ b/main.go
@@ -1,3 +1,4 @@
 package main

+import "fmt"
 func main() {}
`
	baseURL, _ := startE2EServer(t, rawDiff)

	// Step 1: Create a comment
	createPayload, err := json.Marshal(map[string]any{
		"filePath": "main.go",
		"line":     3,
		"body":     "Why import fmt here?",
	})
	require.NoError(t, err)
	status, body := httpPost(t, baseURL+"/api/comments", string(createPayload))
	assert.Equal(t, http.StatusCreated, status)
	var created comment.Comment
	require.NoError(t, json.Unmarshal(body, &created))
	assert.Equal(t, "Why import fmt here?", created.Body)
	assert.Equal(t, "main.go", created.FilePath)
	assert.Equal(t, 3, created.Line)
	assert.NotEmpty(t, created.ID)

	// Step 2: List comments and verify existence
	status, body = httpGet(t, baseURL+"/api/comments")
	assert.Equal(t, http.StatusOK, status)
	var comments []*comment.Comment
	require.NoError(t, json.Unmarshal(body, &comments))
	assert.Len(t, comments, 1)
	assert.Equal(t, created.ID, comments[0].ID)

	// Step 3: Update the comment
	updatePayload, err := json.Marshal(map[string]string{"body": "Updated: fmt is needed for Println"})
	require.NoError(t, err)
	status, body = httpPut(t, baseURL+"/api/comments/"+created.ID, string(updatePayload))
	assert.Equal(t, http.StatusOK, status)
	var updated comment.Comment
	require.NoError(t, json.Unmarshal(body, &updated))
	assert.Equal(t, "Updated: fmt is needed for Println", updated.Body)

	// Step 4: Export and verify updated body is reflected
	status, body = httpGet(t, baseURL+"/api/comments/export")
	assert.Equal(t, http.StatusOK, status)
	assert.Contains(t, string(body), "Updated: fmt is needed for Println")

	// Step 5: Delete the comment
	status, _ = httpDelete(t, baseURL+"/api/comments/"+created.ID)
	assert.Equal(t, http.StatusNoContent, status)

	// Step 6: Verify deletion
	status, body = httpGet(t, baseURL+"/api/comments")
	assert.Equal(t, http.StatusOK, status)
	require.NoError(t, json.Unmarshal(body, &comments))
	assert.Empty(t, comments)
}

func TestE2E_CommentLifecycle_MultiFileFiltering(t *testing.T) {
	rawDiff := `diff --git a/main.go b/main.go
index 1234567..abcdefg 100644
--- a/main.go
+++ b/main.go
@@ -1,3 +1,4 @@
 package main

+import "fmt"
 func main() {}
diff --git a/utils.go b/utils.go
index 1234567..abcdefg 100644
--- a/utils.go
+++ b/utils.go
@@ -1,2 +1,3 @@
 package main

+func helper() {}
`
	baseURL, _ := startE2EServer(t, rawDiff)

	// Create comments on different files
	for _, tc := range []struct {
		filePath string
		line     int
		body     string
	}{
		{"main.go", 3, "Comment on main.go"},
		{"main.go", 4, "Another comment on main.go"},
		{"utils.go", 3, "Comment on utils.go"},
	} {
		payload, err := json.Marshal(map[string]any{
			"filePath": tc.filePath,
			"line":     tc.line,
			"body":     tc.body,
		})
		require.NoError(t, err)
		status, _ := httpPost(t, baseURL+"/api/comments", string(payload))
		assert.Equal(t, http.StatusCreated, status)
	}

	// Filter by main.go
	status, body := httpGet(t, baseURL+"/api/comments?file=main.go")
	assert.Equal(t, http.StatusOK, status)
	var mainComments []*comment.Comment
	require.NoError(t, json.Unmarshal(body, &mainComments))
	assert.Len(t, mainComments, 2, "should have 2 comments for main.go")

	// Filter by utils.go
	status, body = httpGet(t, baseURL+"/api/comments?file=utils.go")
	assert.Equal(t, http.StatusOK, status)
	var utilsComments []*comment.Comment
	require.NoError(t, json.Unmarshal(body, &utilsComments))
	assert.Len(t, utilsComments, 1, "should have 1 comment for utils.go")

	// Markdown export includes all comments
	status, body = httpGet(t, baseURL+"/api/comments/export")
	assert.Equal(t, http.StatusOK, status)
	exportStr := string(body)
	assert.Contains(t, exportStr, "main.go")
	assert.Contains(t, exportStr, "utils.go")
	assert.Contains(t, exportStr, "Comment on main.go")
	assert.Contains(t, exportStr, "Comment on utils.go")

	// JSON export includes all comments
	status, body = httpGet(t, baseURL+"/api/comments/export?format=json")
	assert.Equal(t, http.StatusOK, status)
	var jsonComments []comment.Comment
	require.NoError(t, json.Unmarshal(body, &jsonComments))
	assert.Len(t, jsonComments, 3, "JSON export should have all 3 comments")
}

// --- C. WebSocket Claude Tests ---

func TestE2E_Claude_ChatWithSessionReuse(t *testing.T) {
	// Arrange: Mock Claude runner that returns session ID
	mockOutput1 := `{"type":"system","subtype":"init","session_id":"e2e-session-1"}
{"type":"assistant","message":{"content":[{"type":"text","text":"Hello from E2E!"}]},"session_id":"e2e-session-1"}
{"type":"result","subtype":"success","result":"Hello from E2E!","session_id":"e2e-session-1","stop_reason":"end_turn"}
`
	mockOutput2 := `{"type":"system","subtype":"init","session_id":"e2e-session-1"}
{"type":"assistant","message":{"content":[{"type":"text","text":"Session reused!"}]},"session_id":"e2e-session-1"}
{"type":"result","subtype":"success","result":"Session reused!","session_id":"e2e-session-1","stop_reason":"end_turn"}
`
	callCount := 0
	runner := &sequentialMockRunner{
		outputs: []string{mockOutput1, mockOutput2},
		counter: &callCount,
	}

	dir := t.TempDir()
	srv, err := New("", WithWorkDir(dir), WithNoClaude(true))
	require.NoError(t, err)
	srv.claudeRunner = runner

	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()

	wsURL := "ws" + strings.TrimPrefix(ts.URL, "http") + "/ws/claude"
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	conn, _, err := websocket.Dial(ctx, wsURL, nil)
	require.NoError(t, err)
	defer conn.CloseNow()

	// First chat message
	msg1 := ChatMessage{Type: "chat", Content: "Hello"}
	require.NoError(t, wsjson.Write(ctx, conn, msg1))
	responses1 := readWSResponses(t, ctx, conn)
	assertHasResponseType(t, responses1, "session", "first chat should have session")
	assertHasResponseType(t, responses1, "done", "first chat should complete")

	// Second chat — server should reuse session via -r flag
	msg2 := ChatMessage{Type: "chat", Content: "Follow up"}
	require.NoError(t, wsjson.Write(ctx, conn, msg2))
	responses2 := readWSResponses(t, ctx, conn)
	assertHasResponseType(t, responses2, "done", "second chat should complete")

	// Verify session was reused (runner should have received -r flag)
	assert.Equal(t, 2, callCount, "runner should have been called twice")
	assert.Contains(t, runner.lastArgs, "-r", "second call should include -r flag for session resume")
	assert.Contains(t, runner.lastArgs, "e2e-session-1", "second call should include session ID")
}

func TestE2E_Claude_ReviewFlow(t *testing.T) {
	mockOutput := `{"type":"system","subtype":"init","session_id":"review-e2e"}
{"type":"assistant","message":{"content":[{"type":"text","text":"Code looks good overall."}]},"session_id":"review-e2e"}
{"type":"result","subtype":"success","result":"Code looks good overall.","session_id":"review-e2e","stop_reason":"end_turn"}
`
	dir := t.TempDir()
	srv, err := New("", WithWorkDir(dir), WithNoClaude(true))
	require.NoError(t, err)
	srv.claudeRunner = &mockRunner{output: mockOutput}

	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()

	wsURL := "ws" + strings.TrimPrefix(ts.URL, "http") + "/ws/claude"
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, _, err := websocket.Dial(ctx, wsURL, nil)
	require.NoError(t, err)
	defer conn.CloseNow()

	// Send review request
	msg := ChatMessage{Type: "review", Content: "Review this diff"}
	require.NoError(t, wsjson.Write(ctx, conn, msg))

	responses := readWSResponses(t, ctx, conn)

	// Verify streaming responses
	assertHasResponseType(t, responses, "session", "review should have session init")
	assertHasResponseType(t, responses, "text", "review should have text content")
	assertHasResponseType(t, responses, "done", "review should complete with done")

	// Verify done event has content
	for _, r := range responses {
		if r.Type == "done" {
			assert.NotEmpty(t, r.Content, "done event should have result content")
			assert.NotEmpty(t, r.SessionID, "done event should have session ID")
		}
	}
}

// --- D. Error Handling Tests ---

func TestE2E_ErrorHandling_InvalidCommentInputs(t *testing.T) {
	rawDiff := `diff --git a/main.go b/main.go
index 1234567..abcdefg 100644
--- a/main.go
+++ b/main.go
@@ -1,3 +1,4 @@
 package main

+import "fmt"
 func main() {}
`
	baseURL, _ := startE2EServer(t, rawDiff)

	tests := []struct {
		name    string
		payload string
	}{
		{"empty filePath", `{"filePath":"","line":1,"body":"test"}`},
		{"negative line number", `{"filePath":"main.go","line":-1,"body":"test"}`},
		{"path traversal", `{"filePath":"../../../etc/passwd","line":1,"body":"test"}`},
		{"file not in diff", `{"filePath":"notindiff.go","line":1,"body":"test"}`},
		{"empty body", `{"filePath":"main.go","line":1,"body":""}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, _ := httpPost(t, baseURL+"/api/comments", tt.payload)
			assert.Equal(t, http.StatusBadRequest, status, "invalid input should return 400")
		})
	}
}

func TestE2E_ErrorHandling_NonexistentResources(t *testing.T) {
	rawDiff := `diff --git a/main.go b/main.go
index 1234567..abcdefg 100644
--- a/main.go
+++ b/main.go
@@ -1,3 +1,4 @@
 package main

+import "fmt"
 func main() {}
`
	baseURL, _ := startE2EServer(t, rawDiff)

	// Update nonexistent comment
	status, _ := httpPut(t, baseURL+"/api/comments/nonexistent", `{"body":"updated"}`)
	assert.Equal(t, http.StatusNotFound, status, "updating nonexistent comment should return 404")

	// Delete nonexistent comment
	status, _ = httpDelete(t, baseURL+"/api/comments/nonexistent")
	assert.Equal(t, http.StatusNotFound, status, "deleting nonexistent comment should return 404")

	// Get nonexistent file from diff
	status, _ = httpGet(t, baseURL+"/api/diff/files/nonexistent.go")
	assert.Equal(t, http.StatusNotFound, status, "nonexistent file should return 404")
}

// --- E. Server Config Tests ---

func TestE2E_ServerConfig_ViewModeAndClaudeStatus(t *testing.T) {
	// Start server in unified mode with Claude disabled
	baseURL, srv := startE2EServer(t, "", WithViewMode("unified"))

	// Verify view mode
	status, body := httpGet(t, baseURL+"/api/diff/mode")
	assert.Equal(t, http.StatusOK, status)
	var modeResp map[string]string
	require.NoError(t, json.Unmarshal(body, &modeResp))
	assert.Equal(t, "unified", modeResp["mode"])

	// Verify Claude is disabled (server created with WithNoClaude(true))
	status, body = httpGet(t, baseURL+"/api/claude/status")
	assert.Equal(t, http.StatusOK, status)
	var claudeResp map[string]bool
	require.NoError(t, json.Unmarshal(body, &claudeResp))
	assert.False(t, claudeResp["available"], "Claude should be disabled")

	// Enable Claude by setting runner
	srv.claudeRunner = &mockRunner{output: ""}

	// Verify Claude is now available
	status, body = httpGet(t, baseURL+"/api/claude/status")
	assert.Equal(t, http.StatusOK, status)
	require.NoError(t, json.Unmarshal(body, &claudeResp))
	assert.True(t, claudeResp["available"], "Claude should be enabled after setting runner")
}

// --- Helper Types ---

// sequentialMockRunner returns different outputs for sequential calls.
type sequentialMockRunner struct {
	outputs  []string
	counter  *int
	lastArgs []string
}

func (m *sequentialMockRunner) Run(_ context.Context, args []string) (io.ReadCloser, error) {
	m.lastArgs = args
	idx := *m.counter
	*m.counter++
	if idx >= len(m.outputs) {
		return nil, fmt.Errorf("unexpected call #%d", idx)
	}
	return io.NopCloser(strings.NewReader(m.outputs[idx])), nil
}
