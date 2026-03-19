package comment

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/xuri/excelize/v2"
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

// sortedComments returns a copy of comments sorted by file path then line number.
func sortedComments(comments []*Comment) []*Comment {
	sorted := make([]*Comment, len(comments))
	copy(sorted, comments)
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].FilePath != sorted[j].FilePath {
			return sorted[i].FilePath < sorted[j].FilePath
		}
		return sorted[i].Line < sorted[j].Line
	})
	return sorted
}

func ExportCSV(comments []*Comment) string {
	if comments == nil {
		comments = []*Comment{}
	}

	sorted := sortedComments(comments)

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

func ExportExcel(comments []*Comment) ([]byte, error) {
	if comments == nil {
		comments = []*Comment{}
	}

	sorted := sortedComments(comments)

	f := excelize.NewFile()
	defer f.Close()

	sheet := "Comments"
	f.SetSheetName("Sheet1", sheet)

	headers := []string{"filepath", "review_category", "severity", "comment"}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, h)
	}

	// Header style: light green background + bold black text
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Color: "000000"},
		Fill: excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"C6EFCE"}},
	})
	f.SetCellStyle(sheet, "A1", "D1", headerStyle)

	// WrapText style for comment column
	wrapStyle, _ := f.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{WrapText: true},
	})

	for i, c := range sorted {
		row := i + 2
		aCell, _ := excelize.CoordinatesToCellName(1, row)
		bCell, _ := excelize.CoordinatesToCellName(2, row)
		cCell, _ := excelize.CoordinatesToCellName(3, row)
		dCell, _ := excelize.CoordinatesToCellName(4, row)

		f.SetCellValue(sheet, aCell, c.FilePath)
		f.SetCellValue(sheet, bCell, c.ReviewCategory)
		f.SetCellValue(sheet, cCell, c.Severity)
		f.SetCellValue(sheet, dCell, c.Body)
		f.SetCellStyle(sheet, dCell, dCell, wrapStyle)
	}

	// Column widths
	f.SetColWidth(sheet, "A", "A", 40)
	f.SetColWidth(sheet, "B", "B", 18)
	f.SetColWidth(sheet, "C", "C", 12)
	f.SetColWidth(sheet, "D", "D", 60)

	// Add table with filter arrows and banded rows
	lastRow := len(sorted) + 1
	if lastRow < 2 {
		lastRow = 2 // at least one data row for valid table range
	}
	endCell, _ := excelize.CoordinatesToCellName(4, lastRow)
	showStripes := true
	f.AddTable(sheet, &excelize.Table{
		Range:          "A1:" + endCell,
		Name:           "Comments",
		StyleName:      "TableStyleMedium9",
		ShowRowStripes: &showStripes,
	})

	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
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
