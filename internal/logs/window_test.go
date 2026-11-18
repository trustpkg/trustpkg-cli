package logs

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"
	terminalParameters "github.com/trustpkg/trustpkg-cli/internal/terminal-parameters"
	"github.com/trustpkg/trustpkg-cli/internal/theme"
	"golang.org/x/term"
)

func TestWindowFallbackAndLifecycle(t *testing.T) {
	var output bytes.Buffer
	logger := NewLogger(NewLoggerInput{})
	logger.output = &output
	window, err := logger.StartWindow()
	if err != nil {
		t.Fatal(err)
	}
	panel, err := window.AddPanel("download")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := window.AddPanel("download"); err == nil {
		t.Fatal("duplicate panel name accepted")
	}
	if err := logger.Print("outside"); err == nil {
		t.Fatal("parent logger accepted a write during the window")
	}
	if err := panel.Logger().Print("starting", SeverityInfo); err != nil {
		t.Fatal(err)
	}
	if err := panel.Logger().Loader(2, 1); err != nil {
		t.Fatal(err)
	}
	if err := panel.Finish(); err != nil {
		t.Fatal(err)
	}
	if err := panel.Logger().Print("late"); err == nil {
		t.Fatal("finished panel accepted a write")
	}
	if err := window.RemovePanel(panel); err != nil {
		t.Fatal(err)
	}
	if err := window.Close(); err != nil {
		t.Fatal(err)
	}
	if err := logger.Print("outside"); err != nil {
		t.Fatal(err)
	}
	got := output.String()
	for _, want := range []string{"[download] started\n", "[download] Info: starting\n", "[download] finished\n", "[download] removed\n", "outside\n"} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in %q", want, got)
		}
	}
}

func TestWindowViewNavigationAndResize(t *testing.T) {
	logger := NewLogger(NewLoggerInput{Facts: terminalParameters.Facts{TerminalWidth: 80, TerminalHeight: 12}})
	logger.output = &bytes.Buffer{}
	window, err := logger.StartWindow()
	if err != nil {
		t.Fatal(err)
	}
	window.interactive = true
	defer func() {
		if err := window.Close(); err != nil {
			t.Error(err)
		}
	}()
	panels := make([]*WindowPanel, 5)
	for i := range panels {
		panels[i], err = window.AddPanel(fmt.Sprintf("task-%d", i))
		if err != nil {
			t.Fatal(err)
		}
	}
	for i := 0; i < 10; i++ {
		if err := panels[0].Logger().Print(fmt.Sprintf("line-%d", i)); err != nil {
			t.Fatal(err)
		}
	}
	if err := panels[0].Logger().Loader(10, 5); err != nil {
		t.Fatal(err)
	}
	if err := panels[0].Logger().Print("after progress"); err != nil {
		t.Fatal(err)
	}
	if err := panels[1].Logger().Print("a\tb"); err != nil {
		t.Fatal(err)
	}
	model := &windowModel{window: window}
	view := model.View()
	if !strings.Contains(view, "task-4") || !strings.Contains(view, "after progress") {
		t.Fatalf("missing panel or latest line: %q", view)
	}
	if strings.Contains(view, "\t") {
		t.Fatalf("tab escaped the panel: %q", view)
	}
	model.Update(tea.KeyMsg{Type: tea.KeyUp})
	if panels[0].offset != 1 {
		t.Fatalf("scroll offset = %d, want 1", panels[0].offset)
	}
	model.View()
	if err := panels[0].Logger().Print(strings.Repeat("x", 80)); err != nil {
		t.Fatal(err)
	}
	model.View()
	if panels[0].offset < 4 {
		t.Fatalf("new wrapped lines displaced scroll position: %d", panels[0].offset)
	}
	model.Update(tea.KeyMsg{Type: tea.KeyTab})
	if window.focused != 1 {
		t.Fatalf("focused panel = %d, want 1", window.focused)
	}
	model.Update(tea.WindowSizeMsg{Width: 30, Height: 6})
	view = model.View()
	if !strings.Contains(view, "task-1") {
		t.Fatalf("focused panel not visible after resize: %q", view)
	}
	for _, line := range strings.Split(view, "\n") {
		if ansi.StringWidth(line) > 30 {
			t.Fatalf("line exceeds terminal width: %q", line)
		}
	}
}

