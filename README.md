# trustpkg-cli-forge
trustpkg-cli-forge is  coherent and shared cli system

## Use in another project

The public import is `github.com/trustpkg/trustpkg-cli-forge/pkg/trustpkg-cli-forge`.
After the module has been published, add it with:

```sh
go get github.com/trustpkg/trustpkg-cli-forge/pkg/trustpkg-cli-forge
```

For local development before publication, use `go mod edit -replace
github.com/trustpkg/trustpkg-cli-forge=/path/to/trustpkg-cli-forge` in the consuming
project, then run `go mod tidy`. The module currently requires Go 1.26.10 or
newer.

```go
package main

import (
	"fmt"
	"os"

	cli "github.com/trustpkg/trustpkg-cli-forge/pkg/trustpkg-cli-forge"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	facts, err := cli.GatherFacts(nil)
	if err != nil {
		return err
	}
	handler, err := cli.InitTheme(cli.ThemeInitInput{TerminalFacts: facts})
	if err != nil {
		return err
	}
	logger := cli.NewLogger(cli.NewLoggerInput{Facts: facts, Theme: handler.Theme()})
	return logger.Print("Ready", cli.WithTimestamp, cli.SeveritySuccess)
}
```

Run the terminal information collector and interactive prompt examples:

```sh
make run
```

The examples ask which development libraries to include and optionally accept
an additional package. They display a `dnf install` command but do not run it.
The prompts require an interactive ANSI terminal.

Run the window demo with five simulated concurrent tasks:

```sh
make run-window
```

`GatherFacts` reports progress after each completed step through an optional
callback. Each update contains the current step and the total number of steps;
the collected facts and any error are returned by the function. The 13 steps
cover each TTY check, each capability or derived field, and one terminal size
check (width and height together). If a step fails, the error is returned
without reporting that step as completed.

The public `ProgressCallback`, `ProgressUpdate`, and `NewProgressReporter`
provide progress reporting for other operations.

The default terminal palettes are adapted from the site's light and dark
themes. They preserve the terminal background; their true-color foregrounds
are tested against the corresponding site backgrounds for at least 4.5:1
contrast. Actual contrast depends on the terminal's configured background
and, for the 16-color palettes, its ANSI color settings.
The exported theme constructors return independent values; individual `Theme`
fields can be overridden without configuring the rest of the theme.
Use `cli.RGBForeground(r, g, b)` or `cli.RGBBackground(r, g, b)` for
any 24-bit color; each component ranges from 0 to 255. The package also provides
foreground and background constants for all 16 ANSI colors. Colored output
requires a compatible terminal, and custom colors should be checked for contrast
against the actual background.

Theme TOML validation requires four quoted tables: `["full colors light"]`,
`["full colors dark"]`, `["limited colors light"]`, and
`["limited colors dark"]`. Every table must contain these string keys:
`text`, `heading`, `description`, `muted`, `accent`, `link`, `code`, `key`,
`value`, `border`, `separator`, `background`, `log`, `debug`, `info`,
`warning`, `error`, `success`, `pending`, `in_progress`, `done`, `prompt`,
`input`, `placeholder`, `selected`, `disabled`, `progress_filled`,
`progress_remaining`, and `spinner`. Reset is automatic and is not a TOML key.

Full-color values use `"#RRGGBB"`; limited-color values use one of `"black"`,
`"red"`, `"green"`, `"yellow"`, `"blue"`, `"magenta"`, `"cyan"`, or
`"white"`, optionally prefixed with `bright_` (for example `"bright_red"`).
Unknown tables, keys, and invalid values are rejected. Validation does not yet
apply a theme to terminal output. The file loader chooses the full-color or
16-color table and the light or dark variant using the terminal facts. It
returns an error when the appearance is unknown, and a no-color theme when ANSI
is unavailable.

`cli.GenerateThemeFile` creates `theme.toml` in an existing directory
using the default palettes. It returns an error if the file already exists.
`background = "default"` preserves the terminal's background.

Theme initialization uses an explicitly supplied `ThemeFilePath` when present.
Otherwise it looks for `trustpkg-cli-forge-theme.toml` in the working directory and
uses the built-in defaults only if that file does not exist. Invalid or
unreadable theme files produce an error.

