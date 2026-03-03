package claude

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os/exec"
	"sync"
)

const maxStderrSize = 64 * 1024 // 64KB

// limitedBuffer is a bytes.Buffer that silently discards writes beyond a max size.
type limitedBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
	max int
}

func (lb *limitedBuffer) Write(p []byte) (int, error) {
	lb.mu.Lock()
	defer lb.mu.Unlock()
	remaining := lb.max - lb.buf.Len()
	if remaining <= 0 {
		return len(p), nil // discard silently
	}
	if len(p) > remaining {
		p = p[:remaining]
	}
	lb.buf.Write(p)
	return len(p), nil
}

func (lb *limitedBuffer) Len() int {
	lb.mu.Lock()
	defer lb.mu.Unlock()
	return lb.buf.Len()
}

func (lb *limitedBuffer) String() string {
	lb.mu.Lock()
	defer lb.mu.Unlock()
	return lb.buf.String()
}

// PathLooker abstracts exec.LookPath for testing.
type PathLooker interface {
	LookPath(file string) (string, error)
}

type defaultLookPath struct{}

func (d *defaultLookPath) LookPath(file string) (string, error) {
	return exec.LookPath(file)
}

// Client manages interaction with the Claude Code CLI.
type Client struct {
	binPath string
}

// Option configures a Client.
type Option func(*clientConfig)

type clientConfig struct {
	lookPath PathLooker
}

// WithLookPath sets a custom PathLooker (for testing).
func WithLookPath(lp PathLooker) Option {
	return func(c *clientConfig) {
		c.lookPath = lp
	}
}

// NewClient creates a new Claude Code client.
// Returns an error if the claude CLI is not found.
func NewClient(opts ...Option) (*Client, error) {
	cfg := &clientConfig{
		lookPath: &defaultLookPath{},
	}
	for _, opt := range opts {
		opt(cfg)
	}

	path, err := cfg.lookPath.LookPath("claude")
	if err != nil {
		return nil, err
	}

	return &Client{binPath: path}, nil
}

// Available returns true if the Claude CLI was found.
func (c *Client) Available() bool {
	return c.binPath != ""
}

// cmdReadCloser wraps stdout and ensures cmd.Wait() is called on Close
// to prevent zombie processes. Captures stderr for error diagnostics.
type cmdReadCloser struct {
	io.ReadCloser
	cmd    *exec.Cmd
	stderr *limitedBuffer
}

func (c *cmdReadCloser) Close() error {
	_ = c.ReadCloser.Close()
	if err := c.cmd.Wait(); err != nil {
		if c.stderr.Len() > 0 {
			return fmt.Errorf("%w: %s", err, c.stderr.String())
		}
		return err
	}
	return nil
}

// Run executes the Claude CLI with the given arguments and returns a reader for its stdout.
// The returned ReadCloser must be closed to avoid zombie processes.
func (c *Client) Run(ctx context.Context, args []string) (io.ReadCloser, error) {
	cmd := exec.CommandContext(ctx, c.binPath, args...)
	stderr := &limitedBuffer{max: maxStderrSize}
	cmd.Stderr = stderr
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	return &cmdReadCloser{ReadCloser: stdout, cmd: cmd, stderr: stderr}, nil
}
