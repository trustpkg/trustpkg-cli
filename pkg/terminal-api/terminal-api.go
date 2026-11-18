package terminalApi

import api "github.com/trustpkg/trustpkg-cli-forge/internal/terminal-api"

type (
	Command      = api.Command
	Flag         = api.Flag
	FlagType     = api.FlagType
	FlagValue    = api.FlagValue
	Invocation   = api.Invocation
	Parser       = api.Parser
	Event        = api.Event
	EventKind    = api.EventKind
	EventOptions = api.EventOptions
	EventSource  = api.EventSource
)

const (
	StringFlag     = api.StringFlag
	BoolFlag       = api.BoolFlag
	ListFlag       = api.ListFlag
	InputEvent     = api.InputEvent
	ResizeEvent    = api.ResizeEvent
	InterruptEvent = api.InterruptEvent
	EOFEvent       = api.EOFEvent
)

var ErrEventsUnsupported = api.ErrEventsUnsupported

func NewParser(root Command) (Parser, error) {
	return api.NewParser(root)
}

func NewEventSource(options EventOptions) (EventSource, error) {
	return api.NewEventSource(options)
}
