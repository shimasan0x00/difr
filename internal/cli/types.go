package cli

import (
	"time"

	"github.com/shimasan0x00/difr/internal/diff"
)

// Config holds the CLI configuration.
type Config struct {
	Port           int
	Host           string
	Mode           string // "split" or "unified"
	NoOpen         bool
	NoClaude       bool
	Watch          bool
	ClaudeTimeout  time.Duration
	DiffReq        diff.DiffRequest
}
