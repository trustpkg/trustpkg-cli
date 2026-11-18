package theme

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/pelletier/go-toml/v2"
	"github.com/trustpkg/trustpkg-cli-forge/internal/progress"
)

func GenerateThemeFile(themeDirPath string, onProgress progress.Callback) error {
	if strings.TrimSpace(themeDirPath) == "" {
		return fmt.Errorf("theme directory path cannot be empty")
	}
	report := progress.NewReporter(7, onProgress)
	report()

	sections := []struct {
		name  string
		theme Theme
		full  bool
	}{
		{"full colors light", DefaultFullColorLightTheme(), true},
		{"full colors dark", DefaultFullColorDarkTheme(), true},
		{"limited colors light", DefaultLimitedSixteenColorLightTheme(), false},
		{"limited colors dark", DefaultLimitedSixteenColorDarkTheme(), false},
	}

	document := make(map[string]map[string]string, len(sections))
	for _, section := range sections {
		entries := make(map[string]string)
		fields := reflect.ValueOf(section.theme)
		for i := 0; i < fields.NumField(); i++ {
			field := fields.Type().Field(i)
			name := field.Tag.Get("toml")
			if name == "" || name == "-" {
				continue
			}
			color, err := themeColorValue(fields.Field(i).String(), section.full, name == "background")
			if err != nil {
				return fmt.Errorf("encode %q in theme section %q: %w", name, section.name, err)
			}
			entries[name] = color
		}
		document[section.name] = entries
		report()
	}

	data, err := toml.Marshal(document)
	if err != nil {
		return fmt.Errorf("encode default theme: %w", err)
	}
	report()

	path := filepath.Join(themeDirPath, "theme.toml")
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	if err != nil {
		return fmt.Errorf("create theme file %q: %w", path, err)
	}

	_, writeErr := file.Write(data)
	closeErr := file.Close()
	if err := errors.Join(writeErr, closeErr); err != nil {
		return fmt.Errorf("write theme file %q (file may be incomplete): %w", path, err)
	}

	report()
	return nil
}

func themeColorValue(sequence string, full, background bool) (string, error) {
	if background && sequence == NoColor {
		return "default", nil
	}

	if full {
		var red, green, blue uint8
		pattern := "\x1b[38;2;%d;%d;%dm"
		if background {
			pattern = "\x1b[48;2;%d;%d;%dm"
		}

		n, err := fmt.Sscanf(sequence, pattern, &red, &green, &blue)
		expected := RGBForeground(red, green, blue)
		if background {
			expected = RGBBackground(red, green, blue)
		}

		if err != nil || n != 3 || sequence != expected {
			return "", fmt.Errorf("unsupported RGB color sequence %q", sequence)
		}

		return fmt.Sprintf("#%02X%02X%02X", red, green, blue), nil
	}

	for _, name := range ansiColorNames {
		color, err := terminalColor(name, false, background)
		if err != nil {
			return "", err
		}

		if sequence == color {
			return name, nil
		}
	}

	return "", fmt.Errorf("unsupported ANSI color sequence %q", sequence)
}