The logger's `Loader(steps, currentStep, options...)` redraws a
single progress line and finishes it when all steps are complete. Severity is
optional and defaults to `info`. The logger uses the terminal width when known,
or 80 columns otherwise. When output is redirected, carriage returns are
preserved in the output; files and pipes cannot redraw earlier text.
Pass `cli.LoaderLabel("Downloading")` to add a prefix, optionally alongside a
severity such as `cli.SeverityInfo`. The label is included in the width limit.
`NewLoggerInput.MaxWidth` optionally limits loader, footer and log lines to a
positive number of columns. Long logs wrap; footer status text is shortened
to leave room for the spinner. A nil value leaves the terminal width
(or the 80-column fallback for the loader) as the loader and footer limit;
it adds no width limit to ordinary logs.

`StartFooter(status)` reserves two bottom rows for a status and spinner;
`UpdateFooter(status)` changes the status, and `StopFooter()` stops the
animation and restores scrolling. Use the same logger's `Log` and `Loader` to
write above the footer; independent writers can disrupt it. Unsupported
terminals return `cli.ErrFooterUnsupported` without displaying a footer.
Always call `StopFooter` after a successful start, including on errors.

`Print(message, options...)` writes above an active footer and obeys the same
width limit as `Log`. By default it adds no prefix. Pass `cli.WithTimestamp`
for a local date and time prefix (`[2006-01-02 15:04:05]`), a severity such
as `cli.SeverityError` or `cli.SeverityStart` for a colored `Error:` or
`Start:` prefix, or both:

```go
logger.Print("Connecting")
logger.Print("Connecting", cli.WithTimestamp)
logger.Print("Connecting", cli.SeverityStart)
logger.Print("Failed", cli.WithTimestamp, cli.SeverityError)
logger.Print("Complete", cli.WithTimestamp, cli.SeverityDone)
```

Check the returned error for invalid options, width or output failures.

`Command(command)` displays a command between colored separator lines, without
a prompt or label, using the theme's code color. The separators span the
available terminal or panel width, subject to `MaxWidth`. Select the command lines
without the separators to copy them. It writes above an active footer. Long simple shell commands
are split at unquoted spaces using `\` line continuations so they can be
copied into a POSIX-style shell. If a safe split cannot be determined, the
command is displayed unchanged, even if it exceeds the width limit. Commands
must be valid UTF-8 and cannot contain control characters or newlines.

`StartSection(title)` writes a heading after a starting separator. Call the
returned finish function to write the closing separator, including on errors.
Sections may be nested and must be finished in reverse order. Separators
respect the terminal width and `MaxWidth`. An optional severity colors both
separators; without one, the theme's separator color is used. Separators are
solid by default; pass `cli.DashedSection` to make them dashed.

```go
finish, err := logger.StartSection("Downloading", cli.SeverityInfo, cli.DashedSection)
if err != nil {
	return err
}
defer func() { result = errors.Join(result, finish()) }()
```

`Select(SelectInput)` provides a multi-select list. Up/Down moves the cursor,
Tab selects the current option, Space toggles it, and Enter selects the
current option and confirms. Ctrl+A toggles all options; Ctrl+D confirms
without selecting the current option. Set `Required` to require at least one choice. Each
option supplies a label, value, optional description, and initial selection;
the result contains selected values in option order.

`Input(InputConfig)` provides an editable text field with cursor movement,
optional initial value, placeholder, required flag and validation callback.
Enter confirms and Esc or Ctrl+C cancels either prompt. Both return
`cli.ErrPromptUnsupported` without an interactive ANSI terminal and
`cli.ErrPromptCancelled` on cancellation. They cannot run while the logger
has an active footer, loader, window or another prompt. A configured `MaxWidth`
must be at least 12 columns; a temporarily smaller terminal must be resized or
the prompt cancelled.

`StartWindow()` reserves the logger for a full-terminal grid of independent
panels. `AddPanel(name)` returns a panel with its own `Logger()`, which can be
used by a concurrent worker for `Print`, `Log`, `Loader`, `Command`, and
sections. Panels have separate histories and an automatic spinner; call
`Finish()` when a worker finishes. Add and remove panels with `AddPanel` and
`RemovePanel` while the window is open. Removed or finished panels reject
further writes.

Use Tab / Shift+Tab to select a panel, Up / Down to scroll its history, and
End to return to its newest output. The grid adapts to resizing and pages
panels when the terminal is too small to show them all. Ctrl+C interrupts the
window; subsequent operations and `Close` report the interruption. On
noninteractive terminals, output becomes a labeled stream instead. A panel
does not support its own footer or nested window. Do not write directly to the
terminal or use the parent logger while the window is open.

Call `Close()` after all workers finish, including on errors. It restores the
terminal and keeps the final view visible; a named return value allows cleanup
errors to be reported:

```go
func runTasks(logger *cli.Logger) (result error) {
	window, err := logger.StartWindow()
	if err != nil {
		return err
	}
	defer func() { result = errors.Join(result, window.Close()) }()

	panels := make([]*cli.WindowPanel, 5)
	for i := range panels {
		panels[i], err = window.AddPanel(fmt.Sprintf("Task %d", i+1))
		if err != nil {
			return err
		}
	}
	var workers sync.WaitGroup
	errs := make(chan error, len(panels))
	for _, panel := range panels {
		workers.Add(1)
		go func(panel *cli.WindowPanel) {
			defer workers.Done()
			err := panel.Logger().Print("Working", cli.SeverityInfo)
			errs <- errors.Join(err, panel.Finish())
		}(panel)
	}
	workers.Wait()
	close(errs)
	for err := range errs {
		result = errors.Join(result, err)
	}
	return result
}
```

## Terminal API

The public terminal API collects command-line arguments without running
application actions. Define the command tree and flags, then parse arguments
without the executable name:

```go
import cli "github.com/trustpkg/trustpkg-cli-forge/pkg/trustpkg-cli-forge"

