package main

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/trustpkg/trustpkg-cli-forge/internal/logs"
	terminalParameters "github.com/trustpkg/trustpkg-cli-forge/internal/terminal-parameters"
	"github.com/trustpkg/trustpkg-cli-forge/internal/theme"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() (result error) {
	facts, err := terminalParameters.GatherFacts(nil)
	if err != nil {
		return err
	}
	handler, err := theme.Init(theme.ThemeInitInput{TerminalFacts: facts})
	if err != nil {
		return err
	}
	logger := logs.NewLogger(logs.NewLoggerInput{Facts: facts, Theme: handler.Theme()})
	window, err := logger.StartWindow()
	if err != nil {
		return err
	}
	defer func() { result = errors.Join(result, window.Close()) }()

	names := []string{
		"Gathering facts",
		"Resolving dependencies",
		"Downloading packages",
		"Verifying artifacts",
		"Preparing workspace",
	}
	panels := make([]*logs.WindowPanel, len(names))
	for i, name := range names {
		panels[i], err = window.AddPanel(name)
		if err != nil {
			return err
		}
	}

	results := make(chan error, len(panels))
	for i, panel := range panels {
		delay := 350*time.Millisecond + time.Duration(i)*60*time.Millisecond
		go func(panel *logs.WindowPanel, delay time.Duration) {
			results <- runPanel(panel, delay)
		}(panel, delay)
	}
	for range panels {
		result = errors.Join(result, <-results)
	}
	return result
}

func runPanel(panel *logs.WindowPanel, delay time.Duration) (result error) {
	defer func() { result = errors.Join(result, panel.Finish()) }()
	logger := panel.Logger()
	if err := logger.Print("Starting simulated task", logs.SeverityStart); err != nil {
		return err
	}
	const steps = 10
	for step := 0; step <= steps; step++ {
		if step > 0 {
			time.Sleep(delay)
		}
		if step > 0 && step < steps && step%3 == 0 {
			if err := logger.Print(fmt.Sprintf("Completed %d of %d steps", step, steps), logs.SeverityInfo); err != nil {
				return err
			}
		}
		if err := logger.Loader(steps, step, logs.SeverityInfo); err != nil {
			return err
		}
	}
	return logger.Print("Simulated task completed", logs.SeverityDone)
}
