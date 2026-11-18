//go:build !linux && !darwin && !freebsd && !netbsd && !openbsd && !dragonfly

package terminalApi

func NewEventSource(EventOptions) (EventSource, error) {
	return nil, ErrEventsUnsupported
}
