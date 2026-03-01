package server

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"regexp"
	"time"

	"github.com/coder/websocket"
	"github.com/shimasan0x00/difr/internal/claude"
)

const pingInterval = 30 * time.Second

// validSessionID matches UUID-like or alphanumeric session IDs from Claude CLI.
var validSessionID = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_-]{0,127}$`)

// ChatMessage is a message sent from the client via WebSocket.
type ChatMessage struct {
	Type      string `json:"type"`      // "chat" or "review"
	Content   string `json:"content"`
	SessionID string `json:"sessionId,omitempty"`
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

			if s.claudeRunner == nil {
				writeWS(ctx, conn, WSResponse{Type: "error", Error: "Claude CLI not available"})
				continue
			}

			args, err := buildClaudeArgs(msg)
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

			streamErr := func() error {
				defer reader.Close()
				defer claudeCancel()
				return streamClaudeEvents(ctx, conn, reader)
			}()
			if streamErr != nil {
				writeWS(ctx, conn, WSResponse{Type: "error", Error: streamErr.Error()})
			}
		}
	})
}

func buildClaudeArgs(msg ChatMessage) ([]string, error) {
	args := []string{"-p", msg.Content, "--output-format", "stream-json"}

	if msg.SessionID != "" {
		if !validSessionID.MatchString(msg.SessionID) {
			return nil, fmt.Errorf("invalid session ID: %q", msg.SessionID)
		}
		args = append([]string{"-r", msg.SessionID}, args...)
	}

	if msg.Type == "review" {
		args = append(args, "--max-turns", "1")
	}

	return args, nil
}

// streamClaudeEvents reads NDJSON from reader line-by-line and sends
// each event to the WebSocket immediately, enabling real-time streaming.
func streamClaudeEvents(ctx context.Context, conn *websocket.Conn, reader io.Reader) error {
	const maxLineSize = 1024 * 1024 // 1MB
	scanner := bufio.NewScanner(reader)
	scanner.Buffer(make([]byte, 0, maxLineSize), maxLineSize)

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
				if err := writeWS(ctx, conn, WSResponse{
					Type:      "session",
					SessionID: event.SessionID,
				}); err != nil {
					return err
				}
			}
		case "assistant":
			for _, block := range event.Content {
				if block.Type == "text" {
					if err := writeWS(ctx, conn, WSResponse{
						Type:    "text",
						Content: block.Text,
					}); err != nil {
						return err
					}
				}
			}
		case "result":
			if err := writeWS(ctx, conn, WSResponse{
				Type:      "done",
				Content:   event.Result,
				SessionID: event.SessionID,
			}); err != nil {
				return err
			}
		}
	}
	return scanner.Err()
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
