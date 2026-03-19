package comment

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExportMarkdown_GroupsByFileAndIncludesLineReferences(t *testing.T) {
	comments := []*Comment{
		{ID: "c1", FilePath: "main.go", Line: 10, Body: "This needs refactoring", CreatedAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)},
		{ID: "c2", FilePath: "main.go", Line: 20, Body: "Add error handling", CreatedAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)},
		{ID: "c3", FilePath: "utils.go", Line: 5, Body: "Good approach", CreatedAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)},
	}

	md := ExportMarkdown(comments)

	assert.Contains(t, md, "## main.go")
	assert.Contains(t, md, "## utils.go")
	assert.Contains(t, md, "Line 10")
	assert.Contains(t, md, "This needs refactoring")
}

func TestExportMarkdown_EmptyCommentsReturnsPlaceholder(t *testing.T) {
	md := ExportMarkdown([]*Comment{})

	assert.Equal(t, "# Code Review Comments\n\nNo comments.\n", md)
}

func TestExportJSON_ProducesValidRoundtrippableJSON(t *testing.T) {
	comments := []*Comment{
		{ID: "c1", FilePath: "main.go", Line: 10, Body: "Fix this"},
	}

	jsonStr, err := ExportJSON(comments)
	require.NoError(t, err)

	var parsed []Comment
	require.NoError(t, json.Unmarshal([]byte(jsonStr), &parsed))
	require.Len(t, parsed, 1)
	assert.Equal(t, "Fix this", parsed[0].Body)
}

func TestExportMarkdown_HandlesSpecialCharacters(t *testing.T) {
	comments := []*Comment{
		{ID: "c1", FilePath: "main.go", Line: 1, Body: "Line 1\nLine 2\nLine 3", CreatedAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)},
		{ID: "c2", FilePath: "main.go", Line: 2, Body: "Contains **bold** and `code` and [link](url)", CreatedAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)},
		{ID: "c3", FilePath: "日本語ファイル.go", Line: 3, Body: "Unicode: 🎉 コメント", CreatedAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)},
	}

	md := ExportMarkdown(comments)

	assert.Contains(t, md, "Line 1\nLine 2\nLine 3")
	assert.Contains(t, md, "**bold**")
	assert.Contains(t, md, "`code`")
	assert.Contains(t, md, "## 日本語ファイル.go")
	assert.Contains(t, md, "🎉 コメント")
}

func TestExportJSON_RoundtripsSpecialCharacters(t *testing.T) {
	comments := []*Comment{
		{ID: "c1", FilePath: "main.go", Line: 1, Body: "Line 1\nLine 2"},
		{ID: "c2", FilePath: "日本語.go", Line: 2, Body: "Unicode: 🎉 \"quoted\""},
	}

	jsonStr, err := ExportJSON(comments)
	require.NoError(t, err)

	var parsed []Comment
	require.NoError(t, json.Unmarshal([]byte(jsonStr), &parsed))
	require.Len(t, parsed, 2)
	assert.Equal(t, "Line 1\nLine 2", parsed[0].Body)
	assert.Equal(t, "Unicode: 🎉 \"quoted\"", parsed[1].Body)
	assert.Equal(t, "日本語.go", parsed[1].FilePath)
}

func TestExportMarkdown_FileCommentUsesFileLabel(t *testing.T) {
	comments := []*Comment{
		{ID: "c1", FilePath: "main.go", Line: 0, Body: "General feedback", CreatedAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)},
		{ID: "c2", FilePath: "main.go", Line: 10, Body: "Fix this", CreatedAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)},
	}

	md := ExportMarkdown(comments)

	assert.Contains(t, md, "- **File**: General feedback")
	assert.Contains(t, md, "- **Line 10**: Fix this")
	assert.NotContains(t, md, "Line 0")
}

