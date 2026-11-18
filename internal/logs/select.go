package logs

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	terminalParameters "github.com/trustpkg/trustpkg-cli-forge/internal/terminal-parameters"
)

type SelectOption struct {
	Label       string
	Value       string
	Description string
	Selected    bool
}

type SelectInput struct {
	Title    string
	Options  []SelectOption
	Required bool
}

type selectModel struct {
	promptBase
	options  []SelectOption
	selected []bool
	cursor   int
	required bool
}

func (logger *Logger) Select(input SelectInput) ([]string, error) {
	if logger == nil || logger.output == nil {
		return nil, fmt.Errorf("select requires an initialized logger")
	}
	if err := validatePromptText("select title", input.Title, true); err != nil {
		return nil, err
	}
	if len(input.Options) == 0 {
		return nil, fmt.Errorf("select requires at least one option")
	}
	values := make(map[string]bool, len(input.Options))
	model := &selectModel{
		promptBase: promptBase{logger: logger, title: input.Title, width: defaultLoaderWidth, height: 24},
		options:    append([]SelectOption(nil), input.Options...),
		selected:   make([]bool, len(input.Options)),
	}
	for i, option := range model.options {
		if err := validatePromptText("option label", option.Label, true); err != nil {
			return nil, err
		}
		if err := validatePromptText("option value", option.Value, true); err != nil {
			return nil, err
		}
		if err := validatePromptText("option description", option.Description, false); err != nil {
			return nil, err
		}
		if values[option.Value] {
			return nil, fmt.Errorf("duplicate select value %q", option.Value)
		}
		values[option.Value] = true
		model.selected[i] = option.Selected
	}
	model.required = input.Required
	model.resize(model.width, model.height)
	if err := logger.runPrompt(model); err != nil {
		return nil, err
	}
	if !model.submitted || model.cancelled {
		return nil, ErrPromptCancelled
	}
	result := make([]string, 0)
	for i, option := range model.options {
		if model.selected[i] {
			result = append(result, option.Value)
		}
	}
	return result, nil
}

func (model *selectModel) Init() tea.Cmd { return nil }

func (model *selectModel) Update(message tea.Msg) (tea.Model, tea.Cmd) {
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
		switch message.String() {
		case "up", "k":
			model.cursor = max(0, model.cursor-1)
		case "down", "j":
			model.cursor = min(len(model.options)-1, model.cursor+1)
		case "home":
			model.cursor = 0
		case "end":
			model.cursor = len(model.options) - 1
		case "pgup":
			model.cursor = max(0, model.cursor-max(1, model.height-5))
		case "pgdown":
			model.cursor = min(len(model.options)-1, model.cursor+max(1, model.height-5))
		case " ":
			model.selected[model.cursor] = !model.selected[model.cursor]
			model.message = ""
		case "tab":
			model.selected[model.cursor] = true
			model.message = ""
		case "ctrl+a":
			all := true
			for _, selected := range model.selected {
				all = all && selected
			}
			for i := range model.selected {
				model.selected[i] = !all
			}
			model.message = ""
		case "enter":
			model.selected[model.cursor] = true
			model.submitted = true
			return model, tea.Quit
		case "ctrl+d":
			if model.required && model.count() == 0 {
				model.message = "Select at least one option."
				return model, nil
			}
			model.submitted = true
			return model, tea.Quit
		}
	}
	return model, nil
}

func (model *selectModel) count() int {
	count := 0
	for _, selected := range model.selected {
		if selected {
			count++
		}
	}
	return count
}

func (model *selectModel) View() string {
	if model.cancelled {
		return model.line(model.color("Cancelled", model.logger.theme.Warning)) + "\n"
	}
	if model.submitted {
		marker := "+ "
		if model.logger.terminalFacts.SupportsUnicode == terminalParameters.CapabilitySupported {
			marker = "✓ "
		}
		return model.line(model.color(marker, model.logger.theme.Success)+model.title+
			fmt.Sprintf(" (%d selected)", model.count())) + "\n"
	}
	if model.small() {
		return model.line("Terminal too small; resize or press Esc.")
	}
	lines := []string{
		model.line(model.color("? ", model.logger.theme.Prompt) + model.color(model.title, model.logger.theme.Heading)),
		model.line(model.color("Tab select  Space toggle  Enter select & finish", model.logger.theme.Description)),
	}
	visible := max(1, model.height-5)
	start := max(0, min(model.cursor-visible+1, len(model.options)-visible))
	end := min(len(model.options), start+visible)
	for i := start; i < end; i++ {
		pointer, check := " ", "[ ]"
		if i == model.cursor {
			pointer = ">"
		}
		if model.selected[i] {
			check = "[x]"
		}
		if model.logger.terminalFacts.SupportsUnicode == terminalParameters.CapabilitySupported {
			if i == model.cursor {
				pointer = "›"
			}
			if model.selected[i] {
				check = "[✓]"
			}
		}
		color := model.logger.theme.Text
		if i == model.cursor || model.selected[i] {
			color = model.logger.theme.Selected
		}
		line := model.color(pointer+" "+check+" "+model.options[i].Label, color)
		if model.options[i].Description != "" {
			line += model.color("  "+model.options[i].Description, model.logger.theme.Muted)
		}
		lines = append(lines, model.line(line))
	}
	move := "Up/Down"
	if model.logger.terminalFacts.SupportsUnicode == terminalParameters.CapabilitySupported {
		move = "↑↓"
	}
	status := fmt.Sprintf("%s move  Ctrl+A all  Ctrl+D finish  Esc cancel  %d/%d", move,
		model.cursor+1, len(model.options))
	lines = append(lines, model.line(model.color(status, model.logger.theme.Description)))
	if model.message != "" {
		lines = append(lines, model.warning())
	}
	return strings.Join(lines, "\n")
}
