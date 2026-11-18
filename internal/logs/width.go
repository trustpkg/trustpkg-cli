package logs

import (
	"fmt"
	"os"
	"strings"
	"unicode/utf8"

	"github.com/mattn/go-runewidth"
	"golang.org/x/term"
)

func (logger *Logger) displayWidth(component string) (int, error) {
	width := logger.terminalFacts.TerminalWidth
	if panel, ok := logger.output.(*WindowPanel); ok {
		width = panel.contentWidth()
	} else if logger.terminalFacts.IsStdoutTTY && logger.output == os.Stdout {
		var err error
		width, _, err = term.GetSize(int(os.Stdout.Fd()))
		if err != nil {
			return 0, fmt.Errorf("get terminal width for %s: %w", component, err)
		}
	}
	if width <= 0 {
		width = defaultLoaderWidth
	}
	return logger.limitedWidth(width)
}

func (logger *Logger) limitedWidth(width int) (int, error) {
	if !logger.hasMaxWidth {
		return width, nil
	}
	if logger.maxWidth <= 0 {
		return 0, fmt.Errorf("maximum width must be positive, got %d", logger.maxWidth)
	}
	if width <= 0 || logger.maxWidth < width {
		return logger.maxWidth, nil
	}
	return width, nil
}

func wrapLogMessage(message string, width int) (string, error) {
	if !utf8.ValidString(message) {
		return "", fmt.Errorf("log message must be valid UTF-8 when a maximum width is set")
	}
	lines := strings.Split(message, "\n")
	for i, line := range lines {
		var expanded strings.Builder
		column := 0
		for j, part := range strings.Split(line, "\t") {
			if j > 0 {
				padding := 8 - column%8
				expanded.WriteString(strings.Repeat(" ", padding))
				column += padding
			}
			expanded.WriteString(part)
			column += runewidth.StringWidth(part)
		}
		lines[i] = runewidth.Wrap(expanded.String(), width)
		for _, wrapped := range strings.Split(lines[i], "\n") {
			if runewidth.StringWidth(wrapped) > width {
				return "", fmt.Errorf("log contains a character wider than maximum width %d", width)
			}
		}
	}
	return strings.Join(lines, "\n"), nil
}
