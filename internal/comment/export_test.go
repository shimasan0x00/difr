package comment

import (
	"encoding/json"
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

func TestExportJSON_EmptyCommentsReturnsEmptyArray(t *testing.T) {
	jsonStr, err := ExportJSON([]*Comment{})
	require.NoError(t, err)

	var parsed []Comment
	require.NoError(t, json.Unmarshal([]byte(jsonStr), &parsed))
	assert.Empty(t, parsed)
}
