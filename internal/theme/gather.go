package theme

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/trustpkg/trustpkg-cli-forge/internal/progress"
	terminalParameters "github.com/trustpkg/trustpkg-cli-forge/internal/terminal-parameters"
)

func gatherTheme(input GatherThemeInput) (Theme, error) {
	var path string
	if input.themeFilePath != nil {
		path = *input.themeFilePath
	} else {
		dir, err := os.Getwd()
		if err != nil {
			return Theme{}, fmt.Errorf("get working directory for theme: %w", err)
		}
		path = filepath.Join(dir, "trustpkg-cli-theme.toml")
		if _, err := os.Lstat(path); err != nil {
			if errors.Is(err, os.ErrNotExist) {
				report := progress.NewReporter(2, input.onProgress)
				report()
				theme := gatherDefaultTheme(input.terminalFacts)
				report()
				return theme, nil
			}
			return Theme{}, fmt.Errorf("check theme file %q: %w", path, err)
		}
	}

	report := progress.NewReporter(9, input.onProgress)
	report()
	return gatherThemeFromTomlFile(gatherThemeFromTomlFileInput{
		themeFilePath: path,
		terminalFact:  input.terminalFacts,
		report:        report,
	})
}

func gatherDefaultTheme(facts terminalParameters.Facts) Theme {
	if !facts.SupportsANSI || (!facts.SupportsTrueColor && !facts.SupportsANSI16Colors) {
		return NoColorTheme()
	}

	switch facts.TerminalTheme {
	case terminalParameters.ThemeLight, terminalParameters.ThemeAuto:
		if facts.SupportsTrueColor {
			return DefaultFullColorLightTheme()
		}
		return DefaultLimitedSixteenColorLightTheme()
	case terminalParameters.ThemeDark:
		if facts.SupportsTrueColor {
			return DefaultFullColorDarkTheme()
		}
		return DefaultLimitedSixteenColorDarkTheme()
	default:
		panic("unsupported terminal theme: " + string(facts.TerminalTheme))
	}
}
