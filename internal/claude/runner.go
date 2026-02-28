package claude

import (
	"context"
	"io"
)

// Runner abstracts Claude CLI subprocess execution for testing.
type Runner interface {
	Run(ctx context.Context, args []string) (io.ReadCloser, error)
}
