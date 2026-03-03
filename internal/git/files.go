package git

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// ListTrackedFiles returns all files tracked by git in the repository.
func (c *Client) ListTrackedFiles(ctx context.Context) ([]string, error) {
	cmd := exec.CommandContext(ctx, "git", "ls-files")
	cmd.Dir = c.repoPath

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("git ls-files: %w: %s", err, stderr.String())
	}

	output := strings.TrimSpace(stdout.String())
	if output == "" {
		return []string{}, nil
	}

	return strings.Split(output, "\n"), nil
}
