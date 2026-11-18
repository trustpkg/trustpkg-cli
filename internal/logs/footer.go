package logs

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/mattn/go-runewidth"
	terminalParameters "github.com/trustpkg/trustpkg-cli-forge/internal/terminal-parameters"
	"golang.org/x/term"
)

var ErrFooterUnsupported = errors.New("footer requires an interactive ANSI terminal")

type footerState struct {
	status   string
	frame    int
	width    int
	height   int
	stop     chan struct{}
	done     chan struct{}
	err      error
	stopping bool
}

// StartFooter reserves two terminal rows until StopFooter is called.
func (logger *Logger) StartFooter(status string) error {
	if logger == nil || logger.output == nil {
		return fmt.Errorf("footer requires an initialized logger")
	}
	if err := validateFooterStatus(status); err != nil {
		return err
	}
	if !logger.terminalFacts.IsStdoutTTY || !logger.terminalFacts.SupportsANSI || !logger.terminalFacts.CanAnimate {
		return ErrFooterUnsupported
	}

	logger.mu.Lock()
	defer logger.mu.Unlock()
	if logger.windowActive || logger.promptActive {
		return fmt.Errorf("cannot start footer while a terminal UI is active")
	}
	if logger.footer != nil {
		return fmt.Errorf("footer already active")
	}
	width, height, err := logger.footerSize()
	if err != nil {
		return err
	}
	state := &footerState{status: status, width: width, height: height, stop: make(chan struct{}), done: make(chan struct{})}
	if _, err := fmt.Fprintf(logger.output, "\x1b[1;%dr\x1b[%d;1H", height-2, height-2); err != nil {
		_, restoreErr := fmt.Fprint(logger.output, "\x1b[r")
		return fmt.Errorf("reserve footer rows: %w", errors.Join(err, restoreErr))
	}
	if err := logger.renderFooter(state); err != nil {
		_, restoreErr := fmt.Fprint(logger.output, "\x1b[r")
		return fmt.Errorf("draw footer: %w", errors.Join(err, restoreErr))
	}
	logger.footer = state
	logger.loaderActive = false
	go logger.animateFooter(state)
	return nil
}

// UpdateFooter changes the status displayed beside the spinner.
func (logger *Logger) UpdateFooter(status string) error {
	if logger == nil {
		return fmt.Errorf("footer requires an initialized logger")
	}
	if err := validateFooterStatus(status); err != nil {
		return err
	}
	logger.mu.Lock()
	defer logger.mu.Unlock()
	state := logger.footer
	if state == nil || state.stopping {
		return fmt.Errorf("footer is not active")
	}
	if state.err != nil {
		return fmt.Errorf("footer animation: %w", state.err)
	}
	state.status = status
	if err := logger.renderFooter(state); err != nil {
		state.err = err
		return err
	}
	return nil
}

// StopFooter stops animation and restores the terminal's scrolling region.
func (logger *Logger) StopFooter() error {
	if logger == nil {
		return fmt.Errorf("footer requires an initialized logger")
	}
	logger.mu.Lock()
	state := logger.footer
	if state == nil || state.stopping {
		logger.mu.Unlock()
		return fmt.Errorf("footer is not active")
	}
	state.stopping = true
	close(state.stop)
	logger.mu.Unlock()

	<-state.done
	logger.mu.Lock()
	defer logger.mu.Unlock()
	_, err := fmt.Fprintf(logger.output, "\x1b7\x1b[r\x1b[%d;1H\x1b[2K\x1b[%d;1H\x1b[2K\x1b8", state.height-1, state.height)
	logger.footer = nil
	if joined := errors.Join(state.err, err); joined != nil {
		return fmt.Errorf("stop footer: %w", joined)
	}
	return nil
}

func (logger *Logger) footerSize() (int, int, error) {
	width, height := logger.terminalFacts.TerminalWidth, logger.terminalFacts.TerminalHeight
	if logger.output == os.Stdout {
		var err error
		width, height, err = term.GetSize(int(os.Stdout.Fd()))
		if err != nil {
			return 0, 0, fmt.Errorf("get terminal size for footer: %w", err)
		}
	}
	if width < 4 || height < 4 {
		return 0, 0, fmt.Errorf("terminal too small for footer: %d columns and %d rows", width, height)
	}
	width, err := logger.limitedWidth(width)
	if err != nil {
		return 0, 0, err
	}
	return width, height, nil
}