parser, err := cli.NewParser(cli.Command{
	Name: "trustpkg-cli-forge",
	Flags: []cli.Flag{
		{Name: "output", Short: "o", Type: cli.ListFlag, Persistent: true},
		{Name: "should-use", Type: cli.BoolFlag, Persistent: true},
	},
	Commands: []cli.Command{{Name: "generate", Description: "Generate files"}},
})
if err != nil {
	return err
}
invocation, err := parser.Parse(os.Args[1:])
if err != nil {
	return err
}
if invocation.HelpRequested {
	fmt.Print(invocation.Help)
	return nil
}
fmt.Println(invocation.Command, invocation.Arguments)
fmt.Println(invocation.Flags["output"].List, invocation.Flags["should-use"].Bool)
```

For example, `generate xxfdfd -o ffdfdf,test --should-use true` yields the
command path, positional argument, two output values and the boolean.
List flags accept CSV and combine repeated occurrences; string flags retain
commas. Boolean flags require an explicit value, either `--should-use true`
or `--should-use=false`. `--` ends flag parsing. Flags are local to their
command unless `Persistent` is set, in which case descendants inherit them.
Commands with subcommands reject positional arguments at that level; put
positional arguments on a leaf command. `Changed` distinguishes a supplied
flag from its default. Unknown commands, flags, invalid values and invalid
schemas return errors. `--help`, `-h` and `help <command>` return generated
help in the result rather than printing it.

`NewEventSource` optionally collects input events, terminal size changes and
interrupts. For example:

```go
source, err := cli.NewEventSource(cli.EventOptions{
	Input: os.Stdin, Output: os.Stdout, Keys: true, Resize: true, Interrupts: true,
})
if err != nil {
	return err
}
defer func() { result = errors.Join(result, source.Close()) }()
for {
	event, err := source.Next(ctx)
	if err != nil {
		return err
	}
	switch event.Kind {
	case cli.InputEvent:
		fmt.Println(event.Input) // e.g. input.KeyPressEvent
	case cli.ResizeEvent:
		fmt.Println(event.Width, event.Height)
	case cli.InterruptEvent, cli.EOFEvent:
		return nil
	}
}
```

The snippet assumes a function with a named `result error` return and a
`context.Context` named `ctx`. The event source requires exclusive use of its
input: close it before starting `Select`, `Input`, `Window`, or another reader.
This exclusivity is the caller's responsibility and is not automatically
enforced. On a TTY, key input switches to raw mode; always close the source to
restore the terminal, even after EOF, an error, or context cancellation. `Next`
cancellation only cancels that wait. Keep the supplied files open until
`Close` completes. Input EOF produces `EOFEvent` and stops the source; buffered
events may remain before `Next` returns `io.EOF`. Read and resize errors are
returned by `Next`. Ctrl+C in raw mode
and process interrupt signals produce `InterruptEvent` when enabled. With
interrupts disabled, Ctrl+C is an ordinary input event.

Resize reports the initial size and then polls for changes every 100 ms, so
intermediate sizes can be coalesced. Input events use Charm's `x/input`
decoder; escape sequences or UTF-8 characters split across separate reads may
not be decoded as a single key. Slow consumers can block event producers;
`Close` unblocks them. Terminal events are supported on Linux, macOS and BSD;
other platforms return `ErrEventsUnsupported`. Resize requires a terminal
output; key input may also be a pipe, but the platform must support
cancellable reads from that input.
