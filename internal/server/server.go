package server

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/coder/websocket"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/shimasan0x00/difr/internal/claude"
	"github.com/shimasan0x00/difr/internal/comment"
	"github.com/shimasan0x00/difr/internal/diff"
	feembed "github.com/shimasan0x00/difr/internal/embed"
)

// Server holds the HTTP server configuration and dependencies.
type Server struct {
	router         chi.Router
	diffResult     *diff.DiffResult
	// fileIndex is built once during initialization and never modified after.
	// Concurrent reads without a lock are safe.
	fileIndex      map[string]*diff.DiffFile
	commentStore   *comment.Store
	claudeRunner   claude.Runner
	viewMode       string
	claudeTimeout  time.Duration

	// wsConns tracks active WebSocket connections for graceful shutdown.
	wsMu    sync.Mutex
	wsConns map[*websocket.Conn]struct{}
}

// Option configures a Server.
type Option func(*serverConfig)

type serverConfig struct {
	workDir       string
	noClaude      bool
	viewMode      string
	claudeTimeout time.Duration
}

const defaultClaudeTimeout = 5 * time.Minute

// WithWorkDir sets the working directory for comment storage.
func WithWorkDir(dir string) Option {
	return func(c *serverConfig) { c.workDir = dir }
}

// WithNoClaude disables Claude integration.
func WithNoClaude(noClaude bool) Option {
	return func(c *serverConfig) { c.noClaude = noClaude }
}

// WithViewMode sets the default view mode.
func WithViewMode(mode string) Option {
	return func(c *serverConfig) { c.viewMode = mode }
}

// WithClaudeTimeout sets the timeout for Claude CLI operations.
func WithClaudeTimeout(d time.Duration) Option {
	return func(c *serverConfig) { c.claudeTimeout = d }
}

// New creates a new server with the given raw diff content.
func New(rawDiff string, opts ...Option) (*Server, error) {
	cfg := &serverConfig{
		viewMode: "split",
	}
	for _, opt := range opts {
		opt(cfg)
	}

	// Default to cwd if no workDir specified
	if cfg.workDir == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("getting working directory: %w", err)
		}
		cfg.workDir = cwd
	}

	result, err := diff.Parse(rawDiff)
	if err != nil {
		return nil, fmt.Errorf("parsing diff: %w", err)
	}

	commentPath := filepath.Join(cfg.workDir, ".difr", "comments.json")
	cs := comment.NewStore(commentPath)
	if err := cs.Load(); err != nil {
		return nil, fmt.Errorf("loading comments: %w", err)
	}

	fileIdx := make(map[string]*diff.DiffFile, len(result.Files))
	for i := range result.Files {
		f := &result.Files[i]
		if f.NewPath != "" {
			fileIdx[f.NewPath] = f
		}
		if f.OldPath != "" && f.OldPath != f.NewPath {
			fileIdx[f.OldPath] = f
		}
	}

	claudeTimeout := cfg.claudeTimeout
	if claudeTimeout == 0 {
		claudeTimeout = defaultClaudeTimeout
	}

	s := &Server{
		diffResult:    result,
		fileIndex:     fileIdx,
		commentStore:  cs,
		viewMode:      cfg.viewMode,
		claudeTimeout: claudeTimeout,
		wsConns:       make(map[*websocket.Conn]struct{}),
	}

	// Claude CLI integration (non-fatal if unavailable)
	if !cfg.noClaude {
		if client, err := claude.NewClient(); err == nil {
			s.claudeRunner = client
		} else {
			slog.Info("Claude CLI not available", "err", err)
		}
	}

	if err := s.setupRoutes(); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *Server) setupRoutes() error {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Route("/api", func(r chi.Router) {
		r.Get("/health", s.handleHealthCheck)

		r.Get("/diff", s.handleGetDiff)
		r.Get("/diff/files", s.handleGetDiffFiles)
		r.Get("/diff/files/*", s.handleGetDiffFileByPath)
		r.Get("/diff/stats", s.handleGetDiffStats)
		// View mode is read-only on the server; the frontend manages mode switching via local state.
		r.Get("/diff/mode", s.handleGetViewMode)

		r.Post("/comments", s.handleCreateComment)
		r.Get("/comments", s.handleListComments)
		r.Put("/comments/{id}", s.handleUpdateComment)
		r.Delete("/comments/{id}", s.handleDeleteComment)
		r.Get("/comments/export", s.handleExportComments)

		r.Get("/claude/status", s.handleClaudeStatus)
	})

	r.Handle("/ws/claude", s.handleWSClaude())

	// Serve frontend (embedded in production, reverse proxy in dev)
	feHandler, err := feembed.Handler()
	if err != nil {
		return fmt.Errorf("frontend handler: %w", err)
	}
	r.Handle("/*", feHandler)

	s.router = r
	return nil
}

// handleHealthCheck returns a simple health status for monitoring.
func (s *Server) handleHealthCheck(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// handleGetViewMode returns the server's configured view mode.
func (s *Server) handleGetViewMode(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"mode": s.viewMode})
}

// handleClaudeStatus returns whether Claude CLI is available.
func (s *Server) handleClaudeStatus(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]bool{"available": s.claudeRunner != nil})
}

// Handler returns the HTTP handler.
func (s *Server) Handler() http.Handler {
	return s.router
}

// registerWSConn tracks a WebSocket connection for graceful shutdown.
func (s *Server) registerWSConn(conn *websocket.Conn) {
	s.wsMu.Lock()
	s.wsConns[conn] = struct{}{}
	s.wsMu.Unlock()
}

// unregisterWSConn removes a WebSocket connection from tracking.
func (s *Server) unregisterWSConn(conn *websocket.Conn) {
	s.wsMu.Lock()
	delete(s.wsConns, conn)
	s.wsMu.Unlock()
}

// CloseWebSockets gracefully closes all active WebSocket connections.
func (s *Server) CloseWebSockets() {
	s.wsMu.Lock()
	conns := make([]*websocket.Conn, 0, len(s.wsConns))
	for conn := range s.wsConns {
		conns = append(conns, conn)
	}
	s.wsMu.Unlock()

	for _, conn := range conns {
		conn.Close(websocket.StatusGoingAway, "server shutting down")
	}
}
