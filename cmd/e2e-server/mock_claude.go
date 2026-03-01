package main

import (
	"context"
	"io"
	"strings"
	"sync"
)

// mockRunner implements claude.Runner with fixed responses for E2E testing.
type mockRunner struct {
	mu           sync.Mutex
	chatOutput   string
	reviewOutput string
}

func newMockRunner(chatOutput, reviewOutput string) *mockRunner {
	return &mockRunner{
		chatOutput:   chatOutput,
		reviewOutput: reviewOutput,
	}
}

// Run returns mock NDJSON output. It inspects args for "--max-turns" to
// distinguish review requests from chat requests (mirrors buildClaudeArgs).
func (m *mockRunner) Run(_ context.Context, args []string) (io.ReadCloser, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	output := m.chatOutput
	for _, arg := range args {
		if arg == "--max-turns" {
			output = m.reviewOutput
			break
		}
	}
	return io.NopCloser(strings.NewReader(output)), nil
}
