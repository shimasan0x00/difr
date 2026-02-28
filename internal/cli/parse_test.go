package cli

import (
	"io"
	"strings"
	"testing"

	"github.com/shimasan0x00/diffff/internal/diff"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseDiffRequest(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		stdin    io.Reader
		wantMode diff.DiffMode
		wantFrom string
		wantTo   string
	}{
		{
			name:     "no args defaults to latest commit",
			args:     []string{},
			stdin:    nil,
			wantMode: diff.DiffModeLatestCommit,
		},
		{
			name:     "staged keyword",
			args:     []string{"staged"},
			stdin:    nil,
			wantMode: diff.DiffModeStaged,
		},
		{
			name:     "working keyword",
			args:     []string{"working"},
			stdin:    nil,
			wantMode: diff.DiffModeWorking,
		},
		{
			name:     "single commit hash",
			args:     []string{"abc1234"},
			stdin:    nil,
			wantMode: diff.DiffModeCommit,
			wantFrom: "abc1234",
		},
		{
			name:     "two commits specify range",
			args:     []string{"abc1234", "def5678"},
			stdin:    nil,
			wantMode: diff.DiffModeRange,
			wantFrom: "abc1234",
			wantTo:   "def5678",
		},
		{
			name:     "stdin with content",
			args:     []string{},
			stdin:    strings.NewReader("diff --git a/file.go b/file.go\n"),
			wantMode: diff.DiffModeStdin,
		},
		{
			name:     "stdin takes precedence over args",
			args:     []string{"abc1234"},
			stdin:    strings.NewReader("some diff content"),
			wantMode: diff.DiffModeStdin,
		},
		{
			name:     "nil stdin reader defaults to latest commit",
			args:     []string{},
			stdin:    nil,
			wantMode: diff.DiffModeLatestCommit,
		},
		{
			name:     "empty stdin reader treated as stdin mode",
			args:     []string{},
			stdin:    strings.NewReader(""),
			wantMode: diff.DiffModeStdin,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := ParseDiffRequest(tt.args, tt.stdin)

			require.NoError(t, err)
			assert.Equal(t, tt.wantMode, req.Mode)
			if tt.wantFrom != "" {
				assert.Equal(t, tt.wantFrom, req.From)
			}
			if tt.wantTo != "" {
				assert.Equal(t, tt.wantTo, req.To)
			}
		})
	}
}

func TestParseDiffRequest_WhitespaceOnlyStdinTreatedAsStdinMode(t *testing.T) {
	req, err := ParseDiffRequest([]string{}, strings.NewReader("   \n\t\n  "))

	require.NoError(t, err)
	assert.Equal(t, diff.DiffModeStdin, req.Mode)
}

func TestParseDiffRequest_TooManyArgs(t *testing.T) {
	_, err := ParseDiffRequest([]string{"a", "b", "c"}, nil)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "too many arguments")
}
