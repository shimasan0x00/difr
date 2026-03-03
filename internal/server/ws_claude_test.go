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
{"type":"assistant","message":{"content":[{"type":"text","text":"Hello! How can I help?"}]},"session_id":"ws-test-session"}
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
{"type":"assistant","message":{"content":[{"type":"text","text":"[{\"filePath\":\"main.go\",\"line\":10,\"body\":\"Fix error handling\"}]"}]},"session_id":"review-session"}
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
	args, err := buildClaudeArgs("chat", "hello", "")

	require.NoError(t, err)
	assert.Equal(t, []string{"-p", "hello", "--output-format", "stream-json", "--verbose"}, args)
}

func TestBuildClaudeArgs_WithSessionID(t *testing.T) {
	args, err := buildClaudeArgs("chat", "hello", "sess-abc-123")

	require.NoError(t, err)
	assert.Equal(t, []string{"-r", "sess-abc-123", "-p", "hello", "--output-format", "stream-json", "--verbose"}, args)
}

func TestBuildClaudeArgs_Review(t *testing.T) {
	args, err := buildClaudeArgs("review", "review this", "")

	require.NoError(t, err)
	assert.Equal(t, []string{"-p", "review this", "--output-format", "stream-json", "--verbose", "--max-turns", "1"}, args)
}

