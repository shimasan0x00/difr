package cli

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/shimasan0x00/diffff/internal/git"
	"github.com/shimasan0x00/diffff/internal/server"
	"github.com/shimasan0x00/diffff/internal/watcher"
	"github.com/spf13/cobra"
)

// NewRootCmd creates the root cobra command.
func NewRootCmd() *cobra.Command {
	var cfg Config

	rootCmd := &cobra.Command{
		Use:   "diffff [flags] [commit | from to | staged | working]",
		Short: "Local code review tool with AI assistance",
		Long:  `diffff is a platform-independent code review tool that visualizes git diffs in a web browser with Claude Code integration.`,
		Args:  cobra.MaximumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd, args, &cfg)
		},
	}

	rootCmd.Flags().IntVarP(&cfg.Port, "port", "p", 3333, "Port to listen on")
	rootCmd.Flags().StringVar(&cfg.Host, "host", "127.0.0.1", "Host to bind to")
	rootCmd.Flags().StringVarP(&cfg.Mode, "mode", "m", "split", "Display mode: split or unified")
	rootCmd.Flags().BoolVar(&cfg.NoOpen, "no-open", false, "Don't open browser automatically")
	rootCmd.Flags().BoolVar(&cfg.NoClaude, "no-claude", false, "Disable Claude Code integration")
	rootCmd.Flags().BoolVarP(&cfg.Watch, "watch", "w", false, "Watch for file changes (experimental, log only)")

	return rootCmd
}

func run(cmd *cobra.Command, args []string, cfg *Config) error {
	var stdinReader *os.File
	stat, err := os.Stdin.Stat()
	if err == nil && (stat.Mode()&os.ModeCharDevice) == 0 {
		stdinReader = os.Stdin
	}

	diffReq, err := ParseDiffRequest(args, stdinReader)
	if err != nil {
		return err
	}
	cfg.DiffReq = diffReq

	if cfg.Port < 1 || cfg.Port > 65535 {
		return fmt.Errorf("invalid port %d: must be between 1 and 65535", cfg.Port)
	}

	if cfg.Mode != "split" && cfg.Mode != "unified" {
		return fmt.Errorf("invalid mode %q: must be 'split' or 'unified'", cfg.Mode)
	}

	// Get current working directory as repo path
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("getting working directory: %w", err)
	}

	// Get diff content
	gitClient := git.NewClient(cwd)
	rawDiff, err := gitClient.GetDiff(context.Background(), diffReq)
	if err != nil {
		return fmt.Errorf("getting diff: %w", err)
	}

	// Start server with CLI flags
	srv, err := server.New(rawDiff,
		server.WithWorkDir(cwd),
		server.WithNoClaude(cfg.NoClaude),
		server.WithViewMode(cfg.Mode),
	)
	if err != nil {
		return fmt.Errorf("creating server: %w", err)
	}
	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	fmt.Printf("Starting diffff on http://%s (mode: %s)\n", addr, cfg.Mode)

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("listen: %w", err)
	}

	httpServer := &http.Server{
		Handler:      srv.Handler(),
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Open browser unless --no-open
	if !cfg.NoOpen {
		openBrowser(fmt.Sprintf("http://%s", addr))
	}

	// Start file watcher if --watch is enabled and not stdin mode
	if cfg.Watch && stdinReader == nil {
		w, err := watcher.New(cwd)
		if err != nil {
			slog.Warn("File watcher not available", "err", err)
		} else {
			go func() {
				for ev := range w.Events() {
					slog.Info("File changed", "path", ev.Path, "op", ev.Op)
				}
			}()
			defer w.Close()
		}
	}

	// Graceful shutdown on SIGINT/SIGTERM
	shutdownCh := make(chan os.Signal, 1)
	signal.Notify(shutdownCh, syscall.SIGINT, syscall.SIGTERM)

	errCh := make(chan error, 1)
	go func() {
		errCh <- httpServer.Serve(listener)
	}()

	select {
	case err := <-errCh:
		return err
	case sig := <-shutdownCh:
		fmt.Printf("\nReceived %s, shutting down...\n", sig)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return httpServer.Shutdown(ctx)
	}
}

func openBrowser(url string) {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "linux":
		cmd = "xdg-open"
		args = []string{url}
	case "darwin":
		cmd = "open"
		args = []string{url}
	case "windows":
		cmd = "rundll32"
		args = []string{"url.dll,FileProtocolHandler", url}
	default:
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	c := exec.CommandContext(ctx, cmd, args...)
	if err := c.Start(); err != nil {
		cancel()
		slog.Warn("Failed to open browser", "err", err)
		return
	}
	go func() {
		defer cancel()
		if err := c.Wait(); err != nil {
			slog.Warn("browser process error", "err", err)
		}
	}()
}
