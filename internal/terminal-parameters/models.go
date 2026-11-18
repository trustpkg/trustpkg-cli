package terminalParameters

type Theme string

type Capability string

type Facts struct {
	IsStdinTTY  bool
	IsStdoutTTY bool
	IsStderrTTY bool

	TerminalWidth  int
	TerminalHeight int

	SupportsANSI          bool
	SupportsANSI16Colors  bool
	SupportsANSI256Colors bool
	SupportsTrueColor     bool

	TerminalTheme      Theme
	SupportsUnicode    Capability
	SupportsHyperlinks Capability

	CanAnimate    bool
	IsInteractive bool
}

