package comment

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

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
			if c.Line == 0 {
				sb.WriteString(fmt.Sprintf("- **File**: %s\n", c.Body))
			} else {
				sb.WriteString(fmt.Sprintf("- **Line %d**: %s\n", c.Line, c.Body))
			}
		}
		sb.WriteString("\n")
	}

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