func TestWindowConcurrentPanels(t *testing.T) {
	logger := NewLogger(NewLoggerInput{})
	logger.output = &bytes.Buffer{}
	window, err := logger.StartWindow()
	if err != nil {
		t.Fatal(err)
	}
	var workers sync.WaitGroup
	errs := make(chan error, 5)
	for i := 0; i < 5; i++ {
		panel, err := window.AddPanel(fmt.Sprintf("task-%d", i))
		if err != nil {
			t.Fatal(err)
		}
		workers.Add(1)
		go func() {
			defer workers.Done()
			for j := 0; j < 5; j++ {
				if err := panel.Logger().Print("working"); err != nil {
					errs <- err
					return
				}
			}
			errs <- panel.Finish()
		}()
	}
	workers.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}
	if err := window.Close(); err != nil {
		t.Fatal(err)
	}
}

type failingWriter struct{}

func (failingWriter) Write([]byte) (int, error) { return 0, errors.New("write failed") }

func TestWindowFallbackWriteFailure(t *testing.T) {
	logger := NewLogger(NewLoggerInput{})
	logger.output = failingWriter{}
	window, err := logger.StartWindow()
	if err != nil {
		t.Fatal(err)
	}

	if _, err := window.AddPanel("task"); err == nil {
		t.Fatal("expected write error")
	}
	if err := window.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestWindowReportsRendererFailure(t *testing.T) {
	logger := NewLogger(NewLoggerInput{})
	logger.output = &bytes.Buffer{}
	window, err := logger.StartWindow()
	if err != nil {
		t.Fatal(err)
	}
	output := &windowOutput{writer: failingWriter{}}
	window.output = output
	if _, err := output.Write([]byte("frame")); err == nil {
		t.Fatal("expected renderer error")
	}
	if _, err := window.AddPanel("task"); err == nil {
		t.Fatal("renderer error was not reported")
	}
	if err := window.Close(); err == nil {
		t.Fatal("close did not report renderer error")
	}
}

func TestWindowFallbackWidthAndPanelComponents(t *testing.T) {
	var output bytes.Buffer
	width := 16
	logger := NewLogger(NewLoggerInput{MaxWidth: &width})
	logger.output = &output
	window, err := logger.StartWindow()
	if err != nil {
		t.Fatal(err)
	}
	panel, err := window.AddPanel("task")
	if err != nil {
		t.Fatal(err)
	}
	finish, err := panel.Logger().StartSection("setup", SeverityInfo, DashedSection)
	if err != nil {
		t.Fatal(err)
	}
	if err := panel.Logger().Command("echo hello world"); err != nil {
		t.Fatal(err)
	}
	if err := finish(); err != nil {
		t.Fatal(err)
	}
	if err := panel.Finish(); err != nil {
		t.Fatal(err)
	}
	if err := window.Close(); err != nil {
		t.Fatal(err)
	}
	for _, line := range strings.Split(strings.TrimSuffix(output.String(), "\n"), "\n") {
		if ansi.StringWidth(line) > width {
			t.Errorf("stream line exceeds width: %q", line)
		}
	}
}

func TestCommandFillsAvailableWidth(t *testing.T) {
	tests := []struct {
		name        string
		width       int
		maxWidth    *int
		ansiEnabled bool
		unicode     terminalParameters.Capability
		wantWidth   int
	}{
		{name: "terminal width", width: 60, wantWidth: 60},
		{name: "maximum width", width: 60, maxWidth: intPointer(40), wantWidth: 40},
		{name: "terminal narrower than maximum", width: 40, maxWidth: intPointer(60), wantWidth: 40},
		{name: "fallback width", wantWidth: 80},
		{name: "colored unicode", width: 35, ansiEnabled: true, unicode: terminalParameters.CapabilitySupported, wantWidth: 35},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var output bytes.Buffer
			logger := NewLogger(NewLoggerInput{
				Facts: terminalParameters.Facts{
					TerminalWidth: tt.width, SupportsANSI: tt.ansiEnabled, SupportsUnicode: tt.unicode,
				},
				Theme: theme.Theme{Border: "\x1b[34m", Code: "\x1b[32m", Reset: "\x1b[0m"},
				MaxWidth: tt.maxWidth,
			})
			logger.output = &output
			if err := logger.Command("echo hi"); err != nil {
				t.Fatal(err)
			}
			lines := strings.Split(strings.TrimSuffix(output.String(), "\n"), "\n")
			if len(lines) != 3 || ansi.Strip(lines[1]) != "echo hi" {
				t.Fatalf("unexpected command output: %q", output.String())
			}
			for _, line := range []string{lines[0], lines[2]} {
				if got := ansi.StringWidth(line); got != tt.wantWidth {
					t.Errorf("separator width = %d, want %d", got, tt.wantWidth)
				}
			}
		})
	}
}

func intPointer(value int) *int { return &value }

