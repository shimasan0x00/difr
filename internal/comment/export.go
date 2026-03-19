package comment

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

func formatPrefix(category, severity string) string {
	if category == "" && severity == "" {
		return ""
	}
	if severity == "" {
		return "[" + category + "]"
	}
	if category == "" {
		return "[" + severity + "]"
	}
	return "[" + category + "/" + severity + "]"
}

func ExportMarkdown(comments []*Comment) string {
	if len(comments) == 0 {
		return "# Code Review Comments\n\nNo comments.\n"
	}

	// Group by file
	grouped := make(map[string][]*Comment)
	for _, c := range comments {
		grouped[c.FilePath] = append(grouped[c.FilePath], c)
	}

	// Sort file names
	files := make([]string, 0, len(grouped))
	for f := range grouped {
		files = append(files, f)
	}
	sort.Strings(files)

	var sb strings.Builder
	sb.WriteString("# Code Review Comments\n\n")

	for _, file := range files {
		sb.WriteString(fmt.Sprintf("## %s\n\n", file))

		// Sort comments by line number
		fileComments := grouped[file]
		sort.Slice(fileComments, func(i, j int) bool {
			return fileComments[i].Line < fileComments[j].Line
		})

		for _, c := range fileComments {
			prefix := formatPrefix(c.ReviewCategory, c.Severity)
			if c.Line == 0 {
				if prefix != "" {
					sb.WriteString(fmt.Sprintf("- **File**: %s\n%s\n", prefix, c.Body))
				} else {
					sb.WriteString(fmt.Sprintf("- **File**: %s\n", c.Body))
				}
			} else {
				if prefix != "" {
					sb.WriteString(fmt.Sprintf("- **Line %d**: %s\n%s\n", c.Line, prefix, c.Body))
				} else {
					sb.WriteString(fmt.Sprintf("- **Line %d**: %s\n", c.Line, c.Body))
				}
			}
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

func ExportCSV(comments []*Comment) string {
	if comments == nil {
		comments = []*Comment{}
	}

	// Sort by file path then line number
	sorted := make([]*Comment, len(comments))
	copy(sorted, comments)
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].FilePath != sorted[j].FilePath {
			return sorted[i].FilePath < sorted[j].FilePath
		}
		return sorted[i].Line < sorted[j].Line
	})

	var sb strings.Builder
	w := csv.NewWriter(&sb)
	w.Write([]string{"filepath", "review_category", "severity", "comment"}) //nolint:errcheck
	for _, c := range sorted {
		body := strings.ReplaceAll(c.Body, "\n", `\n`)
		w.Write([]string{c.FilePath, c.ReviewCategory, c.Severity, body}) //nolint:errcheck
	}
	w.Flush()
	return sb.String()
}

func ExportJSON(comments []*Comment) (string, error) {
	if comments == nil {
		comments = []*Comment{}
	}
	data, err := json.MarshalIndent(comments, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}
