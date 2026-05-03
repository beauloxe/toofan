package tui

import (
	"strings"
	"unicode/utf8"

	"github.com/charmbracelet/lipgloss"
	"github.com/vyrx-dev/toofan/internal/game"
	"github.com/vyrx-dev/toofan/internal/theme"
)

// colorText renders visible lines with typed/error/cursor/untyped styling
func colorText(g *game.Game, p theme.Palette, lines []string, top, bot int) string {
	ok := lipgloss.NewStyle().Foreground(p.Typed)
	bad := lipgloss.NewStyle().Foreground(p.Error).Underline(true)
	cur := lipgloss.NewStyle().Foreground(p.Background).Background(p.Cursor)
	dim := lipgloss.NewStyle().Foreground(p.Foreground)

	typed := utf8.RuneCountInString(g.Input())

	// character offset at top of visible window
	startPos := 0
	for i := 0; i < top; i++ {
		startPos += utf8.RuneCountInString(lines[i]) + 1 // +1 for the \n or space separator
	}

	var out strings.Builder
	pos := startPos

	for i := top; i < bot && i < len(lines); i++ {
		if i > top {
			out.WriteByte('\n')
			pos++ // skip the \n in position tracking
		}
		for _, ch := range lines[i] {
			s := string(ch)
			switch {
			case pos < typed && g.Errors()[pos]:
				out.WriteString(bad.Render(s))
			case pos < typed:
				out.WriteString(ok.Render(s))
			case pos == typed:
				out.WriteString(cur.Render(s))
			default:
				out.WriteString(dim.Render(s))
			}
			pos++
		}
	}
	return out.String()
}

// splitLines splits text for display.
// For code mode, splits on actual newlines (preserving indentation).
// For word mode, wraps at word boundaries.
func splitLines(text string, width int, codeMode bool) []string {
	if codeMode {
		return strings.Split(text, "\n")
	}
	return wrapText(text, width)
}

// wrapText breaks text into lines at word boundaries
func wrapText(text string, w int) []string {
	var lines []string
	var line strings.Builder
	lineLen := 0

	for _, word := range strings.Split(text, " ") {
		wordLen := lipgloss.Width(word)
		space := 0
		if lineLen > 0 {
			space = 1
		}

		if lineLen+space+wordLen > w && lineLen > 0 {
			lines = append(lines, line.String())
			line.Reset()
			lineLen = 0
			space = 0
		}

		if space > 0 {
			line.WriteByte(' ')
			lineLen++
		}
		line.WriteString(word)
		lineLen += wordLen
	}

	if line.Len() > 0 {
		lines = append(lines, line.String())
	}
	return lines
}

// cursorLine returns which line the cursor is on
func cursorLine(lines []string, inputLen int) int {
	pos := 0
	for i, line := range lines {
		end := pos + utf8.RuneCountInString(line)
		if inputLen <= end {
			return i
		}
		pos = end + 1
	}
	return max(0, len(lines)-1)
}

func visibleLineWindow(lineCount, curLine, visible int) (int, int) {
	if lineCount <= 0 || visible <= 0 {
		return 0, 0
	}
	if visible >= lineCount {
		return 0, lineCount
	}

	top := curLine - visible + 2
	top = max(0, top)
	top = min(top, lineCount-visible)

	return top, top + visible
}

// col renders text in a fixed-width column.
func col(w int, s string) string {
	return lipgloss.NewStyle().Width(w).Render(s)
}

type tableCell struct {
	width int
	text  string
	style lipgloss.Style
}

func tableHeader(cells ...tableCell) string {
	return tableRow(cells...)
}

func tableRow(cells ...tableCell) string {
	wrapped := make([][]string, len(cells))
	height := 1
	for i, cell := range cells {
		wrapped[i] = wrapCell(cell.text, cell.width)
		if len(wrapped[i]) > height {
			height = len(wrapped[i])
		}
	}

	var lines []string
	for line := 0; line < height; line++ {
		var b strings.Builder
		for i, cell := range cells {
			text := ""
			if line < len(wrapped[i]) {
				text = wrapped[i][line]
			}
			b.WriteString(fixedCell(cell.width, text, cell.style))
		}
		lines = append(lines, b.String())
	}
	return strings.Join(lines, "\n")
}

func wrapCell(s string, w int) []string {
	if w <= 1 {
		return []string{""}
	}
	s = strings.ReplaceAll(s, "\n", " ")
	contentWidth := w - 1
	if lipgloss.Width(s) <= contentWidth {
		return []string{s}
	}

	var lines []string
	var line strings.Builder
	lineWidth := 0
	for _, r := range s {
		rw := lipgloss.Width(string(r))
		if lineWidth > 0 && lineWidth+rw > contentWidth {
			lines = append(lines, line.String())
			line.Reset()
			lineWidth = 0
		}
		line.WriteRune(r)
		lineWidth += rw
	}
	if line.Len() > 0 {
		lines = append(lines, line.String())
	}
	return lines
}

func fixedCell(w int, s string, style lipgloss.Style) string {
	if w <= 0 {
		return ""
	}
	rendered := style.Render(s)
	pad := w - lipgloss.Width(rendered)
	if pad <= 0 {
		return rendered + " "
	}
	return rendered + strings.Repeat(" ", pad)
}