func TestBuildClaudeArgs_RejectsEmptyContent(t *testing.T) {
	tests := []struct {
		name    string
		content string
		wantErr bool
	}{
		{"empty string", "", true},
		{"whitespace only", "   ", true},
		{"tab and newline", "\t\n", true},
		{"valid content", "hello", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := buildClaudeArgs("chat", tt.content, "")

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "content must not be empty")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestBuildClaudeArgs_RejectsUnknownType(t *testing.T) {
	tests := []struct {
		name    string
		msgType string
		wantErr bool
	}{
		{"chat", "chat", false},
		{"review", "review", false},
		{"empty", "", true},
		{"unknown", "foo", true},
		{"uppercase", "Chat", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := buildClaudeArgs(tt.msgType, "hello", "")

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "unknown message type")
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

func TestWSClaude_DoneSentWhenStreamEndsWithoutResult(t *testing.T) {
	// assistant text only — no "result" event. Should still receive "done".
	mockOutput := `{"type":"system","subtype":"init","session_id":"no-result-session"}
{"type":"assistant","message":{"content":[{"type":"text","text":"partial answer"}]},"session_id":"no-result-session"}
`
	s := newTestServerWithClaude(t, &mockRunner{output: mockOutput})
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

	assertHasResponseType(t, responses, "done", "expected done even without result event")
}

func TestWSClaude_DoneSentWhenStreamIsEmpty(t *testing.T) {
	// Empty output (immediate crash) — should still receive "done".
	s := newTestServerWithClaude(t, &mockRunner{output: ""})
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

	assertHasResponseType(t, responses, "done", "expected done even with empty stream")
}

// --- Session resume integration tests ---

// historyTrackingRunner returns different outputs for each call.
type historyTrackingRunner struct {
	outputs  []string
	callArgs [][]string
	callIdx  int
}

func (m *historyTrackingRunner) Run(_ context.Context, args []string) (io.ReadCloser, error) {
	m.callArgs = append(m.callArgs, args)
	idx := m.callIdx
	m.callIdx++
	if idx < len(m.outputs) {
		return io.NopCloser(strings.NewReader(m.outputs[idx])), nil
	}
	return io.NopCloser(strings.NewReader("")), nil
}

func TestWSClaude_SessionResumeOnSecondMessage(t *testing.T) {
	runner := &historyTrackingRunner{
		outputs: []string{
			// First response
			`{"type":"system","subtype":"init","session_id":"s1"}
{"type":"assistant","message":{"content":[{"type":"text","text":"Hello!"}]},"session_id":"s1"}
{"type":"result","subtype":"success","result":"Hello!","session_id":"s1","stop_reason":"end_turn"}
`,
			// Second response
			`{"type":"system","subtype":"init","session_id":"s1"}
{"type":"assistant","message":{"content":[{"type":"text","text":"I remember!"}]},"session_id":"s1"}
{"type":"result","subtype":"success","result":"I remember!","session_id":"s1","stop_reason":"end_turn"}
`,
		},
	}

	s := newTestServerWithClaude(t, runner)
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()

	wsURL := "ws" + strings.TrimPrefix(ts.URL, "http") + "/ws/claude"
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, _, err := websocket.Dial(ctx, wsURL, nil)
	require.NoError(t, err)
	defer conn.CloseNow()

	// First message
	require.NoError(t, wsjson.Write(ctx, conn, ChatMessage{Type: "chat", Content: "Hello"}))
	readWSResponses(t, ctx, conn)

	// Second message — should use -r with extracted session ID
	require.NoError(t, wsjson.Write(ctx, conn, ChatMessage{Type: "chat", Content: "Remember me?"}))
	readWSResponses(t, ctx, conn)

	// Verify first call has no -r flag
	require.Len(t, runner.callArgs, 2)
	assert.NotContains(t, runner.callArgs[0], "-r")

	// Verify second call has -r flag with session ID from first call
	assert.Contains(t, runner.callArgs[1], "-r")
	assert.Contains(t, runner.callArgs[1], "s1")

	// Verify prompt is plain content, no history injection
	secondPrompt := runner.callArgs[1][3] // args: [-r, s1, -p, <prompt>, ...]
	assert.Equal(t, "Remember me?", secondPrompt)
}

func TestWSClaude_ClearResetsSessionID(t *testing.T) {
	runner := &historyTrackingRunner{
		outputs: []string{
			// First response
			`{"type":"system","subtype":"init","session_id":"s1"}
{"type":"assistant","message":{"content":[{"type":"text","text":"Hello!"}]},"session_id":"s1"}
{"type":"result","subtype":"success","result":"Hello!","session_id":"s1","stop_reason":"end_turn"}
`,
			// Response after clear
			`{"type":"system","subtype":"init","session_id":"s2"}
{"type":"assistant","message":{"content":[{"type":"text","text":"Fresh start!"}]},"session_id":"s2"}
{"type":"result","subtype":"success","result":"Fresh start!","session_id":"s2","stop_reason":"end_turn"}
`,
		},
	}

	s := newTestServerWithClaude(t, runner)
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()

	wsURL := "ws" + strings.TrimPrefix(ts.URL, "http") + "/ws/claude"
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, _, err := websocket.Dial(ctx, wsURL, nil)
	require.NoError(t, err)
	defer conn.CloseNow()

	// First message to establish session
	require.NoError(t, wsjson.Write(ctx, conn, ChatMessage{Type: "chat", Content: "Hello"}))
	readWSResponses(t, ctx, conn)

	// Send clear
	require.NoError(t, wsjson.Write(ctx, conn, ChatMessage{Type: "clear"}))
	clearResponses := readWSResponses(t, ctx, conn)
	assertHasResponseType(t, clearResponses, "done", "clear should return done")

	// Message after clear — should NOT have -r flag (new session)
	require.NoError(t, wsjson.Write(ctx, conn, ChatMessage{Type: "chat", Content: "New conversation"}))
	readWSResponses(t, ctx, conn)

	// Verify the post-clear call has no -r flag
	require.Len(t, runner.callArgs, 2)
	assert.NotContains(t, runner.callArgs[1], "-r")
	assert.Equal(t, "New conversation", runner.callArgs[1][1]) // -p <prompt> at index 1
}

// errorAfterDataReader returns initial data, then an error on subsequent reads.
type errorAfterDataReader struct {
	data    string
	readErr error
	done    bool
}

func (r *errorAfterDataReader) Read(p []byte) (int, error) {
	if r.done {
		return 0, r.readErr
	}
	r.done = true
	n := copy(p, r.data)
	return n, nil
}

func (r *errorAfterDataReader) Close() error { return nil }

// errorAfterInitRunner returns an errorAfterDataReader on the first call,
// then normal output on subsequent calls.
type errorAfterInitRunner struct {
	initData string
	readErr  error
	outputs  []string
	callArgs [][]string
	callIdx  int
}

func (m *errorAfterInitRunner) Run(_ context.Context, args []string) (io.ReadCloser, error) {
	m.callArgs = append(m.callArgs, args)
	idx := m.callIdx
	m.callIdx++
	if idx == 0 {
		return &errorAfterDataReader{data: m.initData, readErr: m.readErr}, nil
	}
	outIdx := idx - 1
	if outIdx < len(m.outputs) {
		return io.NopCloser(strings.NewReader(m.outputs[outIdx])), nil
	}
	return io.NopCloser(strings.NewReader("")), nil
}

func TestWSClaude_SessionIDPreservedOnStreamError(t *testing.T) {
	runner := &errorAfterInitRunner{
		initData: "{\"type\":\"system\",\"subtype\":\"init\",\"session_id\":\"err-session\"}\n",
		readErr:  fmt.Errorf("connection reset"),
		outputs: []string{
			// Second call: normal response
			`{"type":"system","subtype":"init","session_id":"err-session"}
{"type":"assistant","message":{"content":[{"type":"text","text":"Recovered!"}]},"session_id":"err-session"}
{"type":"result","subtype":"success","result":"Recovered!","session_id":"err-session","stop_reason":"end_turn"}
`,
		},
	}

	s := newTestServerWithClaude(t, runner)
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()

	wsURL := "ws" + strings.TrimPrefix(ts.URL, "http") + "/ws/claude"
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, _, err := websocket.Dial(ctx, wsURL, nil)
	require.NoError(t, err)
	defer conn.CloseNow()

	// First message: init succeeds but stream errors out
	require.NoError(t, wsjson.Write(ctx, conn, ChatMessage{Type: "chat", Content: "Hello"}))
	responses := readWSResponses(t, ctx, conn)
	assertHasResponseType(t, responses, "error", "expected error from stream failure")

	// Second message: should resume with -r flag despite prior error
	require.NoError(t, wsjson.Write(ctx, conn, ChatMessage{Type: "chat", Content: "Try again"}))
	readWSResponses(t, ctx, conn)

	// Verify second call uses -r with the session ID extracted before the error
	require.Len(t, runner.callArgs, 2)
	assert.Contains(t, runner.callArgs[1], "-r")
	assert.Contains(t, runner.callArgs[1], "err-session")
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
