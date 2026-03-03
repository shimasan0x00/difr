package git

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os/exec"
	"strings"

	"github.com/shimasan0x00/difr/internal/diff"
)

const maxStdinSize = 100 * 1024 * 1024 // 100MB

// GetDiff executes the appropriate git diff command and returns the raw diff output.
func (c *Client) GetDiff(ctx context.Context, req diff.DiffRequest) (string, error) {
	if req.Mode == diff.DiffModeStdin {
		data, err := io.ReadAll(io.LimitReader(req.Stdin, maxStdinSize))
		if err != nil {
			return "", fmt.Errorf("reading stdin: %w", err)
		}
		return string(data), nil
	}

	args, err := c.buildDiffArgs(req)
	if err != nil {
		return "", err
	}

	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = c.repoPath

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		// git diff returns exit code 1 when there are differences, which is normal
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			return stdout.String(), nil
		}
		return "", fmt.Errorf("git diff: %w: %s", err, stderr.String())
	}

	return stdout.String(), nil
}

// validateRef rejects empty refs and refs starting with "-" to prevent command injection.
func validateRef(ref string) error {
	if ref == "" {
		return fmt.Errorf("invalid ref: must not be empty")
	}
	if strings.HasPrefix(ref, "-") {
		return fmt.Errorf("invalid ref %q: must not start with '-'", ref)
	}
	return nil
}

func (c *Client) buildDiffArgs(req diff.DiffRequest) ([]string, error) {
	switch req.Mode {
	case diff.DiffModeLatestCommit:
		return []string{"diff", "HEAD~1", "HEAD", "--"}, nil
	case diff.DiffModeCommit:
		if err := validateRef(req.From); err != nil {
			return nil, err
		}
		return []string{"diff", req.From + "~1", req.From, "--"}, nil
	case diff.DiffModeRange:
		if err := validateRef(req.From); err != nil {
			return nil, err
		}
		if err := validateRef(req.To); err != nil {
			return nil, err
		}
		return []string{"diff", req.From + "..." + req.To, "--"}, nil
	case diff.DiffModeStaged:
		return []string{"diff", "--cached", "--"}, nil
	case diff.DiffModeWorking:
		return []string{"diff", "--"}, nil
	default:
		return nil, fmt.Errorf("unsupported diff mode: %d", req.Mode)
	}
}
