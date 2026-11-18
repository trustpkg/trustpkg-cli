package logs

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"
	terminalParameters "github.com/trustpkg/trustpkg-cli/internal/terminal-parameters"
)

type windowModel struct {
	window *Window
	ready  chan struct{}
}

type windowTick time.Time
type windowReady struct{}

func nextWindowTick() tea.Cmd {
	return tea.Tick(spinnerInterval, func(t time.Time) tea.Msg { return windowTick(t) })
}

func (model *windowModel) Init() tea.Cmd {
	return tea.Batch(nextWindowTick(), func() tea.Msg { return windowReady{} })
}

func (model *windowModel) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	window := model.window
	window.mu.Lock()
	defer window.mu.Unlock()
	switch message := message.(type) {
	case windowReady:
		close(model.ready)
	case tea.WindowSizeMsg:
		window.width, window.height = max(1, message.Width), max(1, message.Height)
		if window.parent.hasMaxWidth {
			window.width = min(window.width, window.parent.maxWidth)
		}
	case windowTick:
		window.frame = (window.frame + 1) % len(spinnerFrames)
		return model, nextWindowTick()
	case tea.KeyMsg:
		switch message.String() {
		case "ctrl+c":
			window.err = fmt.Errorf("window interrupted")
			return model, tea.Quit
		case "tab":
			if len(window.panels) > 0 {
				window.focused = (window.focused + 1) % len(window.panels)
			}
		case "shift+tab":
			if len(window.panels) > 0 {
				window.focused = (window.focused + len(window.panels) - 1) % len(window.panels)
			}
		case "up":
			if len(window.panels) > 0 {
				window.panels[window.focused].offset++
			}
		case "down":
			if len(window.panels) > 0 && window.panels[window.focused].offset > 0 {
				window.panels[window.focused].offset--
			}
		case "end":
			if len(window.panels) > 0 {
				window.panels[window.focused].offset = 0
			}
		}
	}
	return model, nil
}

// layout requires window.mu and pages panels when the terminal is too small.
func (window *Window) layout() (columns, rows, first int) {
	columns = max(1, min(len(window.panels), window.width/24))
	capacity := columns * max(1, (window.height-1)/3)
	first = (window.focused / capacity) * capacity
	visible := min(capacity, len(window.panels)-first)
	rows = max(1, (visible+columns-1)/columns)
	return columns, rows, first
}

func (model *windowModel) View() string {
	window := model.window
	window.mu.Lock()
	defer window.mu.Unlock()
	width, height := window.width, window.height
	if width < 4 || height < 4 {
		return ansi.Truncate("Terminal too small", width, "")
	}
	columns, rows, first := window.layout()
	availableHeight := height - 1
	var lines []string
	for row := 0; row < rows; row++ {
		rowHeight := availableHeight / rows
		if row < availableHeight%rows {
			rowHeight++
		}
		cells := make([][]string, columns)
		for col := 0; col < columns; col++ {
			cellWidth := width / columns
			if col < width%columns {
				cellWidth++
			}
			index := first + row*columns + col
			if index < len(window.panels) {
				cells[col] = window.panelView(window.panels[index], index == window.focused, cellWidth, rowHeight)
			} else {
				cells[col] = emptyCell(cellWidth, rowHeight)
			}
		}
		for line := 0; line < rowHeight; line++ {
			var combined strings.Builder
			for _, cell := range cells {
				combined.WriteString(cell[line])
			}
			lines = append(lines, combined.String())
		}
	}
	help := "Tab: panel  Up/Down: scroll  End: latest  Ctrl+C: interrupt"
	lines = append(lines, padANSI(ansi.Truncate(help, width, ""), width))
	return strings.Join(lines, "\n")
}

func emptyCell(width, height int) []string {
	lines := make([]string, height)
	for i := range lines {
		lines[i] = strings.Repeat(" ", width)
	}
	return lines
}

