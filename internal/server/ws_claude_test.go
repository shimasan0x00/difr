package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/shimasan0x00/difr/internal/claude"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockRunner struct {
	output   string
	err      error
	lastArgs []string
}

func (m *mockRunner) Run(_ context.Context, args []string) (io.ReadCloser, error) {
	m.lastArgs = args
	if m.err != nil {
		return nil, m.err
	}
	return io.NopCloser(strings.NewReader(m.output)), nil
}

func TestWSClaude_ChatFlow(t *testing.T) {
	mockOutput := `{"type":"system","subtype":"init","session_id":"ws-test-session"}
{"type":"assistant","content":[{"type":"text","text":"Hello! How can I help?"}]}
{"type":"result","subtype":"success","result":"Hello! How can I help?","session_id":"ws-test-session","stop_reason":"end_turn"}
`
	s := newTestServerWithClaude(t, &mockRunner{output: mockOutput})
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()

	wsURL := "ws" + strings.TrimPrefix(ts.URL, "http") + "/ws/claude"
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, _, err := websocket.Dial(ctx, wsURL, nil)
	if err != nil {
		t.Fatalf("WebSocket dial failed: %v", err)
	}
	defer conn.CloseNow()

	// Send chat message
	msg := ChatMessage{
		Type:    "chat",
		Content: "Hello",
	}
	if err := wsjson.Write(ctx, conn, msg); err != nil {
		t.Fatalf("write failed: %v", err)
	}

	// Read responses
	responses := readWSResponses(t, ctx, conn)

	require.NotEmpty(t, responses, "expected at least 1 response")

	// Should have session_id in init response
	assertHasResponseType(t, responses, "session", "expected session init response")
	assertHasResponseType(t, responses, "done", "expected done response")

	// Verify session ID is populated
	for _, r := range responses {
		if r.Type == "session" {
			assert.NotEmpty(t, r.SessionID, "session response should have session_id")
		}
	}
}

func TestWSClaude_ReviewFlow(t *testing.T) {
	mockOutput := `{"type":"system","subtype":"init","session_id":"review-session"}
{"type":"assistant","content":[{"type":"text","text":"[{\"filePath\":\"main.go\",\"line\":10,\"body\":\"Fix error handling\"}]"}]}
{"type":"result","subtype":"success","result":"[{\"filePath\":\"main.go\",\"line\":10,\"body\":\"Fix error handling\"}]","session_id":"review-session","stop_reason":"end_turn"}
`
	s := newTestServerWithClaude(t, &mockRunner{output: mockOutput})
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()

	wsURL := "ws" + strings.TrimPrefix(ts.URL, "http") + "/ws/claude"
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, _, err := websocket.Dial(ctx, wsURL, nil)
	if err != nil {
		t.Fatalf("WebSocket dial failed: %v", err)
	}
	defer conn.CloseNow()

	// Send review request
	msg := ChatMessage{
		Type:    "review",
		Content: "Review this diff",
	}
	require.NoError(t, wsjson.Write(ctx, conn, msg))

	// Read until done
	responses := readWSResponses(t, ctx, conn)

	assertHasResponseType(t, responses, "done", "expected done response")
}

func newTestServerWithClaude(t *testing.T, runner claude.Runner) *Server {
	t.Helper()
	dir := t.TempDir()
	s, err := New("", WithWorkDir(dir), WithNoClaude(true))
	require.NoError(t, err)
	s.claudeRunner = runner
	return s
}

// readWSResponses reads WebSocket responses until "done" or "error" type is received.
func readWSResponses(t *testing.T, ctx context.Context, conn *websocket.Conn) []WSResponse {
	t.Helper()
	var responses []WSResponse
	for {
		_, data, err := conn.Read(ctx)
		if err != nil {
			break
		}
		var resp WSResponse
		require.NoError(t, json.Unmarshal(data, &resp), "WebSocket response should be valid JSON")
		responses = append(responses, resp)
		if resp.Type == "done" || resp.Type == "error" {
			break
		}
	}
	return responses
}

func TestBuildClaudeArgs_Chat(t *testing.T) {
	msg := ChatMessage{Type: "chat", Content: "hello"}

	args, err := buildClaudeArgs(msg)

	require.NoError(t, err)
	assert.Equal(t, []string{"-p", "hello", "--output-format", "stream-json"}, args)
}

func TestBuildClaudeArgs_ChatWithSession(t *testing.T) {
	msg := ChatMessage{Type: "chat", Content: "hello", SessionID: "sess-123"}

	args, err := buildClaudeArgs(msg)

	require.NoError(t, err)
	assert.Equal(t, []string{"-r", "sess-123", "-p", "hello", "--output-format", "stream-json"}, args)
}

