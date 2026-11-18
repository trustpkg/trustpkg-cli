package logs

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	terminalParameters "github.com/trustpkg/trustpkg-cli-forge/internal/terminal-parameters"
)

type InputConfig struct {
	Title        string
	Placeholder  string
	InitialValue string
	Required     bool
	Validate     func(string) error
}

type inputModel struct {
	promptBase
	field    textinput.Model
	required bool
	validate func(string) error
}

func (logger *Logger) Input(input InputConfig) (string, error) {
	if logger == nil || logger.output == nil {
		return "", fmt.Errorf("input requires an initialized logger")
	}
	if err := validatePromptText("input title", input.Title, true); err != nil {
		return "", err
	}
	if err := validatePromptText("input placeholder", input.Placeholder, false); err != nil {
		return "", err
	}
	if err := validatePromptText("initial input", input.InitialValue, false); err != nil {
		return "", err
	}
	field := textinput.New()
	field.Prompt = ""
	field.Placeholder = input.Placeholder
	field.SetValue(input.InitialValue)
	model := &inputModel{
		promptBase: promptBase{logger: logger, title: input.Title, width: defaultLoaderWidth, height: 24},
		field:      field,
		required:   input.Required,
		validate:   input.Validate,
	}
	model.resize(model.width, model.height)
	if err := logger.runPrompt(model); err != nil {
		return "", err
	}
	if !model.submitted || model.cancelled {
		return "", ErrPromptCancelled
	}
	return model.field.Value(), nil
}

func (model *inputModel) Init() tea.Cmd {
	return model.field.Focus()
}

func (model *inputModel) resize(width, height int) {
	model.promptBase.resize(width, height)
	model.field.Width = max(1, model.width-4)
}

func (model *inputModel) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	switch message := message.(type) {
	case tea.WindowSizeMsg:
		model.resize(message.Width, message.Height)
	case tea.KeyMsg:
		switch message.String() {
		case "esc", "ctrl+c":
			model.cancelled = true
			return model, tea.Quit
		}
		if model.small() {
			return model, nil
		}
		if message.String() == "enter" {
			if model.field.Err != nil {
				model.message = model.field.Err.Error()
				if model.message == "" {
					model.message = "Input error."
				}
				return model, nil
			}
			value := model.field.Value()
			if model.required && strings.TrimSpace(value) == "" {
				model.message = "Enter a value."
				return model, nil
			}
			if model.validate != nil {
				if err := model.validate(value); err != nil {
					model.message = err.Error()
					if model.message == "" {
						model.message = "Invalid value."
					}
					return model, nil
				}
			}
			model.submitted = true
			return model, tea.Quit
		}
		model.message = ""
	}
	var cmd tea.Cmd
	model.field, cmd = model.field.Update(message)
	if model.field.Err != nil {
		model.message = model.field.Err.Error()
		if model.message == "" {
			model.message = "Input error."
		}
	}
	return model, cmd
}

func (model *inputModel) View() string {
	if model.cancelled {
		return model.line(model.color("Cancelled", model.logger.theme.Warning)) + "\n"
	}
	if model.submitted {
		marker := "+ "
		if model.logger.terminalFacts.SupportsUnicode == terminalParameters.CapabilitySupported {
			marker = "✓ "
		}
		value := model.field.Value()
		if value == "" {
			value = "(skipped)"
		}
		return model.line(model.color(marker, model.logger.theme.Success)+model.title+": "+
			model.color(value, model.logger.theme.Value)) + "\n"
	}
	if model.small() {
		return model.line("Terminal too small; resize or press Esc.")
	}
	lines := []string{
		model.line(model.color("? ", model.logger.theme.Prompt) + model.color(model.title, model.logger.theme.Heading)),
		model.line(model.color("  > ", model.logger.theme.Accent) + model.color(model.field.View(), model.logger.theme.Input)),
		model.line(model.color("Enter confirm  Esc cancel", model.logger.theme.Description)),
	}
	if model.message != "" {
		lines = append(lines, model.warning())
	}
	return strings.Join(lines, "\n")
}