func padANSI(text string, width int) string {
	text = ansi.Truncate(text, width, "")
	return text + "\x1b[0m" + strings.Repeat(" ", max(0, width-ansi.StringWidth(text)))
}

// panelView requires window.mu.
func (window *Window) panelView(panel *WindowPanel, focused bool, width, height int) []string {
	if width < 4 || height < 3 {
		return emptyCell(width, height)
	}
	inner := width - 2
	marker := " "
	if focused {
		marker = ">"
	}
	spinner := spinnerFrames[window.frame]
	if panel.finished {
		spinner = "✓"
		if window.parent.terminalFacts.SupportsUnicode != terminalParameters.CapabilitySupported {
			spinner = "+"
		}
	}
	title := ansi.Truncate(marker+" "+panel.name, max(0, inner-2), "")
	title = padANSI(title, max(0, inner-2)) + " " + paint(spinner, window.parent.theme.Spinner, window.parent.theme.Reset)
	color := window.parent.theme.Border
	if focused {
		color = window.parent.theme.Accent
	}
	border := func(text string) string { return paint(text, color, window.parent.theme.Reset) }
	topLeft, topRight, side, bottomLeft, bottomRight, horizontal := "+", "+", "|", "+", "+", "-"
	if window.parent.terminalFacts.SupportsUnicode == terminalParameters.CapabilitySupported {
		topLeft, topRight, side, bottomLeft, bottomRight, horizontal = "╭", "╮", "│", "╰", "╯", "─"
	}
	lines := []string{border(topLeft) + title + border(topRight)}
	if panel.cacheWidth != inner {
		panel.cacheWidth, panel.cacheCount = inner, 0
		panel.cacheLines, panel.cacheStyle = nil, ""
	}
	for panel.cacheCount < len(panel.lines) {
		panel.cacheLines, panel.cacheStyle = appendStyledHistory(
			panel.cacheLines, panel.lines[panel.cacheCount], panel.cacheStyle, inner,
		)
		panel.cacheCount++
	}
	history := panel.cacheLines
	if panel.current != "" {
		history, _ = appendStyledHistory(history, panel.current, panel.cacheStyle, inner)
	}
	if panel.offset > 0 && panel.lastWidth == inner && len(history) > panel.lastRows {
		panel.offset += len(history) - panel.lastRows
	}
	panel.lastRows, panel.lastWidth = len(history), inner
	visible := height - 2
	maxOffset := max(0, len(history)-visible)
	panel.offset = min(panel.offset, maxOffset)
	end := max(0, len(history)-panel.offset)
	start := max(0, end-visible)
	for i := 0; i < visible; i++ {
		text := ""
		if start+i < end {
			text = history[start+i]
		}
		lines = append(lines, border(side)+padANSI(text, inner)+border(side))
	}
	lines = append(lines, border(bottomLeft+strings.Repeat(horizontal, inner)+bottomRight))
	return lines
}

func appendStyledHistory(history []string, line, style string, width int) ([]string, string) {
	if strings.Contains(line, "\t") {
		var expanded strings.Builder
		column := 0
		for i, part := range strings.Split(line, "\t") {
			if i > 0 {
				padding := 8 - column%8
				expanded.WriteString(strings.Repeat(" ", padding))
				column += padding
			}
			expanded.WriteString(part)
			column += ansi.StringWidth(part)
		}
		line = expanded.String()
	}
	for _, part := range strings.Split(ansi.Hardwrap(line, width, true), "\n") {
		history = append(history, style+part)
		remaining := part
		for {
			start := strings.Index(remaining, "\x1b[")
			if start < 0 {
				break
			}
			remaining = remaining[start+2:]
			end := strings.IndexByte(remaining, 'm')
			if end < 0 {
				break
			}
			params := remaining[:end]
			if params == "" || params == "0" || strings.HasPrefix(params, "0;") {
				style = ""
			}
			if params != "" && params != "0" {
				style += "\x1b[" + params + "m"
			}
			remaining = remaining[end+1:]
		}
	}
	return history, style
}
