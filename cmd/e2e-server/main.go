// Command e2e-server starts a production-mode difr server with mock Claude
// for Playwright E2E browser tests.
package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/shimasan0x00/difr/internal/server"
)

func main() {
	port := flag.Int("port", 4444, "Port to listen on")
	flag.Parse()

	if err := run(*port); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(port int) error {
	tmpDir, err := os.MkdirTemp("", "difr-e2e-*")
	if err != nil {
		return fmt.Errorf("creating temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	mock := newMockRunner(mockChatResponse, mockReviewResponse)

	srv, err := server.New(multiFileDiff,
		server.WithWorkDir(tmpDir),
		server.WithClaudeRunner(mock),
	)
	if err != nil {
		return fmt.Errorf("creating server: %w", err)
	}

	httpSrv := &http.Server{
		Addr:         fmt.Sprintf("127.0.0.1:%d", port),
		Handler:      srv.Handler(),
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		<-ctx.Done()
		slog.Info("Shutting down E2E server...")
		srv.CloseWebSockets()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		httpSrv.Shutdown(shutdownCtx)
	}()

	slog.Info("E2E server started", "addr", httpSrv.Addr)
	if err := httpSrv.ListenAndServe(); err != http.ErrServerClosed {
		return err
	}
	return nil
}
