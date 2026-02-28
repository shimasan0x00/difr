package cli

import (
	"fmt"
	"io"

	"github.com/shimasan0x00/diffff/internal/diff"
)

// ParseDiffRequest parses CLI arguments and stdin into a DiffRequest.
// If stdinReader is non-nil, stdin mode takes precedence over arguments.
func ParseDiffRequest(args []string, stdinReader io.Reader) (diff.DiffRequest, error) {
	if len(args) > 2 {
		return diff.DiffRequest{}, fmt.Errorf("too many arguments: expected at most 2, got %d", len(args))
	}

	if stdinReader != nil {
		return diff.DiffRequest{Mode: diff.DiffModeStdin, Stdin: stdinReader}, nil
	}

	if len(args) == 0 {
		return diff.DiffRequest{Mode: diff.DiffModeLatestCommit}, nil
	}

	if len(args) == 1 {
		switch args[0] {
		case "staged":
			return diff.DiffRequest{Mode: diff.DiffModeStaged}, nil
		case "working":
			return diff.DiffRequest{Mode: diff.DiffModeWorking}, nil
		default:
			return diff.DiffRequest{Mode: diff.DiffModeCommit, From: args[0]}, nil
		}
	}

	return diff.DiffRequest{Mode: diff.DiffModeRange, From: args[0], To: args[1]}, nil
}
