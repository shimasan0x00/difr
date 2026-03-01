package server

import (
	"context"
	"io"
	"strings"
	"testing"

	"github.com/shimasan0x00/difr/internal/claude"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// stubRunner is a minimal claude.Runner for option tests.
type stubRunner struct {
	output string
}

func (s *stubRunner) Run(_ context.Context, _ []string) (io.ReadCloser, error) {
	return io.NopCloser(strings.NewReader(s.output)), nil
}

func TestWithClaudeRunner_InjectsRunner(t *testing.T) {
	// Arrange
	runner := &stubRunner{output: "test"}
	dir := t.TempDir()

	// Act
	srv, err := New("", WithWorkDir(dir), WithClaudeRunner(runner))

	// Assert
	require.NoError(t, err)
	assert.Equal(t, claude.Runner(runner), srv.claudeRunner)
}

func TestWithClaudeRunner_OverridesAutoDetect(t *testing.T) {
	// WithClaudeRunner should take priority even without WithNoClaude
	runner := &stubRunner{output: "mock"}
	dir := t.TempDir()

	srv, err := New("", WithWorkDir(dir), WithClaudeRunner(runner))

	require.NoError(t, err)
	assert.Equal(t, claude.Runner(runner), srv.claudeRunner, "injected runner should be used")
}

func TestWithClaudeRunner_NilRunnerFallsBackToAutoDetect(t *testing.T) {
	dir := t.TempDir()

	// nil runner should not crash, auto-detect proceeds normally
	srv, err := New("", WithWorkDir(dir), WithClaudeRunner(nil), WithNoClaude(true))

	require.NoError(t, err)
	assert.Nil(t, srv.claudeRunner, "nil runner with noClaude should result in nil")
}
