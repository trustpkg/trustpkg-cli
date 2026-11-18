package logs

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"
)

func TestSelectKeyboardAndRequired(t *testing.T) {
	logger := NewLogger(NewLoggerInput{})
	model := &selectModel{
		promptBase: promptBase{logger: logger, title: "Choose", width: 30, height: 6},
		options:    make([]SelectOption, 10),
		selected:   make([]bool, 10),
		required:   true,
	}
	for i := range model.options {
		model.options[i] = SelectOption{Label: fmt.Sprintf("Option %d", i), Value: fmt.Sprint(i)}
	}
	model.Update(tea.KeyMsg{Type: tea.KeyCtrlD})
	if model.submitted || model.message == "" {
		t.Fatal("empty required selection was accepted")
	}
	model.Update(tea.KeyMsg{Type: tea.KeyTab})
	model.Update(tea.KeyMsg{Type: tea.KeyTab})
	if !model.selected[0] || model.submitted {
		t.Fatal("Tab should select without submitting or toggling")
	}
	model.Update(tea.KeyMsg{Type: tea.KeySpace})
	if model.selected[0] {
		t.Fatal("Space should toggle the selection")
	}
	model.Update(tea.KeyMsg{Type: tea.KeySpace})
	model.Update(tea.KeyMsg{Type: tea.KeyEnd})
	view := model.View()
	if !strings.Contains(view, "Option 9") || strings.Contains(view, "Option 0") {
		t.Fatalf("focused option is not visible: %q", view)
	}
	for _, line := range strings.Split(view, "\n") {
		if ansi.StringWidth(line) > 30 {
			t.Fatalf("line exceeds width: %q", line)
		}
	}
	model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if !model.submitted || model.count() != 2 || !model.selected[0] || !model.selected[9] {
		t.Fatalf("unexpected selection: %#v", model.selected)
	}
}

func TestSelectEnterOnSelectedAndOptionalEmpty(t *testing.T) {
	logger := NewLogger(NewLoggerInput{})
	model := &selectModel{
		promptBase: promptBase{logger: logger, width: 30, height: 8},
		options:    []SelectOption{{Label: "One", Value: "one"}},
		selected:   []bool{true},
	}
	model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if !model.submitted || !model.selected[0] {
		t.Fatal("Enter should keep the selected option and submit")
	}
	model = &selectModel{
		promptBase: promptBase{logger: logger, width: 30, height: 8},
		options:    []SelectOption{{Label: "One", Value: "one"}},
		selected:   []bool{false},
	}
	model.Update(tea.KeyMsg{Type: tea.KeyCtrlD})
	if !model.submitted || model.count() != 0 {
		t.Fatal("Ctrl+D should allow an empty optional selection")
	}
}

func TestSelectCancellationAndValidation(t *testing.T) {
	logger := NewLogger(NewLoggerInput{})
	var output bytes.Buffer
	logger.output = &output
	_, err := logger.Select(SelectInput{Title: "Choose", Options: []SelectOption{
		{Label: "One", Value: "same"},
		{Label: "Two", Value: "same"},
	}})
	if err == nil || output.Len() != 0 {
		t.Fatalf("expected duplicate value error and no output, got %v", err)
	}
	_, err = logger.Select(SelectInput{Title: "Choose", Options: []SelectOption{{Label: "One", Value: "one"}}})
	if !errors.Is(err, ErrPromptUnsupported) || output.Len() != 0 {
		t.Fatalf("expected noninteractive error and no output, got %v", err)
	}
	model := &selectModel{promptBase: promptBase{logger: logger, width: 30, height: 10}}
	model.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if !model.cancelled || model.submitted {
		t.Fatal("escape did not cancel")
	}
}

func TestInputEditingAndValidation(t *testing.T) {
	logger := NewLogger(NewLoggerInput{})
	field := textinput.New()
	field.Prompt = ""
	model := &inputModel{
		promptBase: promptBase{logger: logger, title: "Package", width: 30, height: 8},
		field:      field,
		required:   true,
		validate: func(value string) error {
			if value == "bad" {
				return errors.New("invalid package")
			}
			return nil
		},
	}
	model.Init()
	model.resize(30, 8)
	model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if model.submitted || model.message == "" {
		t.Fatal("empty required input was accepted")
	}
	model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("bad")})
	model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if model.submitted || model.message != "invalid package" {
		t.Fatalf("validation error not shown: %q", model.message)
	}
	model.Update(tea.KeyMsg{Type: tea.KeyCtrlU})
	model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("curl")})
	model.Update(tea.KeyMsg{Type: tea.KeyLeft})
	model.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if !model.submitted || model.field.Value() != "cul" {
		t.Fatalf("input editing failed: %q", model.field.Value())
	}
}

func TestInputNoninteractiveAndOptional(t *testing.T) {
	logger := NewLogger(NewLoggerInput{})
	var output bytes.Buffer
	logger.output = &output
	if _, err := logger.Input(InputConfig{Title: "Package"}); !errors.Is(err, ErrPromptUnsupported) || output.Len() != 0 {
		t.Fatalf("expected noninteractive error and no output, got %v", err)
	}
	model := &inputModel{
		promptBase: promptBase{logger: logger, width: 30, height: 8},
		field:      textinput.New(),
	}
	model.Init()
	model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if !model.submitted || model.field.Value() != "" {
		t.Fatal("optional empty input was not accepted")
	}
}
