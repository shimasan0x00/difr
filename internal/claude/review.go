package claude

import (
	"encoding/json"
	"strings"
)

// ReviewComment represents a code review comment from Claude.
type ReviewComment struct {
	FilePath string `json:"filePath"`
	Line     int    `json:"line"`
	Body     string `json:"body"`
}

// ParseReviewComments extracts review comments from Claude's response text.
// The response may contain JSON embedded within markdown code blocks or surrounded by text.
// Returns an empty slice if no valid comments are found.
func ParseReviewComments(text string) []ReviewComment {
	// Try direct JSON parse first
	var comments []ReviewComment
	if err := json.Unmarshal([]byte(text), &comments); err == nil {
		return comments
	}

	// Try to find JSON array in the text using bracket counting
	candidate := extractJSONArray(text)
	if candidate == "" {
		return []ReviewComment{}
	}

	if err := json.Unmarshal([]byte(candidate), &comments); err != nil {
		return []ReviewComment{}
	}

	return comments
}

// extractJSONArray finds the first valid JSON array containing "filePath" in the text
// using bracket counting instead of regex to handle nested brackets correctly.
func extractJSONArray(text string) string {
	for i := 0; i < len(text); i++ {
		if text[i] != '[' {
			continue
		}

		depth := 0
		inString := false
		escaped := false

		for j := i; j < len(text); j++ {
			ch := text[j]

			if escaped {
				escaped = false
				continue
			}

			if ch == '\\' && inString {
				escaped = true
				continue
			}

			if ch == '"' {
				inString = !inString
				continue
			}

			if inString {
				continue
			}

			switch ch {
			case '[':
				depth++
			case ']':
				depth--
				if depth == 0 {
					candidate := text[i : j+1]
					if strings.Contains(candidate, `"filePath"`) {
						return candidate
					}
					// Not the right array, continue searching
					break
				}
			}

			if depth == 0 && ch == ']' {
				break
			}
		}
	}
	return ""
}
