package diff

import "io"

// DiffMode represents how to generate the diff.
type DiffMode int

const (
	DiffModeLatestCommit DiffMode = iota // HEAD~1..HEAD
	DiffModeCommit                       // <commit>~1..<commit>
	DiffModeRange                        // <from>..<to>
	DiffModeStaged                       // --cached
	DiffModeWorking                      // unstaged changes
	DiffModeStdin                        // from stdin pipe
)

// DiffRequest holds the parsed diff request parameters.
type DiffRequest struct {
	Mode  DiffMode
	From  string
	To    string
	Stdin io.Reader
}

// FileStatus represents the type of change for a file.
type FileStatus string

const (
	FileStatusAdded    FileStatus = "added"
	FileStatusDeleted  FileStatus = "deleted"
	FileStatusModified FileStatus = "modified"
	FileStatusRenamed  FileStatus = "renamed"
)

// LineType represents the type of a diff line.
type LineType string

const (
	LineContext LineType = "context"
	LineAdd     LineType = "add"
	LineDelete  LineType = "delete"
)

// DiffFile represents a single file's diff.
type DiffFile struct {
	OldPath  string     `json:"oldPath"`
	NewPath  string     `json:"newPath"`
	Status   FileStatus `json:"status"`
	Language string     `json:"language"`
	IsBinary bool       `json:"isBinary"`
	Hunks    []Hunk     `json:"hunks"`
	Stats    FileStats  `json:"stats"`
}

// Hunk represents a contiguous block of changes.
type Hunk struct {
	OldStart int    `json:"oldStart"`
	OldLines int    `json:"oldLines"`
	NewStart int    `json:"newStart"`
	NewLines int    `json:"newLines"`
	Header   string `json:"header"`
	Lines    []Line `json:"lines"`
}

// Line represents a single line in a diff.
type Line struct {
	Type      LineType `json:"type"`
	Content   string   `json:"content"`
	OldNumber int      `json:"oldNumber,omitempty"`
	NewNumber int      `json:"newNumber,omitempty"`
}

// FileStats holds addition/deletion counts for a file.
type FileStats struct {
	Additions int `json:"additions"`
	Deletions int `json:"deletions"`
}

// DiffMeta holds comparison metadata (what is being compared).
type DiffMeta struct {
	From string `json:"from"` // e.g. "HEAD~1", "main", "staged"
	To   string `json:"to"`   // e.g. "HEAD", "feature/xyz", "working"
	Mode string `json:"mode"` // "commit", "range", "staged", "working", "stdin"
}

// DiffResult holds the complete parsed diff result.
type DiffResult struct {
	Files []DiffFile `json:"files"`
	Stats FileStats  `json:"stats"`
	Meta  DiffMeta   `json:"meta"`
}
