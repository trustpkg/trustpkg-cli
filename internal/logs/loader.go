package logs

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/mattn/go-runewidth"
)

const defaultLoaderWidth = 80

func (logger *Logger) Loader(steps, currentStep int, options ...LoaderOption) error {
	if logger == nil || logger.output == nil {
		return fmt.Errorf("loader requires an initialized logger")
	}
	if steps <= 0 || currentStep < 0 || currentStep > steps {
		return fmt.Errorf("invalid loader progress: %d of %d", currentStep, steps)
	}
	config := loaderOptions{severity: SeverityInfo}
	for _, option := range options {
		if option == nil {
			return fmt.Errorf("loader option cannot be nil")
		}
		switch value := option.(type) {
		case *Severity:
			if value == nil {
				return fmt.Errorf("loader severity cannot be nil")
			}
		case *LoaderLabel:
			if value == nil {
				return fmt.Errorf("loader label cannot be nil")
			}
		}
		if err := option.applyLoaderOption(&config); err != nil {
			return err
		}
	}
	if !utf8.ValidString(config.label) {
		return fmt.Errorf("loader label must be valid UTF-8")
	}
	for _, r := range config.label {
		if unicode.IsControl(r) {
			return fmt.Errorf("loader label must not contain control characters")
		}
	}
	color, err := logger.severityColor(config.severity)
	if err != nil {
		return err
	}

	logger.mu.Lock()
	defer logger.mu.Unlock()

	if logger.windowActive || logger.promptActive {
		return fmt.Errorf("cannot write to logger while a terminal UI is active")
	}
	if logger.footer != nil && logger.footer.err != nil {
		return fmt.Errorf("footer animation: %w", logger.footer.err)
	}

	width, err := logger.displayWidth("loader")
	if err != nil {
		return err
	}

	line := logger.loaderLine(steps, currentStep, width, color, config.label)
	end := ""
	if currentStep == steps {
		end = "\r\n"
	}
	if _, err := fmt.Fprint(logger.output, "\r", line, end); err != nil {
		return fmt.Errorf("write loader progress: %w", err)
	}
	logger.loaderActive = currentStep < steps
	return nil
}

func (logger *Logger) severityColor(severity Severity) (string, error) {
	switch severity {
	case SeverityLog:
		return logger.theme.Log, nil
	case SeverityDebug:
		return logger.theme.Debug, nil
	case SeverityInfo:
		return logger.theme.Info, nil
	case SeverityWarning:
		return logger.theme.Warning, nil
	case SeverityError:
		return logger.theme.Error, nil
	case SeveritySuccess:
		return logger.theme.Success, nil
	case SeverityStart:
		return logger.theme.InProgress, nil
	case SeverityDone:
		return logger.theme.Done, nil
	default:
		return "", fmt.Errorf("invalid severity %q", severity)
	}
}

func (logger *Logger) loaderLine(steps, currentStep, width int, color, label string) string {
	ratio := float64(currentStep) / float64(steps)
	percentage := int(ratio * 100)
	if currentStep < steps && percentage == 100 {
		percentage = 99
	}
	suffix := fmt.Sprintf(" %d/%d %d%%", currentStep, steps, percentage)
	compact := fmt.Sprintf("%d%%", percentage)
	labelWidth := width - len(suffix) - 4
	if labelWidth <= 0 {
		labelWidth = width - len(compact) - 1
	}
	prefix := ""
	if label != "" && labelWidth > 0 {
		prefix = runewidth.Truncate(label, labelWidth, "")
		if prefix != "" {
			prefix += " "
		}
	}
	prefixWidth := runewidth.StringWidth(prefix)
	barWidth := width - prefixWidth - len(suffix) - 2
	displayPrefix := prefix
	if logger.terminalFacts.SupportsANSI {
		displayPrefix = paint(prefix, logger.theme.Text, logger.theme.Reset)
	}
	if barWidth < 1 {
		available := width - prefixWidth
		if len(compact) > available {
			compact = compact[:available]
		}
		return displayPrefix + compact + strings.Repeat(" ", available-len(compact))
	}

	filled := int(ratio * float64(barWidth))
	if currentStep < steps && filled == barWidth {
		filled--
	}
	completed := strings.Repeat("=", filled)
	if filled > 0 && currentStep < steps {
		completed = completed[:len(completed)-1] + ">"
	}
	remaining := strings.Repeat(" ", barWidth-filled)
	if !logger.terminalFacts.SupportsANSI {
		return displayPrefix + "[" + completed + remaining + "]" + suffix
	}
	return displayPrefix + "[" + paint(completed, color, logger.theme.Reset) +
		paint(remaining, logger.theme.ProgressRemaining, logger.theme.Reset) + "]" +
		paint(suffix, logger.theme.Text, logger.theme.Reset)
}

func paint(text, color, reset string) string {
	if text == "" || color == "" {
		return text
	}
	return color + text + reset
}
