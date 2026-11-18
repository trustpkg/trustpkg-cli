package terminalApi

import (
	"context"
	"errors"
	"os"

	"github.com/charmbracelet/x/input"
)

var ErrEventsUnsupported = errors.New("terminal events are not supported on this platform")

type EventKind string

const (
	InputEvent     EventKind = "input"
	ResizeEvent    EventKind = "resize"
	InterruptEvent EventKind = "interrupt"
	EOFEvent       EventKind = "eof"
)

type Event struct {
	Kind   EventKind
	Input  input.Event
	Width  int
	Height int
}

type EventOptions struct {
	Input      *os.File
	Output     *os.File
	Keys       bool
	Resize     bool
	Interrupts bool
}

type EventSource interface {
	Next(context.Context) (Event, error)
	Close() error
}
