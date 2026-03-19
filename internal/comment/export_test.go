package comment

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xuri/excelize/v2"
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

// --- Excel export tests ---

func openExcel(t *testing.T, data []byte) *excelize.File {
	t.Helper()
	f, err := excelize.OpenReader(bytes.NewReader(data))
	require.NoError(t, err)
	t.Cleanup(func() { f.Close() })
	return f
}

func TestExportExcel_BasicOutput(t *testing.T) {
	comments := []*Comment{
		{ID: "c1", FilePath: "main.go", Line: 10, Body: "Fix this", ReviewCategory: "MUST", Severity: "Critical"},
		{ID: "c2", FilePath: "a.go", Line: 5, Body: "Check this"},
	}

	data, err := ExportExcel(comments)
	require.NoError(t, err)

	f := openExcel(t, data)

	// Sheet name
	assert.Equal(t, "Comments", f.GetSheetName(f.GetActiveSheetIndex()))

	// Header row (7 columns)
	h1, _ := f.GetCellValue("Comments", "A1")
	h2, _ := f.GetCellValue("Comments", "B1")
	h3, _ := f.GetCellValue("Comments", "C1")
	h4, _ := f.GetCellValue("Comments", "D1")
	h5, _ := f.GetCellValue("Comments", "E1")
	h6, _ := f.GetCellValue("Comments", "F1")
	h7, _ := f.GetCellValue("Comments", "G1")
	assert.Equal(t, "filepath", h1)
	assert.Equal(t, "review_category", h2)
	assert.Equal(t, "severity", h3)
	assert.Equal(t, "comment", h4)
	assert.Equal(t, "resolved", h5)
	assert.Equal(t, "reviewer_confirmed", h6)
	assert.Equal(t, "notes", h7)

	// Data rows (sorted: a.go first, then main.go)
	a2, _ := f.GetCellValue("Comments", "A2")
	d2, _ := f.GetCellValue("Comments", "D2")
	a3, _ := f.GetCellValue("Comments", "A3")
	b3, _ := f.GetCellValue("Comments", "B3")
	c3, _ := f.GetCellValue("Comments", "C3")
	d3, _ := f.GetCellValue("Comments", "D3")
	assert.Equal(t, "a.go", a2)
	assert.Equal(t, "Check this", d2)
	assert.Equal(t, "main.go", a3)
	assert.Equal(t, "MUST", b3)
	assert.Equal(t, "Critical", c3)
	assert.Equal(t, "Fix this", d3)

	// Extra columns are empty (for manual fill)
	e2, _ := f.GetCellValue("Comments", "E2")
	f2, _ := f.GetCellValue("Comments", "F2")
	g2, _ := f.GetCellValue("Comments", "G2")
	assert.Empty(t, e2)
	assert.Empty(t, f2)
	assert.Empty(t, g2)
}

func TestExportExcel_PreservesCellInternalNewlines(t *testing.T) {
	comments := []*Comment{
		{ID: "c1", FilePath: "main.go", Line: 1, Body: "Line 1\nLine 2\nLine 3"},
	}

	data, err := ExportExcel(comments)
	require.NoError(t, err)

	f := openExcel(t, data)
	val, _ := f.GetCellValue("Comments", "D2")
	assert.Equal(t, "Line 1\nLine 2\nLine 3", val)
}

func TestExportExcel_HeaderHasLightGreenBackground(t *testing.T) {
	comments := []*Comment{}

	data, err := ExportExcel(comments)
	require.NoError(t, err)

	f := openExcel(t, data)

	// Check first and last header cells share the same style
	for _, cell := range []string{"A1", "G1"} {
		styleID, _ := f.GetCellStyle("Comments", cell)
		style, err := f.GetStyle(styleID)
		require.NoError(t, err)

		assert.True(t, style.Font.Bold, "header font should be bold (%s)", cell)
		require.NotNil(t, style.Fill)
		require.NotEmpty(t, style.Fill.Color)
		assert.Equal(t, "C6EFCE", style.Fill.Color[0])
		require.NotEmpty(t, style.Border, "header cells should have borders (%s)", cell)
	}
}

func TestExportExcel_DataRowsHaveNoBgColorAndBorders(t *testing.T) {
	comments := []*Comment{
		{ID: "c1", FilePath: "main.go", Line: 10, Body: "Fix this"},
	}

	data, err := ExportExcel(comments)
	require.NoError(t, err)

	f := openExcel(t, data)

	// Data cell A2 should have no background fill
	styleID, _ := f.GetCellStyle("Comments", "A2")
	style, err := f.GetStyle(styleID)
	require.NoError(t, err)
	assert.Empty(t, style.Fill.Color, "data rows should have no background color")

	// Data cells should have borders
	require.NotEmpty(t, style.Border, "data cells should have borders")
}

func TestExportExcel_EmptyComments(t *testing.T) {
	data, err := ExportExcel([]*Comment{})
	require.NoError(t, err)

	f := openExcel(t, data)
	h1, _ := f.GetCellValue("Comments", "A1")
	assert.Equal(t, "filepath", h1)

	// No data row
	a2, _ := f.GetCellValue("Comments", "A2")
	assert.Empty(t, a2)
}

func TestExportExcel_NilComments(t *testing.T) {
	data, err := ExportExcel(nil)
	require.NoError(t, err)

	f := openExcel(t, data)
	h1, _ := f.GetCellValue("Comments", "A1")
	assert.Equal(t, "filepath", h1)
}

func TestExportExcel_SortedByFilePathThenLine(t *testing.T) {
	comments := []*Comment{
		{ID: "c1", FilePath: "b.go", Line: 20, Body: "b20"},
		{ID: "c2", FilePath: "a.go", Line: 10, Body: "a10"},
		{ID: "c3", FilePath: "b.go", Line: 5, Body: "b5"},
		{ID: "c4", FilePath: "a.go", Line: 1, Body: "a1"},
	}

	data, err := ExportExcel(comments)
	require.NoError(t, err)

	f := openExcel(t, data)

	d2, _ := f.GetCellValue("Comments", "D2")
	d3, _ := f.GetCellValue("Comments", "D3")
	d4, _ := f.GetCellValue("Comments", "D4")
	d5, _ := f.GetCellValue("Comments", "D5")
	assert.Equal(t, "a1", d2)
	assert.Equal(t, "a10", d3)
	assert.Equal(t, "b5", d4)
	assert.Equal(t, "b20", d5)
}

func TestExcelFilename_ContainsTodaysDate(t *testing.T) {
	filename := ExcelFilename()
	today := time.Now().Format("20060102")

	assert.Equal(t, "review_"+today+".xlsx", filename)
}

func TestExportExcel_SpecialCharactersPreserved(t *testing.T) {
	comments := []*Comment{
		{ID: "c1", FilePath: "日本語.go", Line: 1, Body: "Unicode: 🎉 \"quoted\" コメント"},
	}

	data, err := ExportExcel(comments)
	require.NoError(t, err)

	f := openExcel(t, data)

	a2, _ := f.GetCellValue("Comments", "A2")
	d2, _ := f.GetCellValue("Comments", "D2")
	assert.Equal(t, "日本語.go", a2)
	assert.Equal(t, "Unicode: 🎉 \"quoted\" コメント", d2)
}
