package logs

import (
	"fmt"
	"strings"
	"time"
)

func (logger *Logger) Print(message string, options ...PrintOption) error {
	if logger == nil || logger.output == nil {
		return fmt.Errorf("print requires an initialized logger")
	}
	config := printOptions{severity: SeverityLog}
	for _, option := range options {
		if option == nil {
			return fmt.Errorf("print option cannot be nil")
		}
		switch value := option.(type) {
		case *Severity:
			if value == nil {
				return fmt.Errorf("print severity cannot be nil")
			}
		case *PrintTimestamp:
			if value == nil {
				return fmt.Errorf("print timestamp option cannot be nil")
			}
		}
		if err := option.applyPrintOption(&config); err != nil {
			return err
		}
	}

	if _, err := logger.severityColor(config.severity); err != nil {
		return err
	}
	var prefix strings.Builder
	if config.timestamp {
		fmt.Fprintf(&prefix, "[%s] ", time.Now().Format("2006-01-02 15:04:05"))
	}
	if config.severitySet {
		level := string(config.severity)
		fmt.Fprintf(&prefix, "%s%s: ", strings.ToUpper(level[:1]), level[1:])
	}
	return logger.Log(prefix.String()+message, config.severity)
}
