package terminalParameters

const (
	ThemeDark  Theme = "dark"
	ThemeLight Theme = "light"
	ThemeAuto  Theme = "auto"
)

const (
	CapabilitySupported   Capability = "supported"
	CapabilityUnsupported Capability = "unsupported"
	CapabilityUnknown     Capability = "unknown"
)

const gatherTotalSteps = 13

const (
	envColorForegroundBackground = "COLORFGBG"
	envLocaleAll                 = "LC_ALL"
	envLocaleCharacterType       = "LC_CTYPE"
	envLanguage                  = "LANG"
	envWindowsTerminalSession    = "WT_SESSION"
	envKittyWindowID             = "KITTY_WINDOW_ID"
	envWezTermPane               = "WEZTERM_PANE"
	envTerminalProgram           = "TERM_PROGRAM"
	envTerminalType              = "TERM"
	envColorTerminal             = "COLORTERM"
)

const (
	colorForegroundBackgroundSeparator = ";"
	darkBackgroundMin                  = 0
	darkBackgroundMax                  = 6
	lightBackgroundMin                 = 7

	localeUTF8Hyphenated = "utf-8"
	localeUTF8Compact    = "utf8"
	localeC              = "c"
	localePOSIX          = "posix"
	localeCPrefix        = "c."

	terminalProgramITerm   = "iterm.app"
	terminalProgramWezTerm = "wezterm"
	terminalProgramVSCode  = "vscode"
	terminalProgramGhostty = "ghostty"
	terminalProgramHyper   = "hyper"

	terminalTypeDumb       = "dumb"
	terminalType256Color   = "256color"
	terminalTypeDirect     = "direct"
	terminalTrueColor      = "truecolor"
	colorTerminal24Bit     = "24bit"
)
