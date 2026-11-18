package theme

import (
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	terminalParameters "github.com/trustpkg/trustpkg-cli-forge/internal/terminal-parameters"

	"github.com/pelletier/go-toml/v2"
)

type gatherThemeFromTomlFileInput struct {
	terminalFact  terminalParameters.Facts
	themeFilePath string
	report        func()
}

func (input *gatherThemeFromTomlFileInput) validateFile() error {
	if input == nil || strings.TrimSpace(input.themeFilePath) == "" {
		return fmt.Errorf("theme file path cannot be empty")
	}

	path := input.themeFilePath
	name := filepath.Base(path)
	if filepath.Ext(name) != ".toml" || strings.TrimSpace(strings.TrimSuffix(name, ".toml")) == "" {
		return fmt.Errorf("theme file must have a non-empty name and a .toml extension: %s", path)
	}

	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("cannot stat theme file %q: %w", path, err)
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("theme path is not a regular file: %s", path)
	}

	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("cannot open theme file %q for reading: %w", path, err)
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("cannot close theme file %q: %w", path, err)
	}

	return nil
}

func (input *gatherThemeFromTomlFileInput) validateThemeFileContent() error {
	_, err := input.readValidatedThemeFile()
	return err
}

func (input *gatherThemeFromTomlFileInput) readValidatedThemeFile() (map[string]any, error) {
	if err := input.validateFile(); err != nil {
		return nil, err
	}
	input.reportStep()

	data, err := os.ReadFile(input.themeFilePath)
	if err != nil {
		return nil, fmt.Errorf("cannot read theme file %q: %w", input.themeFilePath, err)
	}
	input.reportStep()
	var document map[string]any
	if err := toml.Unmarshal(data, &document); err != nil {
		return nil, fmt.Errorf("invalid theme TOML in %q: %w", input.themeFilePath, err)
	}
	input.reportStep()
	if err := validateThemeDocument(document, input.report); err != nil {
		return nil, fmt.Errorf("invalid theme file %q: %w", input.themeFilePath, err)
	}
	return document, nil
}

func (input *gatherThemeFromTomlFileInput) reportStep() {
	if input.report != nil {
		input.report()
	}
}

func validateThemeDocument(document map[string]any, report func()) error {
	sections := []struct {
		name string
		full bool
	}{
		{"full colors light", true},
		{"full colors dark", true},
		{"limited colors light", false},
		{"limited colors dark", false},
	}
	allowedSections := make(map[string]bool, len(sections))
	for _, section := range sections {
		allowedSections[section.name] = true
	}
	for name := range document {
		if !allowedSections[name] {
			return fmt.Errorf("unknown theme section %q", name)
		}
	}

	themeType := reflect.TypeOf(Theme{})
	keys := make(map[string]bool, themeType.NumField())
	for field := range themeType.Fields() {
		name := field.Tag.Get("toml")
		if name != "" && name != "-" {
			keys[name] = true
		}
	}

	for _, section := range sections {
		value, exists := document[section.name]
		if !exists {
			return fmt.Errorf("missing theme section %q", section.name)
		}
		entries, ok := value.(map[string]any)
		if !ok {
			return fmt.Errorf("theme section %q must be a table", section.name)
		}
		for name := range entries {
			if !keys[name] {
				return fmt.Errorf("unknown color %q in theme section %q", name, section.name)
			}
		}
		for i := 0; i < themeType.NumField(); i++ {
			name := themeType.Field(i).Tag.Get("toml")
			if name == "" || name == "-" {
				continue
			}
			value, exists := entries[name]
			if !exists {
				return fmt.Errorf("missing color %q in theme section %q", name, section.name)
			}
			color, ok := value.(string)
			if !ok || !validThemeColor(color, section.full, name == "background") {
				return fmt.Errorf("invalid color for %q in theme section %q", name, section.name)
			}
		}
		if report != nil {
			report()
		}
	}
	return nil
}

func validThemeColor(color string, full, background bool) bool {
	if background && color == "default" {
		return true
	}
	if full {
		if len(color) != 7 || color[0] != '#' {
			return false
		}
		_, err := hex.DecodeString(color[1:])
		return err == nil
	}
	_, ok := ansiColorIndex(color)
	return ok
}