func TestExportMarkdown_WithCategoryAndSeverityPrefix(t *testing.T) {
	comments := []*Comment{
		{ID: "c1", FilePath: "main.go", Line: 10, Body: "Fix this", ReviewCategory: "MUST", Severity: "Critical", CreatedAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)},
		{ID: "c2", FilePath: "main.go", Line: 20, Body: "Consider this", ReviewCategory: "IMO", CreatedAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)},
		{ID: "c3", FilePath: "main.go", Line: 30, Body: "No prefix", CreatedAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)},
		{ID: "c4", FilePath: "main.go", Line: 0, Body: "File comment", ReviewCategory: "FYI", Severity: "Low", CreatedAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)},
	}

	md := ExportMarkdown(comments)

	assert.Contains(t, md, "- **Line 10**: [MUST/Critical]\nFix this")
	assert.Contains(t, md, "- **Line 20**: [IMO]\nConsider this")
	assert.Contains(t, md, "- **Line 30**: No prefix")
	assert.Contains(t, md, "- **File**: [FYI/Low]\nFile comment")
}

func TestExportCSV_BasicOutput(t *testing.T) {
	comments := []*Comment{
		{ID: "c1", FilePath: "main.go", Line: 10, Body: "Fix this", ReviewCategory: "MUST", Severity: "Critical"},
		{ID: "c2", FilePath: "a.go", Line: 5, Body: "Check this"},
	}

	csvStr := ExportCSV(comments)

	assert.Contains(t, csvStr, "filepath,review_category,severity,comment\n")
	assert.Contains(t, csvStr, "a.go,,,Check this\n")
	assert.Contains(t, csvStr, "main.go,MUST,Critical,Fix this\n")
}

func TestExportCSV_EscapesSpecialCharacters(t *testing.T) {
	comments := []*Comment{
		{ID: "c1", FilePath: "main.go", Line: 1, Body: "Has \"quotes\" and, commas"},
		{ID: "c2", FilePath: "test.go", Line: 2, Body: "Has\nnewlines"},
	}

	csvStr := ExportCSV(comments)

	assert.Contains(t, csvStr, `"Has ""quotes"" and, commas"`)
	// Newlines in body are escaped to literal \n
	assert.Contains(t, csvStr, `Has\nnewlines`)
	assert.NotContains(t, csvStr, "Has\nnewlines")
}

func TestExportCSV_EmptyComments(t *testing.T) {
	csvStr := ExportCSV([]*Comment{})

	assert.Equal(t, "filepath,review_category,severity,comment\n", csvStr)
}

func TestExportCSV_NilComments(t *testing.T) {
	csvStr := ExportCSV(nil)

	assert.Equal(t, "filepath,review_category,severity,comment\n", csvStr)
}

func TestExportCSV_SortedByFilePathThenLine(t *testing.T) {
	comments := []*Comment{
		{ID: "c1", FilePath: "b.go", Line: 20, Body: "b20"},
		{ID: "c2", FilePath: "a.go", Line: 10, Body: "a10"},
		{ID: "c3", FilePath: "b.go", Line: 5, Body: "b5"},
		{ID: "c4", FilePath: "a.go", Line: 1, Body: "a1"},
	}

	csvStr := ExportCSV(comments)

	lines := strings.Split(strings.TrimSpace(csvStr), "\n")
	require.Len(t, lines, 5) // header + 4 data rows
	assert.Contains(t, lines[1], "a.go")
	assert.Contains(t, lines[1], "a1")
	assert.Contains(t, lines[2], "a.go")
	assert.Contains(t, lines[2], "a10")
	assert.Contains(t, lines[3], "b.go")
	assert.Contains(t, lines[3], "b5")
	assert.Contains(t, lines[4], "b.go")
	assert.Contains(t, lines[4], "b20")
}

func TestExportJSON_EmptyCommentsReturnsEmptyArray(t *testing.T) {
	jsonStr, err := ExportJSON([]*Comment{})
	require.NoError(t, err)

	var parsed []Comment
	require.NoError(t, json.Unmarshal([]byte(jsonStr), &parsed))
	assert.Empty(t, parsed)
}
