package logs

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"unicode"
	"unicode/utf8"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"
	"golang.org/x/term"
)

type Window struct {
	mu          sync.Mutex
	parent      *Logger
	panels      []*WindowPanel
	interactive bool
	program     *tea.Program
	output      *windowOutput
	done        chan struct{}
	err         error
	closed      bool
	closing     bool
	width       int
	height      int
	focused     int
	frame       int
}

type WindowPanel struct {
	window     *Window
	name       string
	logger     *Logger
	lines      []string
	current    string
	offset     int
	finished   bool
	removed    bool
	cacheWidth int
	cacheCount int
	cacheLines []string
	cacheStyle string
	lastRows   int
	lastWidth  int
}

type windowOutput struct {
	*os.File
	mu     sync.Mutex
	writer io.Writer
	err    error
}

func (output *windowOutput) Write(data []byte) (int, error) {
	output.mu.Lock()
	defer output.mu.Unlock()
	if output.err != nil {
		return 0, output.err
	}
	n, err := output.writer.Write(data)
	if err == nil && n != len(data) {
		err = io.ErrShortWrite
	}
	if err != nil {
		output.err = err
	}
	return n, err
}

func (output *windowOutput) failure() error {
	output.mu.Lock()
	defer output.mu.Unlock()
	return output.err
}

func (logger *Logger) StartWindow() (*Window, error) {
	if logger == nil || logger.output == nil {
		return nil, fmt.Errorf("window requires an initialized logger")
	}
	logger.mu.Lock()
	if logger.windowActive || logger.promptActive || logger.footer != nil || len(logger.sections) != 0 || logger.loaderActive {
		logger.mu.Unlock()
		return nil, fmt.Errorf("window requires a logger without an active window, footer, section or loader")
	}
	if _, ok := logger.output.(*WindowPanel); ok {
		logger.mu.Unlock()
		return nil, fmt.Errorf("cannot start a window inside a panel")
	}
	if _, err := logger.limitedWidth(0); err != nil {
		logger.mu.Unlock()
		return nil, err
	}
	logger.windowActive = true
	logger.mu.Unlock()

	facts := logger.terminalFacts
	window := &Window{
		parent:      logger,
		interactive: facts.IsStdinTTY && facts.IsStdoutTTY && facts.SupportsANSI && logger.output == os.Stdout,
		width:       facts.TerminalWidth,
		height:      facts.TerminalHeight,
	}
	if window.width <= 0 {
		window.width = defaultLoaderWidth
	}
	if logger.hasMaxWidth {
		window.width = min(window.width, logger.maxWidth)
	}
	if window.height <= 0 {
		window.height = 24
	}
	if !window.interactive {
		return window, nil
	}

	terminalState, err := term.GetState(int(os.Stdin.Fd()))
	if err != nil {
		logger.mu.Lock()
		logger.windowActive = false
		logger.mu.Unlock()
		return nil, fmt.Errorf("get terminal state for window: %w", err)
	}
	ready := make(chan struct{})
	window.done = make(chan struct{})
	window.output = &windowOutput{File: os.Stdout, writer: logger.output}
	window.program = tea.NewProgram(&windowModel{window: window, ready: ready}, tea.WithInput(os.Stdin), tea.WithOutput(window.output))
	go func() {
		_, err := window.program.Run()
		restoreErr := term.Restore(int(os.Stdin.Fd()), terminalState)
		if err != nil {
			_, cursorErr := fmt.Fprint(logger.output, "\x1b[?2004l\x1b[?25h")
			err = errors.Join(err, cursorErr)
		}
		err = errors.Join(err, restoreErr)
		window.mu.Lock()
		if err != nil {
			window.err = errors.Join(window.err, err)
		}
		window.mu.Unlock()
		close(window.done)
	}()
	select {
	case <-ready:
		return window, nil
	case <-window.done:
		logger.mu.Lock()
		logger.windowActive = false
		logger.mu.Unlock()
		window.mu.Lock()
		err := window.err
		window.mu.Unlock()
		if err == nil {
			err = fmt.Errorf("terminal window stopped during startup")
		}
		return nil, err
	}
}

func (window *Window) AddPanel(name string) (*WindowPanel, error) {
	if window == nil {
		return nil, fmt.Errorf("window is nil")
	}
	if !utf8.ValidString(name) || strings.TrimSpace(name) == "" {
		return nil, fmt.Errorf("panel name must be nonempty valid UTF-8")
	}
	for _, r := range name {
		if unicode.IsControl(r) {
			return nil, fmt.Errorf("panel name must not contain control characters")
		}
	}
	window.mu.Lock()
	defer window.mu.Unlock()
	if err := window.checkActive(); err != nil {
		return nil, err
	}
	for _, panel := range window.panels {
		if panel.name == name {
			return nil, fmt.Errorf("panel %q already exists", name)
		}
	}
	panel := &WindowPanel{window: window, name: name}
	facts := window.parent.terminalFacts
	facts.IsStdoutTTY = false
	facts.SupportsANSI = window.interactive
	facts.CanAnimate = false
	input := NewLoggerInput{Facts: facts, Theme: window.parent.theme}
	if window.parent.hasMaxWidth {
		input.MaxWidth = &window.parent.maxWidth
	}
	panel.logger = NewLogger(input)
	panel.logger.output = panel
	window.panels = append(window.panels, panel)
	if !window.interactive {
		if err := window.writeStream(name, "started"); err != nil {
			window.panels = window.panels[:len(window.panels)-1]
			return nil, fmt.Errorf("write panel start: %w", err)
		}
	}
	return panel, nil
}

