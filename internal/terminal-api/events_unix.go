//go:build linux || darwin || freebsd || netbsd || openbsd || dragonfly

package terminalApi

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/charmbracelet/x/input"
	"golang.org/x/term"
)

type eventResult struct {
	event Event
	err   error
}

type eventSource struct {
	options  EventOptions
	reader   *input.Reader
	state    *term.State
	inputFD  int
	outputFD int
	ctx      context.Context
	cancel   context.CancelFunc
	results  chan eventResult
	done     chan struct{}
	workers  sync.WaitGroup
	once     sync.Once
	closeErr error
}

func NewEventSource(options EventOptions) (EventSource, error) {
	if !options.Keys && !options.Resize && !options.Interrupts {
		return nil, errors.New("at least one terminal event type must be enabled")
	}
	if options.Keys && options.Input == nil {
		return nil, errors.New("input is required for key events")
	}
	if options.Resize && (options.Output == nil || !term.IsTerminal(int(options.Output.Fd()))) {
		return nil, errors.New("a terminal output is required for resize events")
	}
	s := &eventSource{options: options, results: make(chan eventResult, 32), done: make(chan struct{})}
	if options.Resize {
		s.outputFD = int(options.Output.Fd())
		if _, _, err := term.GetSize(s.outputFD); err != nil {
			return nil, fmt.Errorf("get terminal size: %w", err)
		}
	}
	if options.Keys {
		s.inputFD = int(options.Input.Fd())
		var err error
		s.reader, err = input.NewReader(options.Input, os.Getenv("TERM"), 0)
		if err != nil {
			return nil, fmt.Errorf("open terminal input: %w", err)
		}
		if term.IsTerminal(s.inputFD) {
			s.state, err = term.MakeRaw(s.inputFD)
			if err != nil {
				return nil, errors.Join(fmt.Errorf("enter raw mode: %w", err), s.reader.Close())
			}
		}
	}
	s.ctx, s.cancel = context.WithCancel(context.Background())
	if options.Keys {
		s.workers.Add(1)
		go s.readInput()
	}
	if options.Resize {
		s.workers.Add(1)
		go s.readSize()
	}
	if options.Interrupts {
		signals := make(chan os.Signal, 1)
		signal.Notify(signals, os.Interrupt)
		s.workers.Add(1)
		go s.readInterrupts(signals)
	}
	go func() {
		s.workers.Wait()
		close(s.results)
		close(s.done)
	}()
	return s, nil
}

func (s *eventSource) send(result eventResult) bool {
	select {
	case s.results <- result:
		return true
	case <-s.ctx.Done():
		return false
	}
}

func (s *eventSource) readInput() {
	defer s.workers.Done()
	for {
		events, err := s.reader.ReadEvents()
		for _, event := range events {
			result := Event{Kind: InputEvent, Input: event}
			if key, ok := event.(input.KeyPressEvent); ok && key.Keystroke() == "ctrl+c" && s.options.Interrupts {
				result = Event{Kind: InterruptEvent}
			}
			if !s.send(eventResult{event: result}) {
				return
			}
		}
		if err != nil {
			if s.ctx.Err() == nil {
				if errors.Is(err, io.EOF) {
					s.send(eventResult{event: Event{Kind: EOFEvent}})
				} else {
					s.send(eventResult{err: fmt.Errorf("read terminal input: %w", err)})
				}
				s.cancel()
			}
			return
		}
	}
}

func (s *eventSource) readSize() {
	defer s.workers.Done()
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	width, height := 0, 0
	first := true
	for {
		w, h, err := term.GetSize(s.outputFD)
		if err != nil {
			s.send(eventResult{err: fmt.Errorf("get terminal size: %w", err)})
			return
		}
		if first || w != width || h != height {
			first = false
			width, height = w, h
			if !s.send(eventResult{event: Event{Kind: ResizeEvent, Width: w, Height: h}}) {
				return
			}
		}
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

func (s *eventSource) readInterrupts(signals chan os.Signal) {
	defer s.workers.Done()
	defer signal.Stop(signals)
	for {
		select {
		case <-s.ctx.Done():
			return
		case <-signals:
			if !s.send(eventResult{event: Event{Kind: InterruptEvent}}) {
				return
			}
		}
	}
}

func (s *eventSource) Next(ctx context.Context) (Event, error) {
	if ctx == nil {
		return Event{}, errors.New("context must not be nil")
	}
	select {
	case result, ok := <-s.results:
		if !ok {
			return Event{}, io.EOF
		}
		return result.event, result.err
	case <-ctx.Done():
		return Event{}, ctx.Err()
	}
}

func (s *eventSource) Close() error {
	s.once.Do(func() {
		s.cancel()
		if s.reader != nil && !s.reader.Cancel() {
			s.closeErr = errors.Join(errors.New("could not cancel terminal input"), s.reader.Close())
			if s.state != nil {
				s.closeErr = errors.Join(s.closeErr, term.Restore(s.inputFD, s.state))
			}
			return
		}
		<-s.done
		if s.reader != nil {
			s.closeErr = s.reader.Close()
		}
		if s.state != nil {
			s.closeErr = errors.Join(s.closeErr, term.Restore(s.inputFD, s.state))
		}
	})
	return s.closeErr
}
