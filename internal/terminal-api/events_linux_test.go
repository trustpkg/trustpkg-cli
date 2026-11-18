//go:build linux

package terminalApi

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/charmbracelet/x/input"
	"golang.org/x/sys/unix"
	"golang.org/x/term"
)

func nextEvent(t *testing.T, source EventSource) Event {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	event, err := source.Next(ctx)
	if err != nil {
		t.Fatal(err)
	}
	return event
}

func openPTY(t *testing.T) (*os.File, *os.File) {
	t.Helper()
	fd, err := unix.Open("/dev/ptmx", unix.O_RDWR|unix.O_NOCTTY|unix.O_CLOEXEC, 0)
	if err != nil {
		t.Fatal(err)
	}
	master := os.NewFile(uintptr(fd), "ptmx")
	t.Cleanup(func() { master.Close() })
	if err := unix.IoctlSetPointerInt(fd, unix.TIOCSPTLCK, 0); err != nil {
		t.Fatal(err)
	}
	number, err := unix.IoctlGetInt(fd, unix.TIOCGPTN)
	if err != nil {
		t.Fatal(err)
	}
	slave, err := os.OpenFile(fmt.Sprintf("/dev/pts/%d", number), os.O_RDWR, 0)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { slave.Close() })
	if err := unix.IoctlSetWinsize(int(slave.Fd()), unix.TIOCSWINSZ, &unix.Winsize{Col: 80, Row: 24}); err != nil {
		t.Fatal(err)
	}
	return master, slave
}

func TestTerminalEventsAndRestore(t *testing.T) {
	master, slave := openPTY(t)
	before, err := term.GetState(int(slave.Fd()))
	if err != nil {
		t.Fatal(err)
	}
	source, err := NewEventSource(EventOptions{Input: slave, Output: slave, Keys: true, Resize: true, Interrupts: true})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := source.Close(); err != nil {
			t.Error(err)
		}
	})
	if event := nextEvent(t, source); event.Kind != ResizeEvent || event.Width != 80 || event.Height != 24 {
		t.Fatalf("initial size: %+v", event)
	}
	if _, err := master.Write([]byte("a\x1b[A\x03")); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"a", "up"} {
		event := nextEvent(t, source)
		key, ok := event.Input.(input.KeyPressEvent)
		if event.Kind != InputEvent || !ok || key.Keystroke() != name {
			t.Fatalf("expected %s, got %+v", name, event)
		}
	}
	if event := nextEvent(t, source); event.Kind != InterruptEvent {
		t.Fatalf("expected interrupt, got %+v", event)
	}
	if err := unix.IoctlSetWinsize(int(slave.Fd()), unix.TIOCSWINSZ, &unix.Winsize{Col: 110, Row: 35}); err != nil {
		t.Fatal(err)
	}
	if event := nextEvent(t, source); event.Kind != ResizeEvent || event.Width != 110 || event.Height != 35 {
		t.Fatalf("resized: %+v", event)
	}
	if err := source.Close(); err != nil {
		t.Fatal(err)
	}
	after, err := term.GetState(int(slave.Fd()))
	if err != nil || !reflect.DeepEqual(before, after) {
		t.Fatalf("terminal state not restored: %v", err)
	}
}

func TestPipeEOFAndCancellation(t *testing.T) {
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()
	source, err := NewEventSource(EventOptions{Input: reader, Keys: true})
	if err != nil {
		t.Fatal(err)
	}
	defer source.Close()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := source.Next(ctx); !errors.Is(err, context.Canceled) {
		t.Fatalf("expected cancellation, got %v", err)
	}
	if _, err := writer.Write([]byte("z")); err != nil {
		t.Fatal(err)
	}
	writer.Close()
	if event := nextEvent(t, source); event.Kind != InputEvent {
		t.Fatalf("expected key, got %+v", event)
	}
	if event := nextEvent(t, source); event.Kind != EOFEvent {
		t.Fatalf("expected EOF event, got %+v", event)
	}
	if _, err := source.Next(context.Background()); !errors.Is(err, io.EOF) {
		t.Fatalf("expected EOF, got %v", err)
	}
}

func TestCloseWithUnreadEvents(t *testing.T) {
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}

	defer reader.Close()
	defer writer.Close()
	source, err := NewEventSource(EventOptions{Input: reader, Keys: true})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { source.Close() })
	if _, err := writer.Write([]byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")); err != nil {
		t.Fatal(err)
	}
	s := source.(*eventSource)
	deadline := time.Now().Add(3 * time.Second)
	for len(s.results) < cap(s.results) && time.Now().Before(deadline) {
		time.Sleep(time.Millisecond)
	}
	if len(s.results) != cap(s.results) {
		t.Fatal("event queue did not fill")
	}
	done := make(chan error, 1)
	go func() { done <- source.Close() }()
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("Close blocked with unread events")
	}
}

func TestResizeErrorIsReported(t *testing.T) {
	_, slave := openPTY(t)
	source, err := NewEventSource(EventOptions{Output: slave, Resize: true})
	if err != nil {
		t.Fatal(err)
	}
	defer source.Close()
	nextEvent(t, source)
	if err := slave.Close(); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if _, err := source.Next(ctx); err == nil || errors.Is(err, io.EOF) {
		t.Fatalf("expected resize error, got %v", err)
	}
}

func TestExternalInterrupt(t *testing.T) {
	source, err := NewEventSource(EventOptions{Interrupts: true})
	if err != nil {
		t.Fatal(err)
	}
	defer source.Close()
	process, err := os.FindProcess(os.Getpid())
	if err != nil {
		t.Fatal(err)
	}
	if err := process.Signal(os.Interrupt); err != nil {
		t.Fatal(err)
	}
	if event := nextEvent(t, source); event.Kind != InterruptEvent {
		t.Fatalf("expected interrupt, got %+v", event)
	}
}
