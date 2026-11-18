package trustpkgCli

import (
	"github.com/trustpkg/trustpkg-cli/internal/progress"
	terminalParameters "github.com/trustpkg/trustpkg-cli/internal/terminal-parameters"
)

type (
	Facts            = terminalParameters.Facts
	TerminalTheme    = terminalParameters.Theme
	Capability       = terminalParameters.Capability
	ProgressUpdate   = progress.Update
	ProgressCallback = progress.Callback
)

const (
	ThemeDark             = terminalParameters.ThemeDark
	ThemeLight            = terminalParameters.ThemeLight
	ThemeAuto             = terminalParameters.ThemeAuto
	CapabilitySupported   = terminalParameters.CapabilitySupported
	CapabilityUnsupported = terminalParameters.CapabilityUnsupported
	CapabilityUnknown     = terminalParameters.CapabilityUnknown
)

func GatherFacts(onProgress ProgressCallback) (Facts, error) {
	return terminalParameters.GatherFacts(onProgress)
}

func NewProgressReporter(totalSteps int, onProgress ProgressCallback) func() {
	return progress.NewReporter(totalSteps, onProgress)
}