func (panel *WindowPanel) Logger() *Logger {
	if panel == nil {
		return nil
	}
	return panel.logger
}

func (panel *WindowPanel) Finish() error {
	if panel == nil || panel.window == nil {
		return fmt.Errorf("panel is nil")
	}
	window := panel.window
	window.mu.Lock()
	defer window.mu.Unlock()
	if err := window.checkActive(); err != nil {
		return err
	}
	if panel.removed || panel.finished {
		return fmt.Errorf("panel %q is already finished or removed", panel.name)
	}
	if !window.interactive {
		if err := window.writeStream(panel.name, "finished"); err != nil {
			return fmt.Errorf("write panel finish: %w", err)
		}
	}
	panel.finished = true
	return nil
}

func (window *Window) RemovePanel(panel *WindowPanel) error {
	if window == nil || panel == nil || panel.window != window {
		return fmt.Errorf("panel does not belong to window")
	}
	window.mu.Lock()
	defer window.mu.Unlock()
	if err := window.checkActive(); err != nil {
		return err
	}
	for i, candidate := range window.panels {
		if candidate == panel {
			if !window.interactive {
				if err := window.writeStream(panel.name, "removed"); err != nil {
					return fmt.Errorf("write panel removal: %w", err)
				}
			}
			panel.removed = true
			window.panels = append(window.panels[:i], window.panels[i+1:]...)
			if i < window.focused {
				window.focused--
			}
			if window.focused >= len(window.panels) {
				window.focused = max(0, len(window.panels)-1)
			}
			return nil
		}
	}
	return fmt.Errorf("panel %q is already removed", panel.name)
}

func (window *Window) Close() error {
	if window == nil {
		return fmt.Errorf("window is nil")
	}
	window.mu.Lock()
	if window.closed || window.closing {
		window.mu.Unlock()
		return fmt.Errorf("window is already closed")
	}
	window.closing = true
	program, done := window.program, window.done
	window.mu.Unlock()
	if program != nil {
		program.Quit()
		<-done
	}
	window.mu.Lock()
	window.closed = true
	err := window.err
	if window.output != nil {
		err = errors.Join(err, window.output.failure())
	}
	window.mu.Unlock()
	window.parent.mu.Lock()
	window.parent.windowActive = false
	window.parent.mu.Unlock()
	return err
}

func (window *Window) checkActive() error {
	if window.err != nil {
		return fmt.Errorf("window: %w", window.err)
	}
	if window.output != nil {
		if err := window.output.failure(); err != nil {
			return fmt.Errorf("window output: %w", err)
		}
	}
	if window.closed || window.closing {
		return fmt.Errorf("window is closed")
	}
	return nil
}

func (panel *WindowPanel) contentWidth() int {
	window := panel.window
	window.mu.Lock()
	defer window.mu.Unlock()
	if !window.interactive {
		return window.width
	}
	columns, _, _ := window.layout()
	return max(1, window.width/columns-2)
}

func (panel *WindowPanel) Write(data []byte) (int, error) {
	window := panel.window
	window.mu.Lock()
	defer window.mu.Unlock()
	if err := window.checkActive(); err != nil {
		return 0, err
	}
	if panel.finished || panel.removed {
		return 0, fmt.Errorf("panel %q is finished or removed", panel.name)
	}
	if !window.interactive {
		text := strings.TrimRight(strings.ReplaceAll(ansi.Strip(string(data)), "\r", ""), "\n")
		for _, line := range strings.Split(text, "\n") {
			if err := window.writeStream(panel.name, line); err != nil {
				return 0, fmt.Errorf("write panel log: %w", err)
			}
		}
		return len(data), nil
	}

	text := string(data)
	for len(text) > 0 {
		i := strings.IndexAny(text, "\r\n")
		if i < 0 {
			panel.current += text
			break
		}
		panel.current += text[:i]
		if text[i] == '\r' && (i+1 == len(text) || text[i+1] != '\n') {
			panel.current = ""
		} else {
			panel.lines = append(panel.lines, panel.current)
			panel.current = ""
			if text[i] == '\r' {
				i++
			}
		}
		text = text[i+1:]
	}
	return len(data), nil
}

// writeStream requires window.mu.
func (window *Window) writeStream(name, text string) error {
	line := fmt.Sprintf("[%s] %s", name, text)
	if window.parent.hasMaxWidth {
		var err error
		line, err = wrapLogMessage(line, window.parent.maxWidth)
		if err != nil {
			return err
		}
	}
	n, err := fmt.Fprintln(window.parent.output, line)
	if err == nil && n != len(line)+1 {
		err = io.ErrShortWrite
	}
	return err
}
