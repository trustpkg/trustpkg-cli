package theme

import (
	"github.com/trustpkg/trustpkg-cli/internal/progress"
	terminalParameters "github.com/trustpkg/trustpkg-cli/internal/terminal-parameters"
)

type Handler struct {
	theme Theme
}

type ThemeInitInput struct {
	ThemeFilePath *string
	TerminalFacts terminalParameters.Facts
	OnProgress    progress.Callback
}

func Init(input ThemeInitInput) (*Handler, error) {
	value, err := gatherTheme(GatherThemeInput{
		themeFilePath: input.ThemeFilePath,
		terminalFacts: input.TerminalFacts,
		onProgress:    input.OnProgress,
	})
	if err != nil {
		return nil, err
	}
	return &Handler{theme: value}, nil
}

func (handler *Handler) Theme() Theme {
	return handler.theme
}

func (handler *Handler) generateThemeFile(themeDirPath string, onProgress progress.Callback) error {
	return GenerateThemeFile(themeDirPath, onProgress)
}
