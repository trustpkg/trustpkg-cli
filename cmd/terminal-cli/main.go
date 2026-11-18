package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/trustpkg/trustpkg-cli-forge/internal/logs"
	"github.com/trustpkg/trustpkg-cli-forge/internal/progress"
	terminalParameters "github.com/trustpkg/trustpkg-cli-forge/internal/terminal-parameters"
	theme "github.com/trustpkg/trustpkg-cli-forge/internal/theme"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() (result error) {
	facts, err := terminalParameters.GatherFacts(func(update progress.Update) {
		fmt.Printf("Step %d/%d\n", update.CurrentStep, update.TotalSteps)
	})
	if err != nil {
		return err
	}

	handler, err := theme.Init(theme.ThemeInitInput{
		ThemeFilePath: nil,
		TerminalFacts: facts,
		OnProgress: func(update progress.Update) {
			fmt.Printf("Step %d/%d\n", update.CurrentStep, update.TotalSteps)
		},
	})
	if err != nil {
		return err
	}

	fmt.Println(handler.Theme().Accent)

	fmt.Printf("%sAccent%s\n", handler.Theme().Error, handler.Theme().Reset)

	maxWidth := 120
	logger := logs.NewLogger(logs.NewLoggerInput{Facts: facts, Theme: handler.Theme(), MaxWidth: &maxWidth})
	if err := runLoggerExamples(logger); err != nil {
		return err
	}
	return runPromptExamples(logger)
}

func runLoggerExamples(logger *logs.Logger) (result error) {
	const steps = 10
	footerActive := false
	if err := logger.StartFooter("Preparing image download"); err != nil {
		if !errors.Is(err, logs.ErrFooterUnsupported) {
			return err
		}
		fmt.Fprintln(os.Stderr, "Warning:", err)
	} else {
		footerActive = true
		defer func() { result = errors.Join(result, logger.StopFooter()) }()
	}

	finishSection, err := logger.StartSection("Logger examples", logs.SeverityInfo, logs.DashedSection)
	if err != nil {
		return err
	}
	defer func() { result = errors.Join(result, finishSection()) }()

	if err := logger.Print("Plain message"); err != nil {
		return err
	}
	if err := logger.Print("Message with timestamp", logs.WithTimestamp); err != nil {
		return errors.Join(result, err)
	}
	if err := logger.Print("Log message", logs.SeverityLog); err != nil {
		return err
	}
	if err := logger.Print("Debug message", logs.SeverityDebug); err != nil {
		return err
	}
	if err := logger.Print("Information message", logs.SeverityInfo); err != nil {
		return err
	}
	if err := logger.Print("Warning message", logs.SeverityWarning); err != nil {
		return err
	}
	if err := logger.Print("Error message", logs.WithTimestamp, logs.SeverityError); err != nil {
		return err
	}
	if err := logger.Print("Success message", logs.WithTimestamp, logs.SeveritySuccess); err != nil {
		return err
	}
	if err := logger.Print("Starting", logs.WithTimestamp, logs.SeverityStart); err != nil {
		return err
	}
	if err := logger.Print("Finished", logs.WithTimestamp, logs.SeverityDone); err != nil {
		return err
	}
	if err := logger.Command("docker run --rm --name example-container --publish 8080:80 --env APP_ENV=development --volume /tmp/example:/data example/image:latest"); err != nil {
		return err
	}

	for step := 0; step <= steps; step++ {
		if step > 0 {
			time.Sleep(time.Millisecond * 100)
		}
		if err := logger.Loader(steps, step, logs.SeverityLog); err != nil {
			return err
		}
	}

	if footerActive {
		if err := logger.UpdateFooter("Downloading image layers"); err != nil {
			return err
		}
	}
	for step := 0; step <= steps; step++ {
		if step > 0 {
			time.Sleep(time.Millisecond * 100)
		}
		if err := logger.Loader(steps, step, logs.SeverityLog); err != nil {
			return err
		}
	}
	if footerActive {
		if err := logger.UpdateFooter("Verifying image"); err != nil {
			return err
		}
	}
	for step := 0; step <= steps; step++ {
		if step > 0 {
			time.Sleep(time.Millisecond * 100)
		}
		if err := logger.Loader(steps, step, logs.SeverityLog, logs.LoaderLabel("Verifying")); err != nil {
			return err
		}
	}
	return nil
}

func runPromptExamples(logger *logs.Logger) (result error) {
	packages, err := logger.Select(logs.SelectInput{
		Title:    "Which development libraries should be installed?",
		Required: true,
		Options: []logs.SelectOption{
			{Label: "OpenSSL", Value: "openssl-devel", Description: "TLS and cryptography"},
			{Label: "libcurl", Value: "libcurl-devel", Description: "HTTP client"},
			{Label: "SQLite", Value: "sqlite-devel", Description: "Embedded database"},
			{Label: "zlib", Value: "zlib-devel", Description: "Compression"},
			{Label: "libxml2", Value: "libxml2-devel", Description: "XML parsing"},
		},
	})
	if err != nil {
		return err
	}
	extra, err := logger.Input(logs.InputConfig{
		Title:       "Additional package (optional)",
		Placeholder: "e.g. gcc",
		Validate:    validatePackageName,
	})
	if err != nil {
		return err
	}
	if extra != "" {
		packages = append(packages, extra)
	}

	finish, err := logger.StartSection("Installation command", logs.SeverityInfo)
	if err != nil {
		return err
	}
	defer func() { result = errors.Join(result, finish()) }()
	if err := logger.Print("Selected packages: "+strings.Join(packages, ", "), logs.SeverityInfo); err != nil {
		return err
	}
	return logger.Command("sudo dnf install -- " + strings.Join(packages, " "))
}

func validatePackageName(name string) error {
	if name == "" {
		return nil
	}
	for i, r := range name {
		alphanumeric := r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9'
		if !alphanumeric && (i == 0 || !strings.ContainsRune("._+-", r)) {
			return fmt.Errorf("Use a package name with letters, digits, '.', '_', '+' or '-'.")
		}
	}
	return nil
}
