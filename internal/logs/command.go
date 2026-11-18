package logs

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/mattn/go-runewidth"
	terminalParameters "github.com/trustpkg/trustpkg-cli/internal/terminal-parameters"
)

func (logger *Logger) Command(command string) error {
	if logger == nil || logger.output == nil {
		return fmt.Errorf("command requires an initialized logger")
	}
	if !utf8.ValidString(command) {
		return fmt.Errorf("command must be valid UTF-8")
	}
	for _, r := range command {
		if unicode.IsControl(r) {
			return fmt.Errorf("command must not contain control characters")
		}
	}

	width, err := logger.displayWidth("command")
	if err != nil {
		return err
	}
	command = wrapCommand(command, width)
	borderCharacter := "-"
	if logger.terminalFacts.SupportsUnicode == terminalParameters.CapabilitySupported {
		borderCharacter = "─"
	}
	border := strings.Repeat(borderCharacter, width)
	if logger.terminalFacts.SupportsANSI {
		border = paint(border, logger.theme.Border, logger.theme.Reset)
		command = paint(command, logger.theme.Code, logger.theme.Reset)
	}
	return logger.writeLog(border+"\n"+command+"\n"+border, "")
}

func wrapCommand(command string, width int) string {
	if runewidth.StringWidth(command) <= width {
		return command
	}

	var breaks []int
	var quote byte
	for i := 0; i < len(command); i++ {
		c := command[i]
		if quote == '\'' {
			if c == '\'' {
				quote = 0
			}
			continue
		}
		if c == '\\' {
			i++
			if i == len(command) {
				return command
			}
			continue
		}
		if c == '$' || c == '`' {
			return command
		}
		if quote == '"' {
			if c == '"' {
				quote = 0
			}
			continue
		}
		switch c {
		case '\'', '"':
			quote = c
		case '#', '<', '>', ';', '&', '|', '(', ')':
			return command
		case ' ':
			breaks = append(breaks, i)
		}
	}
	if quote != 0 {
		return command
	}

	var result strings.Builder
	start := 0
	for runewidth.StringWidth(command[start:]) > width {
		split := -1
		for _, candidate := range breaks {
			if candidate <= start {
				continue
			}
			if runewidth.StringWidth(command[start:candidate])+1 > width {
				break
			}
			split = candidate
		}
		if split < 0 {
			return command
		}
		result.WriteString(command[start:split])
		result.WriteString("\\\n")
		start = split
	}
	result.WriteString(command[start:])
	return result.String()
}
