package diff

import (
	"path/filepath"
	"strings"

	"github.com/bluekeyes/go-gitdiff/gitdiff"
)

// langMap maps file extensions to language identifiers for syntax highlighting.
var langMap = map[string]string{
	"go":    "go",
	"ts":    "typescript",
	"tsx":   "tsx",
	"js":    "javascript",
	"jsx":   "jsx",
	"py":    "python",
	"rs":    "rust",
	"java":  "java",
	"rb":    "ruby",
	"php":   "php",
	"c":     "c",
	"cpp":   "cpp",
	"h":     "c",
	"hpp":   "cpp",
	"cs":    "csharp",
	"swift": "swift",
	"kt":    "kotlin",
	"sql":   "sql",
	"sh":    "bash",
	"bash":  "bash",
	"zsh":   "bash",
	"yaml":  "yaml",
	"yml":   "yaml",
	"json":  "json",
	"xml":   "xml",
	"html":  "html",
	"css":   "css",
	"scss":  "scss",
	"md":    "markdown",
	"toml":  "toml",
	"proto": "protobuf",
	"dart":  "dart",
}

// Parse parses a raw unified diff string into a DiffResult.
func Parse(raw string) (*DiffResult, error) {
	if strings.TrimSpace(raw) == "" {
		return &DiffResult{Files: []DiffFile{}}, nil
	}

	files, _, err := gitdiff.Parse(strings.NewReader(raw))
	if err != nil {
		return nil, err
	}

	result := &DiffResult{Files: []DiffFile{}}
	for _, f := range files {
		df := convertFile(f)
		result.Files = append(result.Files, df)
		result.Stats.Additions += df.Stats.Additions
		result.Stats.Deletions += df.Stats.Deletions
	}

	return result, nil
}

func convertFile(f *gitdiff.File) DiffFile {
	df := DiffFile{
		OldPath:  f.OldName,
		NewPath:  f.NewName,
		Status:   detectFileStatus(f),
		IsBinary: f.IsBinary,
		Hunks:    []Hunk{},
	}

	// Determine display path for language detection
	displayPath := df.NewPath
	if displayPath == "" || displayPath == "/dev/null" {
		displayPath = df.OldPath
	}
	df.Language = detectLanguage(displayPath)

	// Convert text fragments to hunks
	for _, frag := range f.TextFragments {
		hunk := convertHunk(frag)
		df.Hunks = append(df.Hunks, hunk)
		for _, line := range hunk.Lines {
			switch line.Type {
			case LineAdd:
				df.Stats.Additions++
			case LineDelete:
				df.Stats.Deletions++
			}
		}
	}

	return df
}

func convertHunk(frag *gitdiff.TextFragment) Hunk {
	h := Hunk{
		OldStart: int(frag.OldPosition),
		OldLines: int(frag.OldLines),
		NewStart: int(frag.NewPosition),
		NewLines: int(frag.NewLines),
		Header:   strings.TrimSpace(frag.Comment),
		Lines:    []Line{},
	}

	oldNum := int(frag.OldPosition)
	newNum := int(frag.NewPosition)

	for _, line := range frag.Lines {
		dl := Line{
			Content: line.Line,
		}
		switch line.Op {
		case gitdiff.OpContext:
			dl.Type = LineContext
			dl.OldNumber = oldNum
			dl.NewNumber = newNum
			oldNum++
			newNum++
		case gitdiff.OpAdd:
			dl.Type = LineAdd
			dl.NewNumber = newNum
			newNum++
		case gitdiff.OpDelete:
			dl.Type = LineDelete
			dl.OldNumber = oldNum
			oldNum++
		}
		h.Lines = append(h.Lines, dl)
	}

	return h
}

func detectFileStatus(f *gitdiff.File) FileStatus {
	if f.IsNew {
		return FileStatusAdded
	}
	if f.IsDelete {
		return FileStatusDeleted
	}
	if f.IsRename {
		return FileStatusRenamed
	}
	return FileStatusModified
}

func detectLanguage(path string) string {
	ext := strings.TrimPrefix(filepath.Ext(path), ".")
	if lang, ok := langMap[ext]; ok {
		return lang
	}
	return ext
}
