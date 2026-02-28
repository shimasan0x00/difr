package claude

import (
	"bufio"
	"encoding/json"
	"io"
)

// ContentBlock represents a content block in an assistant message.
type ContentBlock struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
}

// StreamEvent represents a single event from Claude Code's stream-json output.
type StreamEvent struct {
	Type       string         `json:"type"`
	SubType    string         `json:"subtype,omitempty"`
	SessionID  string         `json:"session_id,omitempty"`
	Content    []ContentBlock `json:"content,omitempty"`
	Result     string         `json:"result,omitempty"`
	StopReason string         `json:"stop_reason,omitempty"`
}

// ParseStreamEvents parses NDJSON stream-json output into StreamEvent slices.
// Invalid JSON lines are silently skipped.
func ParseStreamEvents(r io.Reader) ([]StreamEvent, error) {
	const maxLineSize = 1024 * 1024 // 1MB per line to handle large Claude responses
	var events []StreamEvent
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, maxLineSize), maxLineSize)
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		var event StreamEvent
		if err := json.Unmarshal(line, &event); err != nil {
			// Skip invalid JSON lines
			continue
		}
		events = append(events, event)
	}
	return events, scanner.Err()
}

// ExtractSessionID returns the session ID from the init event.
func ExtractSessionID(events []StreamEvent) string {
	for _, e := range events {
		if e.Type == "system" && e.SubType == "init" {
			return e.SessionID
		}
	}
	return ""
}

// ExtractResultText returns the result text from the result event.
func ExtractResultText(events []StreamEvent) string {
	for _, e := range events {
		if e.Type == "result" {
			return e.Result
		}
	}
	return ""
}
