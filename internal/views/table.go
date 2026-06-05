package views

import (
	"strings"

	"github.com/charmbracelet/lipgloss"

	"incidencias_tui/internal/styles"
)

type Table struct {
	Headers []string
	Rows    [][]string
	Cursor  int
	Offset  int
	PageSz  int
}

func NewTable(headers []string) *Table {
	return &Table{
		Headers: headers,
		PageSz:  15,
	}
}

// renderCell pads or truncates text to exactly the given width using plain spaces.
// This ensures the rendered width is predictable and doesn't cause line wrapping.
func renderCell(style lipgloss.Style, text string, width int) string {
	// Truncate to fit
	if lipgloss.Width(text) > width {
		text = truncate(text, width)
	}
	// Pad with spaces on the right to fill the exact width
	text = padRight(text, width)
	return style.Render(text)
}

func padRight(s string, width int) string {
	w := lipgloss.Width(s)
	if w >= width {
		return s
	}
	return s + strings.Repeat(" ", width-w)
}

func (t *Table) Render(colWidths []int) string {
	if len(t.Headers) == 0 || len(colWidths) == 0 {
		return ""
	}

	numCols := len(colWidths)
	if numCols > len(t.Headers) {
		numCols = len(t.Headers)
	}

	// Build header row
	var headerCells []string
	for i := 0; i < numCols; i++ {
		h := truncate(t.Headers[i], colWidths[i])
		headerCells = append(headerCells, renderCell(styles.TblHeader, h, colWidths[i]))
	}
	header := strings.Join(headerCells, " ")

	// Build separator
	var sepParts []string
	for _, w := range colWidths[:numCols] {
		sepParts = append(sepParts, strings.Repeat("─", w))
	}
	separator := styles.TblHeader.Render(strings.Join(sepParts, "─"))

	// Build rows
	var lines []string
	lines = append(lines, header)
	lines = append(lines, separator)

	end := t.Offset + t.PageSz
	if end > len(t.Rows) {
		end = len(t.Rows)
	}

	for i := t.Offset; i < end; i++ {
		row := t.Rows[i]
		var cells []string

		var style lipgloss.Style
		if i == t.Cursor {
			style = styles.TblSelected
		} else if i%2 == 0 {
			style = styles.TblRow
		} else {
			style = styles.TblRowAlt
		}

		for j := 0; j < numCols; j++ {
			val := ""
			if j < len(row) {
				val = row[j]
			}
			if j == 0 && i == t.Cursor {
				val = "▸" + val
			}
			cells = append(cells, renderCell(style, val, colWidths[j]))
		}
		lines = append(lines, strings.Join(cells, " "))
	}

	return strings.Join(lines, "\n")
}

func truncate(s string, maxLen int) string {
	if maxLen < 1 {
		return ""
	}
	if lipgloss.Width(s) <= maxLen {
		return s
	}
	if maxLen < 3 {
		return s[:maxLen]
	}
	// Cut by display width
	w := 0
	for i, r := range s {
		rw := lipgloss.Width(string(r))
		if w+rw > maxLen-1 {
			return s[:i] + "…"
		}
		w += rw
	}
	return s
}