func TestBuildClaudeArgs_Review(t *testing.T) {
	msg := ChatMessage{Type: "review", Content: "review this"}

	args, err := buildClaudeArgs(msg)

	require.NoError(t, err)
	assert.Equal(t, []string{"-p", "review this", "--output-format", "stream-json", "--max-turns", "1"}, args)
}

func TestBuildClaudeArgs_RejectsInvalidSessionID(t *testing.T) {
	tests := []struct {
		name      string
		sessionID string
		wantErr   bool
	}{
		{"valid alphanumeric", "abc123", false},
		{"valid UUID-like", "sess-abc-123", false},
		{"dash prefix", "-malicious", true},
		{"empty string", "", false},
		{"contains spaces", "sess 123", true},
		{"starts with underscore", "_sess", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := ChatMessage{Type: "chat", Content: "hello", SessionID: tt.sessionID}
			_, err := buildClaudeArgs(msg)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestWSClaude_RunnerError(t *testing.T) {
	runner := &mockRunner{err: fmt.Errorf("CLI not found")}
	s := newTestServerWithClaude(t, runner)
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()

	wsURL := "ws" + strings.TrimPrefix(ts.URL, "http") + "/ws/claude"
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, _, err := websocket.Dial(ctx, wsURL, nil)
	require.NoError(t, err)
	defer conn.CloseNow()

	msg := ChatMessage{Type: "chat", Content: "hello"}
	require.NoError(t, wsjson.Write(ctx, conn, msg))

	responses := readWSResponses(t, ctx, conn)

	assertHasResponseType(t, responses, "error", "expected error response when runner fails")
	for _, r := range responses {
		if r.Type == "error" {
			assert.Contains(t, r.Error, "CLI not found")
		}
	}
}

func TestStreamClaudeEvents_ClosesDeferredOnPanic(t *testing.T) {
	// Verify that reader.Close() is called even if the function returns an error
	var closed atomic.Bool
	reader := &trackingCloser{
		Reader: strings.NewReader(`{"type":"result","result":"done","session_id":"s1"}` + "\n"),
		onClose: func() {
			closed.Store(true)
		},
	}

	s := newTestServerWithClaude(t, &closerTrackingRunner{reader: reader})
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()

	wsURL := "ws" + strings.TrimPrefix(ts.URL, "http") + "/ws/claude"
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, _, err := websocket.Dial(ctx, wsURL, nil)
	require.NoError(t, err)
	defer conn.CloseNow()

	msg := ChatMessage{Type: "chat", Content: "hello"}
	require.NoError(t, wsjson.Write(ctx, conn, msg))

	readWSResponses(t, ctx, conn)

	assert.True(t, closed.Load(), "reader.Close() should be called via defer")
}

// trackingCloser is an io.ReadCloser that tracks whether Close() was called.
type trackingCloser struct {
	io.Reader
	onClose func()
}

func (tc *trackingCloser) Close() error {
	tc.onClose()
	return nil
}

// closerTrackingRunner returns the provided ReadCloser.
type closerTrackingRunner struct {
	reader io.ReadCloser
}

func (r *closerTrackingRunner) Run(_ context.Context, _ []string) (io.ReadCloser, error) {
	return r.reader, nil
}

func TestWSClaude_RejectsMessageExceedingReadLimit(t *testing.T) {
	s := newTestServerWithClaude(t, &mockRunner{output: ""})
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()

	wsURL := "ws" + strings.TrimPrefix(ts.URL, "http") + "/ws/claude"
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, _, err := websocket.Dial(ctx, wsURL, nil)
	require.NoError(t, err)
	defer conn.CloseNow()

	// Send a message larger than the 1MB read limit
	largeContent := strings.Repeat("x", 1<<20+1) // 1MB + 1 byte
	msg := []byte(fmt.Sprintf(`{"type":"chat","content":"%s"}`, largeContent))

	err = conn.Write(ctx, websocket.MessageText, msg)
	if err != nil {
		// Write may fail if the server closes the connection fast enough
		return
	}

	// The server should close the connection after receiving an oversized message
	_, _, err = conn.Read(ctx)
	assert.Error(t, err, "connection should be closed after oversized message")
}

func TestBuildClaudeArgs_SessionIDBoundaryValues(t *testing.T) {
	tests := []struct {
		name      string
		sessionID string
		wantErr   bool
	}{
		{"single character", "a", false},
		{"128 characters (max length)", strings.Repeat("a", 128), false},
		{"129 characters (exceeds max)", strings.Repeat("a", 129), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := ChatMessage{Type: "chat", Content: "hello", SessionID: tt.sessionID}

			_, err := buildClaudeArgs(msg)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// assertHasResponseType checks that at least one response has the given type.
func assertHasResponseType(t *testing.T, responses []WSResponse, respType string, msgAndArgs ...interface{}) {
	t.Helper()
	for _, r := range responses {
		if r.Type == respType {
			return
		}
	}
	assert.Fail(t, "response type not found: "+respType, msgAndArgs...)
}
