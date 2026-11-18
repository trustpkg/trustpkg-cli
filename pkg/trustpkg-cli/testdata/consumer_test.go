package consumer

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/charmbracelet/x/input"
	cli "github.com/trustpkg/trustpkg-cli-forge/pkg/trustpkg-cli"
)

func TestPublicComponents(t *testing.T) {
	var updates []cli.ProgressUpdate
	facts, err := cli.GatherFacts(func(update cli.ProgressUpdate) {
		updates = append(updates, update)
	})
	if err != nil || len(updates) == 0 {
		t.Fatalf("facts: %v, progress: %v", err, updates)
	}
	dir := t.TempDir()
	if err := cli.GenerateThemeFile(dir, nil); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "theme.toml")
	handler, err := cli.InitTheme(cli.ThemeInitInput{TerminalFacts: facts, ThemeFilePath: &path})
	if err != nil {
		t.Fatal(err)
	}
	logger := cli.NewLogger(cli.NewLoggerInput{Facts: facts, Theme: handler.Theme()})
	if err := logger.Print("ready", cli.WithTimestamp, cli.SeverityDone); err != nil {
		t.Fatal(err)
	}
	finish, err := logger.StartSection("commands", cli.SeverityInfo, cli.DashedSection)
	if err != nil {
		t.Fatal(err)
	}
	if err := logger.Command("go version"); err != nil {
		t.Fatal(err)
	}
	if err := finish(); err != nil {
		t.Fatal(err)
	}

	var _ *cli.Window
	var _ *cli.WindowPanel
	var _ = cli.SelectInput{Options: []cli.SelectOption{{Label: "x", Value: "x"}}}
	var _ = cli.InputConfig{Title: "name"}
	var _ cli.LoaderOption = cli.LoaderLabel("loading")
	var _ cli.PrintOption = cli.SeverityError
	var _ cli.SectionOption = cli.SeverityInfo
	var _ cli.TerminalTheme = cli.ThemeDark
	var _ cli.Capability = cli.CapabilityUnknown
}

func TestPublicParser(t *testing.T) {
	parser, err := cli.NewParser(cli.Command{
		Name: "cli",
		Flags: []cli.Flag{
			{Name: "output", Short: "o", Type: cli.ListFlag, Persistent: true},
			{Name: "enabled", Type: cli.BoolFlag, Persistent: true},
		},
		Commands: []cli.Command{{Name: "generate"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	result, err := parser.Parse([]string{"-o", "a,b", "--enabled", "true", "generate", "file"})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(result.Command, []string{"cli", "generate"}) ||
		!reflect.DeepEqual(result.Arguments, []string{"file"}) ||
		!reflect.DeepEqual(result.Flags["output"].List, []string{"a", "b"}) ||
		!result.Flags["enabled"].Bool {
		t.Fatalf("unexpected result: %+v", result)
	}
}

func TestPublicEvents(t *testing.T) {
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()
	defer writer.Close()
	source, err := cli.NewEventSource(cli.EventOptions{Input: reader, Keys: true})
	if errors.Is(err, cli.ErrEventsUnsupported) {
		t.Skip(err)
	}
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := source.Close(); err != nil {
			t.Error(err)
		}
	}()
	if _, err := writer.Write([]byte("x")); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	event, err := source.Next(ctx)
	if err != nil || event.Kind != cli.InputEvent {
		t.Fatalf("input event: %+v, %v", event, err)
	}
	key, ok := event.Input.(input.KeyPressEvent)
	if !ok || key.Keystroke() != "x" {
		t.Fatalf("key: %+v", event.Input)
	}
	event, err = source.Next(ctx)
	if err != nil || event.Kind != cli.EOFEvent {
		t.Fatalf("EOF event: %+v, %v", event, err)
	}
	if _, err := source.Next(ctx); !errors.Is(err, io.EOF) {
		t.Fatalf("expected EOF: %v", err)
	}
}
