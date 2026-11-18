package theme

import (
	"encoding/hex"
	"fmt"
	"reflect"

	terminalParameters "github.com/trustpkg/trustpkg-cli/internal/terminal-parameters"
)

var ansiColorNames = [...]string{
	"black", "red", "green", "yellow", "blue", "magenta", "cyan", "white",
	"bright_black", "bright_red", "bright_green", "bright_yellow",
	"bright_blue", "bright_magenta", "bright_cyan", "bright_white",
}

func ansiColorIndex(name string) (int, bool) {
	for i, color := range ansiColorNames {
		if name == color {
			return i, true
		}
	}

	return 0, false
}

func gatherThemeFromValidatedTomlFile(input gatherThemeFromTomlFileInput) (Theme, error) {
	document, err := input.readValidatedThemeFile()
	if err != nil {
		return Theme{}, err
	}

	facts := input.terminalFact
	if !facts.SupportsANSI {
		input.reportStep()
		return NoColorTheme(), nil
	}

	var appearance string
	switch facts.TerminalTheme {
	case terminalParameters.ThemeLight:
		appearance = "light"
	case terminalParameters.ThemeDark:
		appearance = "dark"
	default:
		return Theme{}, fmt.Errorf("cannot select theme for terminal appearance %q", facts.TerminalTheme)
	}

	full := facts.SupportsTrueColor
	var section string

	switch {
	case full:
		section = "full colors " + appearance
	case facts.SupportsANSI16Colors:
		section = "limited colors " + appearance
	default:
		return Theme{}, fmt.Errorf("terminal does not report support for 16 colors or true color")
	}

	entries, ok := document[section].(map[string]any)
	if !ok {
		return Theme{}, fmt.Errorf("invalid theme section %q", section)
	}

	theme := Theme{Reset: reset}
	fields := reflect.ValueOf(&theme).Elem()

	for i := 0; i < fields.NumField(); i++ {
		field := fields.Type().Field(i)
		name := field.Tag.Get("toml")
		if name == "" || name == "-" {
			continue
		}
		color, ok := entries[name].(string)
		if !ok {
			return Theme{}, fmt.Errorf("invalid color %q in theme section %q", name, section)
		}
		sequence, err := terminalColor(color, full, name == "background")
		if err != nil {
			return Theme{}, fmt.Errorf("invalid color %q in theme section %q: %w", name, section, err)
		}
		fields.Field(i).SetString(sequence)
	}

	input.reportStep()
	return theme, nil
}

func terminalColor(color string, full, background bool) (string, error) {
	if background && color == "default" {
		return NoColor, nil
	}
	if full {
		if len(color) != 7 || color[0] != '#' {
			return "", fmt.Errorf("expected #RRGGBB")
		}

		rgb, err := hex.DecodeString(color[1:])
		if err != nil {
			return "", fmt.Errorf("invalid RGB color: %w", err)
		}

		if background {
			return RGBBackground(rgb[0], rgb[1], rgb[2]), nil
		}

		return RGBForeground(rgb[0], rgb[1], rgb[2]), nil
	}

	index, ok := ansiColorIndex(color)
	if !ok {
		return "", fmt.Errorf("unknown ANSI color %q", color)
	}

	code := 30
	if background {
		code = 40
	}

	if index >= 8 {
		code += 60
		index -= 8
	}

	return fmt.Sprintf("\x1b[%dm", code+index), nil
}

func gatherThemeFromTomlFile(input gatherThemeFromTomlFileInput) (Theme, error) {
	return gatherThemeFromValidatedTomlFile(input)
}
