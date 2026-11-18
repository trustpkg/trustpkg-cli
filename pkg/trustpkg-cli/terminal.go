package trustpkgCli

import terminalApi "github.com/trustpkg/trustpkg-cli/internal/terminal-api"

type (
	Command      = terminalApi.Command
	Flag         = terminalApi.Flag
	FlagType     = terminalApi.FlagType
	FlagValue    = terminalApi.FlagValue
	Invocation   = terminalApi.Invocation
	Parser       = terminalApi.Parser
	Event        = terminalApi.Event
	EventKind    = terminalApi.EventKind
	EventOptions = terminalApi.EventOptions
	EventSource  = terminalApi.EventSource
)

const (
	StringFlag     = terminalApi.StringFlag
	BoolFlag       = terminalApi.BoolFlag
	ListFlag       = terminalApi.ListFlag
	InputEvent     = terminalApi.InputEvent
	ResizeEvent    = terminalApi.ResizeEvent
	InterruptEvent = terminalApi.InterruptEvent
	EOFEvent       = terminalApi.EOFEvent
)

var ErrEventsUnsupported = terminalApi.ErrEventsUnsupported

func NewParser(root Command) (Parser, error) {
	return terminalApi.NewParser(root)
}

func NewEventSource(options EventOptions) (EventSource, error) {
	return terminalApi.NewEventSource(options)
}
