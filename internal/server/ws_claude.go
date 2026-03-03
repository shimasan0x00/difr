package server

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/coder/websocket"
	"github.com/shimasan0x00/difr/internal/claude"
)

const pingInterval = 30 * time.Second

// ChatMessage is a message sent from the client via WebSocket.
type ChatMessage struct {
	Type    string `json:"type"` // "chat", "review", or "clear"
	Content string `json:"content"`
}

// WSResponse is a message sent to the client via WebSocket.
type WSResponse struct {
	Type      string `json:"type"`                // "session", "text", "done", "error"
	Content   string `json:"content,omitempty"`
	SessionID string `json:"sessionId,omitempty"`
	Error     string `json:"error,omitempty"`
}

func (s *Server) handleWSClaude() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
			OriginPatterns: []string{"localhost:*", "127.0.0.1:*"},
		})
		if err != nil {
			slog.Error("websocket accept error", "err", err)
			return
		}
		defer conn.CloseNow()
		conn.SetReadLimit(1 << 20) // 1MB max message size

		s.registerWSConn(conn)
		defer s.unregisterWSConn(conn)

		ctx := r.Context()

		// Start ping/pong keepalive to prevent proxy/NAT timeout
		pingCtx, pingCancel := context.WithCancel(ctx)
		defer pingCancel()
		go func() {
			ticker := time.NewTicker(pingInterval)
			defer ticker.Stop()
			for {
				select {
				case <-pingCtx.Done():
					return
				case <-ticker.C:
					if err := conn.Ping(pingCtx); err != nil {
						return
					}
				}
			}
		}()

		var sessionID string

		for {
			_, data, err := conn.Read(ctx)
			if err != nil {
				return
			}

			var msg ChatMessage
			if err := json.Unmarshal(data, &msg); err != nil {
				writeWS(ctx, conn, WSResponse{Type: "error", Error: "invalid message"})
				continue
			}

			// Handle clear message: reset session and acknowledge
			if msg.Type == "clear" {
				sessionID = ""
				writeWS(ctx, conn, WSResponse{Type: "done"})
				continue
			}

			if s.claudeRunner == nil {
				writeWS(ctx, conn, WSResponse{Type: "error", Error: "Claude CLI not available"})
				continue
			}

			args, err := buildClaudeArgs(msg.Type, msg.Content, sessionID)
			if err != nil {
				writeWS(ctx, conn, WSResponse{Type: "error", Error: err.Error()})
				continue
			}

			claudeCtx, claudeCancel := context.WithTimeout(ctx, s.claudeTimeout)
			reader, err := s.claudeRunner.Run(claudeCtx, args)
			if err != nil {
				claudeCancel()
				writeWS(ctx, conn, WSResponse{Type: "error", Error: err.Error()})
				continue
			}

			doneSent, extractedSessionID, streamErr := func() (bool, string, error) {
				defer reader.Close()
				defer claudeCancel()
				return streamClaudeEvents(ctx, conn, reader)
			}()
			if extractedSessionID != "" {
				sessionID = extractedSessionID
			}
			if streamErr != nil {
				writeWS(ctx, conn, WSResponse{Type: "error", Error: streamErr.Error()})
			} else if !doneSent {
				writeWS(ctx, conn, WSResponse{Type: "done"})
			}
		}
	})
}

func buildClaudeArgs(msgType, content, sessionID string) ([]string, error) {
	if msgType != "chat" && msgType != "review" {
		return nil, fmt.Errorf("unknown message type: %q", msgType)
	}

	if strings.TrimSpace(content) == "" {
		return nil, fmt.Errorf("content must not be empty")
	}

	args := []string{"-p", content, "--output-format", "stream-json", "--verbose"}

	if sessionID != "" {
		args = append([]string{"-r", sessionID}, args...)
	}

	if msgType == "review" {
		args = append(args, "--max-turns", "1")
	}

	return args, nil
}

// streamClaudeEvents reads NDJSON from reader line-by-line and sends
// each event to the WebSocket immediately, enabling real-time streaming.
// Returns (doneSent, extractedSessionID, error).
func streamClaudeEvents(ctx context.Context, conn *websocket.Conn, reader io.Reader) (bool, string, error) {
	const maxLineSize = 1024 * 1024 // 1MB
	scanner := bufio.NewScanner(reader)
	scanner.Buffer(make([]byte, 4096), maxLineSize)

	var doneSent bool
	var extractedSessionID string

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var event claude.StreamEvent
		if err := json.Unmarshal(line, &event); err != nil {
			slog.Warn("stream parse skip", "err", err)
			continue
		}

		switch event.Type {
		case "system":
			if event.SubType == "init" {
				extractedSessionID = event.SessionID
				if err := writeWS(ctx, conn, WSResponse{
					Type:      "session",
					SessionID: event.SessionID,
				}); err != nil {
					return doneSent, extractedSessionID, err
				}
			}
		case "assistant":
			for _, block := range event.ContentBlocks() {
				if block.Type == "text" {
					if err := writeWS(ctx, conn, WSResponse{
						Type:    "text",
						Content: block.Text,
					}); err != nil {
						return doneSent, extractedSessionID, err
					}
				}
			}
		case "result":
			doneSent = true
			if err := writeWS(ctx, conn, WSResponse{
				Type:      "done",
				Content:   event.Result,
				SessionID: event.SessionID,
			}); err != nil {
				return doneSent, extractedSessionID, err
			}
		}
	}
	return doneSent, extractedSessionID, scanner.Err()
}

func writeWS(ctx context.Context, conn *websocket.Conn, resp WSResponse) error {
	data, err := json.Marshal(resp)
	if err != nil {
		slog.Error("writeWS marshal error", "err", err)
		return err
	}
	if err := conn.Write(ctx, websocket.MessageText, data); err != nil {
		slog.Error("writeWS write error", "err", err)
		return err
	}
	return nil
}
