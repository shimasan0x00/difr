package git

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/shimasan0x00/difr/internal/diff"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupTestRepo creates a temporary git repository with a commit for testing.
func setupTestRepo(t *testing.T) string {
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

// addFileAndCommit creates a file and commits it.
func addFileAndCommit(t *testing.T, dir, filename, content, message string) {
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

func TestGetDiff_LatestCommit(t *testing.T) {
	dir := setupTestRepo(t)
	addFileAndCommit(t, dir, "hello.go", "package main\n", "initial")
	addFileAndCommit(t, dir, "hello.go", "package main\n\nfunc hello() {}\n", "add func")

	client := NewClient(dir)
	d, err := client.GetDiff(context.Background(), diff.DiffRequest{Mode: diff.DiffModeLatestCommit})
	require.NoError(t, err)

	assert.Contains(t, d, "diff --git")
	assert.Contains(t, d, "hello.go")
	assert.Contains(t, d, "+func hello()")
}

func TestGetDiff_Staged(t *testing.T) {
	dir := setupTestRepo(t)
	addFileAndCommit(t, dir, "file.go", "package main\n", "initial")

	// Stage a change
	fpath := filepath.Join(dir, "file.go")
	require.NoError(t, os.WriteFile(fpath, []byte("package main\n\nvar x = 1\n"), 0o644))
	cmd := exec.Command("git", "add", "file.go")
	cmd.Dir = dir
	require.NoError(t, cmd.Run())

	client := NewClient(dir)
	d, err := client.GetDiff(context.Background(), diff.DiffRequest{Mode: diff.DiffModeStaged})
	require.NoError(t, err)

	assert.Contains(t, d, "+var x = 1")
}

func TestGetDiff_Working(t *testing.T) {
	dir := setupTestRepo(t)
	addFileAndCommit(t, dir, "file.go", "package main\n", "initial")

	// Make unstaged change
	fpath := filepath.Join(dir, "file.go")
	require.NoError(t, os.WriteFile(fpath, []byte("package main\n\nvar y = 2\n"), 0o644))

	client := NewClient(dir)
	d, err := client.GetDiff(context.Background(), diff.DiffRequest{Mode: diff.DiffModeWorking})
	require.NoError(t, err)

	assert.Contains(t, d, "+var y = 2")
}

func TestGetDiff_Range(t *testing.T) {
	dir := setupTestRepo(t)
	addFileAndCommit(t, dir, "file.go", "v1\n", "v1")

	cmd := exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = dir
	out, err := cmd.Output()
	require.NoError(t, err)
	from := strings.TrimSpace(string(out))

	addFileAndCommit(t, dir, "file.go", "v2\n", "v2")

	cmd = exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = dir
	out, err = cmd.Output()
	require.NoError(t, err)
	to := strings.TrimSpace(string(out))

	client := NewClient(dir)
	d, err := client.GetDiff(context.Background(), diff.DiffRequest{
		Mode: diff.DiffModeRange,
		From: from,
		To:   to,
	})
	require.NoError(t, err)

	assert.Contains(t, d, "-v1")
	assert.Contains(t, d, "+v2")
}

func TestGetDiff_SingleCommit(t *testing.T) {
	dir := setupTestRepo(t)
	addFileAndCommit(t, dir, "file.go", "v1\n", "v1")
	addFileAndCommit(t, dir, "file.go", "v2\n", "v2")

	cmd := exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = dir
	out, err := cmd.Output()
	require.NoError(t, err)
	commitHash := strings.TrimSpace(string(out))

	client := NewClient(dir)
	d, err := client.GetDiff(context.Background(), diff.DiffRequest{
		Mode: diff.DiffModeCommit,
		From: commitHash,
	})
	require.NoError(t, err)

	assert.Contains(t, d, "-v1")
	assert.Contains(t, d, "+v2")
}

func TestGetDiff_Stdin(t *testing.T) {
	dir := setupTestRepo(t)
	client := NewClient(dir)

	stdinContent := "diff --git a/test.go b/test.go\n--- a/test.go\n+++ b/test.go\n@@ -1 +1 @@\n-old\n+new\n"
	d, err := client.GetDiff(context.Background(), diff.DiffRequest{
		Mode:  diff.DiffModeStdin,
		Stdin: strings.NewReader(stdinContent),
	})
	require.NoError(t, err)

	assert.Equal(t, stdinContent, d)
}

func TestBuildDiffArgs_RejectsDashPrefixedRef(t *testing.T) {
	client := NewClient(".")

	tests := []struct {
		name string
		req  diff.DiffRequest
	}{
		{"commit mode with dash prefix", diff.DiffRequest{Mode: diff.DiffModeCommit, From: "--exec=malicious"}},
		{"range mode with dash prefix in From", diff.DiffRequest{Mode: diff.DiffModeRange, From: "-x", To: "HEAD"}},
		{"range mode with dash prefix in To", diff.DiffRequest{Mode: diff.DiffModeRange, From: "HEAD", To: "--evil"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := client.buildDiffArgs(tt.req)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "must not start with '-'")
		})
	}
}

func TestGetDiff_EmptyDiff(t *testing.T) {
	dir := setupTestRepo(t)
	addFileAndCommit(t, dir, "file.go", "unchanged\n", "initial")

	client := NewClient(dir)
	d, err := client.GetDiff(context.Background(), diff.DiffRequest{Mode: diff.DiffModeWorking})
	require.NoError(t, err)

	assert.Empty(t, d)
}
