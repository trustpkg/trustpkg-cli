package terminalParameters

import (
	"os"
	"strconv"
	"strings"
)

type environmentChecker interface {
	theme() Theme
	unicode() Capability
	hyperlinks() Capability
	supportsANSI(isTTY bool) bool
	supportsANSI256Colors(supportsANSI bool) bool
	supportsTrueColor(supportsANSI bool) bool
}

type processEnvironmentChecker struct{}

var _ environmentChecker = processEnvironmentChecker{}

func (processEnvironmentChecker) theme() Theme {
	parts := strings.Split(os.Getenv(envColorForegroundBackground), colorForegroundBackgroundSeparator)
	background, err := strconv.Atoi(parts[len(parts)-1])
	if err != nil {
		return ThemeAuto
	}
	if background >= darkBackgroundMin && background <= darkBackgroundMax {
		return ThemeDark
	}
	if background >= lightBackgroundMin {
		return ThemeLight
	}
	return ThemeAuto
}

func (processEnvironmentChecker) unicode() Capability {
	locale := os.Getenv(envLocaleAll)
	if locale == "" {
		locale = os.Getenv(envLocaleCharacterType)
	}
	if locale == "" {
		locale = os.Getenv(envLanguage)
	}

	locale = strings.ToLower(locale)
	if strings.Contains(locale, localeUTF8Hyphenated) || strings.Contains(locale, localeUTF8Compact) {
		return CapabilitySupported
	}
	if locale == localeC || locale == localePOSIX || strings.HasPrefix(locale, localeCPrefix) {
		return CapabilityUnsupported
	}
	return CapabilityUnknown
}

func (processEnvironmentChecker) hyperlinks() Capability {
	if os.Getenv(envWindowsTerminalSession) != "" ||
		os.Getenv(envKittyWindowID) != "" ||
		os.Getenv(envWezTermPane) != "" {
		return CapabilitySupported
	}

	switch strings.ToLower(os.Getenv(envTerminalProgram)) {
	case terminalProgramITerm, terminalProgramWezTerm, terminalProgramVSCode, terminalProgramGhostty, terminalProgramHyper:
		return CapabilitySupported
	default:
		return CapabilityUnknown
	}
}

func (processEnvironmentChecker) supportsANSI(isTTY bool) bool {
	return isTTY && !strings.EqualFold(os.Getenv(envTerminalType), terminalTypeDumb)
}

func (processEnvironmentChecker) supportsANSI256Colors(supportsANSI bool) bool {
	termType := strings.ToLower(os.Getenv(envTerminalType))
	colorTerm := strings.ToLower(os.Getenv(envColorTerminal))

	return supportsANSI &&
		(strings.Contains(termType, terminalType256Color) || colorTerm == terminalTrueColor || colorTerm == colorTerminal24Bit)
}

func (processEnvironmentChecker) supportsTrueColor(supportsANSI bool) bool {
	termType := strings.ToLower(os.Getenv(envTerminalType))
	colorTerm := strings.ToLower(os.Getenv(envColorTerminal))

	return supportsANSI &&
		(colorTerm == terminalTrueColor || colorTerm == colorTerminal24Bit ||
			strings.Contains(termType, terminalTypeDirect) || strings.Contains(termType, terminalTrueColor))
}
