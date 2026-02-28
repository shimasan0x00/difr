package claude

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func readTestDataStream(t *testing.T, name string) string {
	t.Helper()
	data, err := os.ReadFile("testdata/" + name)
	require.NoError(t, err)
	return string(data)
}

func TestParseStreamEvents_ParsesAllEventTypes(t *testing.T) {
	input := readTestDataStream(t, "stream_chat.jsonl")

	events, err := ParseStreamEvents(strings.NewReader(input))

	require.NoError(t, err)
	require.Len(t, events, 3)

	// system init
	assert.Equal(t, "system", events[0].Type)
	assert.Equal(t, "init", events[0].SubType)
	assert.Equal(t, "test-session-123", events[0].SessionID)

	// assistant
	assert.Equal(t, "assistant", events[1].Type)
	require.NotEmpty(t, events[1].Content)
	assert.Equal(t, "Here is my review of the code.", events[1].Content[0].Text)

	// result
	assert.Equal(t, "result", events[2].Type)
	assert.Equal(t, "success", events[2].SubType)
	assert.Equal(t, "Here is my review of the code.", events[2].Result)
}

func TestParseStreamEvents_DetectsErrorResult(t *testing.T) {
	input := readTestDataStream(t, "stream_error.jsonl")

	events, err := ParseStreamEvents(strings.NewReader(input))

	require.NoError(t, err)
	last := events[len(events)-1]
	assert.Equal(t, "error_max_turns", last.SubType)
}

func TestParseStreamEvents_SkipsInvalidJSONLines(t *testing.T) {
	input := `{"type":"system","subtype":"init","session_id":"s1"}
not valid json
{"type":"result","subtype":"success","result":"done","session_id":"s1","stop_reason":"end_turn"}
`

	events, err := ParseStreamEvents(strings.NewReader(input))

	require.NoError(t, err)
	assert.Len(t, events, 2, "invalid JSON line should be skipped")
}

func TestParseStreamEvents_EmptyInputReturnsNoEvents(t *testing.T) {
	events, err := ParseStreamEvents(strings.NewReader(""))

	require.NoError(t, err)
	assert.Empty(t, events)
}

func TestExtractSessionID_ReturnsIDFromSystemEvent(t *testing.T) {
	input := readTestDataStream(t, "stream_chat.jsonl")
	events, err := ParseStreamEvents(strings.NewReader(input))
	require.NoError(t, err)

	sessionID := ExtractSessionID(events)

	assert.Equal(t, "test-session-123", sessionID)
}

func TestExtractResultText_ReturnsTextFromResultEvent(t *testing.T) {
	input := readTestDataStream(t, "stream_chat.jsonl")
	events, err := ParseStreamEvents(strings.NewReader(input))
	require.NoError(t, err)

	text := ExtractResultText(events)

	assert.Equal(t, "Here is my review of the code.", text)
}
