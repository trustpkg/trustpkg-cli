package logs

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"unicode"
	"unicode/utf8"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"
	"golang.org/x/term"
)

var (
	ErrPromptUnsupported = errors.New("prompt requires an interactive ANSI terminal")
	ErrPromptCancelled   = errors.New("prompt cancelled")
)

type promptBase struct {
	logger    *Logger
	title     string
	width     int
	height    int
	message   string
	submitted bool
	cancelled bool
}

func validatePromptText(name, value string, required bool) error {
	if !utf8.ValidString(value) {
		return fmt.Errorf("%s must be valid UTF-8", name)
	}
	if required && strings.TrimSpace(value) == "" {
		return fmt.Errorf("%s must be nonempty", name)
	}
	for _, r := range value {
		if unicode.IsControl(r) {
			return fmt.Errorf("%s must not contain control characters", name)
		}
	}
	return nil
}

func (base *promptBase) resize(width, height int) {
	base.width, base.height = max(1, width), max(1, height)
	if base.logger.hasMaxWidth {
		base.width = min(base.width, base.logger.maxWidth)
	}
}

func (base *promptBase) color(text, color string) string {
	return paint(text, color, base.logger.theme.Reset)
}

func (base *promptBase) line(text string) string {
	return ansi.Truncate(text, base.width, "")
}

func (base *promptBase) small() bool {
	return base.width < 12 || base.height < 6
}

func (base *promptBase) warning() string {
	text := strings.Map(func(r rune) rune {
		if unicode.IsControl(r) {
			return -1
		}
		return r
	}, base.message)
	return base.line(base.color(text, base.logger.theme.Warning))
}

func (logger *Logger) runPrompt(model interface {
	tea.Model
	resize(int, int)
}) error {
	if logger == nil || logger.output == nil {
		return fmt.Errorf("prompt requires an initialized logger")
	}
	facts := logger.terminalFacts
	if !facts.IsStdinTTY || !facts.IsStdoutTTY || !facts.SupportsANSI || logger.output != os.Stdout ||
		!term.IsTerminal(int(os.Stdin.Fd())) || !term.IsTerminal(int(os.Stdout.Fd())) {
		return ErrPromptUnsupported
	}

	logger.mu.Lock()
	if logger.promptActive || logger.windowActive || logger.footer != nil || logger.loaderActive {
		logger.mu.Unlock()
		return fmt.Errorf("prompt requires a logger without an active terminal UI, footer or loader")
	}
	if _, err := logger.limitedWidth(0); err != nil {
		logger.mu.Unlock()
		return err
	}
	if logger.hasMaxWidth && logger.maxWidth < 12 {
		logger.mu.Unlock()
		return fmt.Errorf("prompt requires a maximum width of at least 12 columns")
	}
	logger.promptActive = true
	logger.mu.Unlock()
	defer func() {
		logger.mu.Lock()
		logger.promptActive = false
		logger.mu.Unlock()
	}()

	state, err := term.GetState(int(os.Stdin.Fd()))
	if err != nil {
		return fmt.Errorf("get terminal state for prompt: %w", err)
	}
	width, height, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		return fmt.Errorf("get terminal size for prompt: %w", err)
	}
	model.resize(width, height)
	output := &windowOutput{File: os.Stdout, writer: logger.output}
	program := tea.NewProgram(model, tea.WithInput(os.Stdin), tea.WithOutput(output))
	_, runErr := program.Run()
	restoreErr := term.Restore(int(os.Stdin.Fd()), state)
	var cursorErr error
	if runErr != nil {
		_, cursorErr = fmt.Fprint(logger.output, "\x1b[?2004l\x1b[?25h")
	}
	if err := errors.Join(runErr, restoreErr, cursorErr, output.failure()); err != nil {
		return fmt.Errorf("run prompt: %w", err)
	}
	return nil
}
