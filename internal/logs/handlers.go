package logs

import (
	"io"
	"os"
	"sync"

	terminalParameters "github.com/trustpkg/trustpkg-cli/internal/terminal-parameters"
	"github.com/trustpkg/trustpkg-cli/internal/theme"
)

type Logger struct {
	terminalFacts terminalParameters.Facts
	theme         theme.Theme
	maxWidth      int
	hasMaxWidth   bool
	output        io.Writer
	mu            sync.Mutex
	footer        *footerState
	loaderActive  bool
	sections      []*sectionState
	windowActive  bool
	promptActive  bool
}

type NewLoggerInput struct {
	Facts    terminalParameters.Facts
	Theme    theme.Theme
	MaxWidth *int
}

func NewLogger(input NewLoggerInput) *Logger {
	logger := &Logger{terminalFacts: input.Facts, theme: input.Theme, output: os.Stdout}
	if input.MaxWidth != nil {
		logger.maxWidth = *input.MaxWidth
		logger.hasMaxWidth = true
	}
	return logger
}