func TestWindowPanelCommandFillsPanelWidth(t *testing.T) {
	logger := NewLogger(NewLoggerInput{Facts: terminalParameters.Facts{TerminalWidth: 60, TerminalHeight: 12}})
	logger.output = &bytes.Buffer{}
	window, err := logger.StartWindow()
	if err != nil {
		t.Fatal(err)
	}
	window.interactive = true
	defer func() {
		if err := window.Close(); err != nil {
			t.Error(err)
		}
	}()
	panel, err := window.AddPanel("task")
	if err != nil {
		t.Fatal(err)
	}
	if err := panel.Logger().Command("echo hi"); err != nil {
		t.Fatal(err)
	}
	if len(panel.lines) != 3 || ansi.StringWidth(panel.lines[0]) != 58 || ansi.StringWidth(panel.lines[2]) != 58 {
		t.Fatalf("unexpected panel command separators: %#v", panel.lines)
	}
}

func TestWindowStyledHistory(t *testing.T) {
	var history []string
	var style string
	history, style = appendStyledHistory(history, "\x1b[31mhello", style, 3)
	history, style = appendStyledHistory(history, "world\x1b[0m", style, 3)
	if len(history) != 4 || !strings.HasPrefix(history[1], "\x1b[31m") ||
		!strings.HasPrefix(history[2], "\x1b[31m") || style != "" {
		t.Fatalf("style lost across wrapped or multiline text: %#v, %q", history, style)
	}
}

func TestWindowInteractiveTerminal(t *testing.T) {
	if os.Getenv("TRUSTPKG_TEST_TTY") != "1" {
		t.Skip("requires a real terminal")
	}
	if !term.IsTerminal(int(os.Stdin.Fd())) || !term.IsTerminal(int(os.Stdout.Fd())) {
		t.Fatal("test requires terminal input and output")
	}
	state, err := term.GetState(int(os.Stdin.Fd()))
	if err != nil {
		t.Fatal(err)
	}
	width, height, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		t.Fatal(err)
	}
	logger := NewLogger(NewLoggerInput{Facts: terminalParameters.Facts{
		IsStdinTTY: true, IsStdoutTTY: true, SupportsANSI: true,
		TerminalWidth: width, TerminalHeight: height,
	}})
	window, err := logger.StartWindow()
	if err != nil {
		t.Fatal(err)
	}
	closed := false
	defer func() {
		if !closed {
			if err := window.Close(); err != nil {
				t.Error(err)
			}
		}
	}()
	for i := 0; i < 5; i++ {
		panel, err := window.AddPanel(fmt.Sprintf("task-%d", i))
		if err != nil {
			t.Fatal(err)
		}
		if err := panel.Logger().Print("working"); err != nil {
			t.Fatal(err)
		}
		if err := panel.Finish(); err != nil {
			t.Fatal(err)
		}
	}
	time.Sleep(300 * time.Millisecond)
	if os.Getenv("TRUSTPKG_TEST_KEYS") == "1" {
		window.mu.Lock()
		width, focused := window.width, window.focused
		window.mu.Unlock()
		if width != 60 || focused != 1 {
			t.Errorf("resize or keyboard input not handled: width=%d focused=%d", width, focused)
		}
	}
	if err := window.Close(); err != nil {
		t.Fatal(err)
	}
	closed = true
	after, err := term.GetState(int(os.Stdin.Fd()))
	if err != nil {
		t.Fatal(err)
	}
	if *state != *after {
		t.Fatal("terminal input mode was not restored")
	}
}

func TestWindowInteractiveInterrupt(t *testing.T) {
	if os.Getenv("TRUSTPKG_TEST_TTY") != "1" {
		t.Skip("requires a real terminal")
	}
	state, err := term.GetState(int(os.Stdin.Fd()))
	if err != nil {
		t.Fatal(err)
	}
	width, height, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		t.Fatal(err)
	}
	logger := NewLogger(NewLoggerInput{Facts: terminalParameters.Facts{
		IsStdinTTY: true, IsStdoutTTY: true, SupportsANSI: true,
		TerminalWidth: width, TerminalHeight: height,
	}})
	window, err := logger.StartWindow()
	if err != nil {
		t.Fatal(err)
	}
	closed := false
	defer func() {
		if !closed {
			_ = window.Close()
		}
	}()
	panel, err := window.AddPanel("interrupt-ready")
	if err != nil {
		t.Fatal(err)
	}
	select {
	case <-window.done:
	case <-time.After(5 * time.Second):
		t.Fatal("window did not respond to Ctrl+C")
	}
	if err := panel.Logger().Print("late"); err == nil {
		t.Error("interrupted window accepted a log")
	}
	if err := window.Close(); err == nil {
		t.Error("interrupt was not reported")
	}
	closed = true
	after, err := term.GetState(int(os.Stdin.Fd()))
	if err != nil {
		t.Fatal(err)
	}
	if *state != *after {
		t.Fatal("terminal input mode was not restored")
	}
}