func (logger *Logger) renderFooter(state *footerState) error {
	width, height, err := logger.footerSize()
	if err != nil {
		return err
	}
	var resize, cursorAfterResize string
	if width != state.width || height != state.height {
		resize = fmt.Sprintf("\x1b[r\x1b[%d;1H\x1b[2K\x1b[%d;1H\x1b[2K\x1b[1;%dr",
			state.height-1, state.height, height-2)
		cursorAfterResize = fmt.Sprintf("\x1b[%d;1H", height-2)
		state.width, state.height = width, height
	}
	border := "-"
	if logger.terminalFacts.SupportsUnicode == terminalParameters.CapabilitySupported {
		border = "─"
	}
	status := ""
	prefix := ""
	spaces := ""
	if width >= 2 {
		prefix = " "
		if width >= 3 {
			status = runewidth.Truncate(state.status, width-3, "")
		}
		spaces = strings.Repeat(" ", width-2-runewidth.StringWidth(status))
	}
	_, err = fmt.Fprintf(logger.output,
		"\x1b7%s\x1b[%d;1H\x1b[2K%s\x1b[%d;1H\x1b[2K%s%s%s%s\x1b8%s",
		resize, height-1, paint(strings.Repeat(border, width), logger.theme.Border, logger.theme.Reset),
		height, prefix, paint(status, logger.theme.Description, logger.theme.Reset), spaces,
		paint(spinnerFrames[state.frame], logger.theme.Spinner, logger.theme.Reset), cursorAfterResize)
	if err != nil {
		return fmt.Errorf("render footer: %w", err)
	}
	return nil
}

func validateFooterStatus(status string) error {
	if !utf8.ValidString(status) {
		return fmt.Errorf("footer status must be valid UTF-8")
	}
	for _, r := range status {
		if unicode.IsControl(r) {
			return fmt.Errorf("footer status must not contain control characters")
		}
	}
	return nil
}

// Log writes a message above an active footer when called on the same logger.
func (logger *Logger) Log(message string, severity ...Severity) error {
	if logger == nil || logger.output == nil {
		return fmt.Errorf("log requires an initialized logger")
	}
	if len(severity) > 1 {
		return fmt.Errorf("log accepts at most one severity")
	}
	level := SeverityLog
	if len(severity) == 1 {
		level = severity[0]
	}
	color, err := logger.severityColor(level)
	if err != nil {
		return err
	}
	for _, r := range message {
		if unicode.IsControl(r) && r != '\n' && r != '\t' {
			return fmt.Errorf("log message must not contain terminal control characters")
		}
	}
	if logger.hasMaxWidth {
		width, err := logger.limitedWidth(0)
		if err != nil {
			return err
		}
		message, err = wrapLogMessage(message, width)
		if err != nil {
			return err
		}
	}
	return logger.writeLog(message, color)
}

func (logger *Logger) writeLog(message, color string) error {
	logger.mu.Lock()
	defer logger.mu.Unlock()
	return logger.writeLogLocked(message, color)
}

func (logger *Logger) writeLogLocked(message, color string) error {
	if logger.windowActive || logger.promptActive {
		return fmt.Errorf("cannot write to logger while a terminal UI is active")
	}
	if logger.footer != nil && logger.footer.err != nil {
		return fmt.Errorf("footer animation: %w", logger.footer.err)
	}
	panel, isPanel := logger.output.(*WindowPanel)
	if (logger.footer != nil || isPanel) && logger.loaderActive {
		if !isPanel || panel.window.interactive {
			if _, err := fmt.Fprint(logger.output, "\r\n"); err != nil {
				return fmt.Errorf("finish loader line before log: %w", err)
			}
		}
		logger.loaderActive = false
	}
	if logger.terminalFacts.SupportsANSI {
		message = paint(message, color, logger.theme.Reset)
	}
	if _, err := fmt.Fprintln(logger.output, message); err != nil {
		return fmt.Errorf("write log: %w", err)
	}
	return nil
}
