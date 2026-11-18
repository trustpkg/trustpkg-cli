package logs

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

	terminalParameters "github.com/trustpkg/trustpkg-cli/internal/terminal-parameters"
)

type sectionState struct {
	closed bool
}

func (logger *Logger) StartSection(title string, options ...SectionOption) (func() error, error) {
	if logger == nil || logger.output == nil {
		return nil, fmt.Errorf("section requires an initialized logger")
	}
	var config sectionOptions
	for _, option := range options {
		if option == nil {
			return nil, fmt.Errorf("section option cannot be nil")
		}
		switch value := option.(type) {
		case *Severity:
			if value == nil {
				return nil, fmt.Errorf("section severity cannot be nil")
			}
		case *SectionDashed:
			if value == nil {
				return nil, fmt.Errorf("section line style cannot be nil")
			}
		}
		if err := option.applySectionOption(&config); err != nil {
			return nil, err
		}
	}
	color := logger.theme.Separator
	if config.severitySet {
		var err error
		color, err = logger.severityColor(config.severity)
		if err != nil {
			return nil, err
		}
	}
	if !utf8.ValidString(title) || strings.TrimSpace(title) == "" {
		return nil, fmt.Errorf("section title must be nonempty valid UTF-8")
	}
	for _, r := range title {
		if unicode.IsControl(r) {
			return nil, fmt.Errorf("section title must not contain control characters")
		}
	}
	width, err := logger.displayWidth("section")
	if err != nil {
		return nil, err
	}
	title, err = wrapLogMessage(title, width)
	if err != nil {
		return nil, fmt.Errorf("section title: %w", err)
	}

	logger.mu.Lock()
	defer logger.mu.Unlock()
	border := logger.sectionBorder(width, color, config.dashed)
	if logger.terminalFacts.SupportsANSI {
		title = paint(title, logger.theme.Heading, logger.theme.Reset)
	}
	if err := logger.writeLogLocked(border+"\n"+title, ""); err != nil {
		return nil, err
	}
	state := &sectionState{}
	logger.sections = append(logger.sections, state)
	return func() error {
		logger.mu.Lock()
		defer logger.mu.Unlock()
		if state.closed {
			return fmt.Errorf("section is already closed")
		}
		if len(logger.sections) == 0 || logger.sections[len(logger.sections)-1] != state {
			return fmt.Errorf("section is not the active section")
		}
		width, err := logger.displayWidth("section")
		if err != nil {
			return err
		}
		if err := logger.writeLogLocked(logger.sectionBorder(width, color, config.dashed), ""); err != nil {
			return err
		}
		state.closed = true
		logger.sections = logger.sections[:len(logger.sections)-1]
		return nil
	}, nil
}

func (logger *Logger) sectionBorder(width int, color string, dashed bool) string {
	character := "-"
	if logger.terminalFacts.SupportsUnicode == terminalParameters.CapabilitySupported {
		character = "─"
		if dashed {
			character = "┄"
		}
	}
	border := strings.Repeat(character, width)
	if dashed && character == "-" {
		border = strings.Repeat("- ", (width+1)/2)[:width]
	}
	if logger.terminalFacts.SupportsANSI {
		border = paint(border, color, logger.theme.Reset)
	}
	return border
}
