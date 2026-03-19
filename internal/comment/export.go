package comment

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

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

// ExcelFilename returns the xlsx filename with today's date: review_YYYYMMDD.xlsx
func ExcelFilename() string {
	return "review_" + time.Now().Format("20060102") + ".xlsx"
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

	headers := []string{"filepath", "review_category", "severity", "comment", "resolved", "reviewer_confirmed", "notes"}
	colCount := len(headers)
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, h)
	}

	thinBorder := []excelize.Border{
		{Type: "left", Color: "000000", Style: 1},
		{Type: "right", Color: "000000", Style: 1},
		{Type: "top", Color: "000000", Style: 1},
		{Type: "bottom", Color: "000000", Style: 1},
	}

	// Header style: light green background + bold black text + border
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font:   &excelize.Font{Bold: true, Color: "000000"},
		Fill:   excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"C6EFCE"}},
		Border: thinBorder,
	})
	lastHeaderCell, _ := excelize.CoordinatesToCellName(colCount, 1)
	f.SetCellStyle(sheet, "A1", lastHeaderCell, headerStyle)

	// Data cell style: border only (no background)
	dataStyle, _ := f.NewStyle(&excelize.Style{
		Border: thinBorder,
	})

	// Data cell style with WrapText for comment/notes columns
	dataWrapStyle, _ := f.NewStyle(&excelize.Style{
		Border:    thinBorder,
		Alignment: &excelize.Alignment{WrapText: true},
	})

	for i, c := range sorted {
		row := i + 2
		for col := 1; col <= colCount; col++ {
			cell, _ := excelize.CoordinatesToCellName(col, row)
			switch col {
			case 1:
				f.SetCellValue(sheet, cell, c.FilePath)
			case 2:
				f.SetCellValue(sheet, cell, c.ReviewCategory)
			case 3:
				f.SetCellValue(sheet, cell, c.Severity)
			case 4:
				f.SetCellValue(sheet, cell, c.Body)
			}
			// columns 5-7 (resolved, reviewer_confirmed, notes) left empty
		}

		// Apply styles
		firstCell, _ := excelize.CoordinatesToCellName(1, row)
		cCell, _ := excelize.CoordinatesToCellName(3, row)
		dCell, _ := excelize.CoordinatesToCellName(4, row)
		eCell, _ := excelize.CoordinatesToCellName(5, row)
		fCell, _ := excelize.CoordinatesToCellName(6, row)
		gCell, _ := excelize.CoordinatesToCellName(7, row)

		f.SetCellStyle(sheet, firstCell, cCell, dataStyle)
		f.SetCellStyle(sheet, dCell, dCell, dataWrapStyle)
		f.SetCellStyle(sheet, eCell, fCell, dataStyle)
		f.SetCellStyle(sheet, gCell, gCell, dataWrapStyle)
	}

	// Column widths
	f.SetColWidth(sheet, "A", "A", 40)
	f.SetColWidth(sheet, "B", "B", 18)
	f.SetColWidth(sheet, "C", "C", 12)
	f.SetColWidth(sheet, "D", "D", 60)
	f.SetColWidth(sheet, "E", "E", 12)
	f.SetColWidth(sheet, "F", "F", 20)
	f.SetColWidth(sheet, "G", "G", 30)

	// AutoFilter for header row
	f.AutoFilter(sheet, "A1:"+lastHeaderCell, nil)

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
