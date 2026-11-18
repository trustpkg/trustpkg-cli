package facts

import (
	"github.com/trustpkg/trustpkg-cli-forge/internal/progress"
	terminalParameters "github.com/trustpkg/trustpkg-cli-forge/internal/terminal-parameters"
)

type (
	Facts         = terminalParameters.Facts
	TerminalTheme = terminalParameters.Theme
	Capability    = terminalParameters.Capability
)

const (
	ThemeDark             = terminalParameters.ThemeDark
	ThemeLight            = terminalParameters.ThemeLight
	ThemeAuto             = terminalParameters.ThemeAuto
	CapabilitySupported   = terminalParameters.CapabilitySupported
	CapabilityUnsupported = terminalParameters.CapabilityUnsupported
	CapabilityUnknown     = terminalParameters.CapabilityUnknown
)

func GatherTerminalFacts(onProgress progress.Callback) (Facts, error) {
	return terminalParameters.GatherFacts(onProgress)
}
