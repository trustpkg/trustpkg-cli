package terminalParameters

import (
	"fmt"
	"os"

	"github.com/trustpkg/trustpkg-cli/internal/progress"
	"golang.org/x/term"
)

func GatherFacts(onProgress progress.Callback) (Facts, error) {
	facts := Facts{TerminalTheme: ThemeAuto}
	reportProgress := progress.NewReporter(gatherTotalSteps, onProgress)

	facts.IsStdinTTY = term.IsTerminal(int(os.Stdin.Fd()))
	reportProgress()

	facts.IsStdoutTTY = term.IsTerminal(int(os.Stdout.Fd()))
	reportProgress()

	facts.IsStderrTTY = term.IsTerminal(int(os.Stderr.Fd()))
	reportProgress()

	facts.IsInteractive = facts.IsStdinTTY && facts.IsStdoutTTY
	reportProgress()

	environment := processEnvironmentChecker{}

	facts.TerminalTheme = environment.theme()
	reportProgress()

	facts.SupportsUnicode = environment.unicode()
	reportProgress()

	facts.SupportsHyperlinks = environment.hyperlinks()
	reportProgress()

	facts.SupportsANSI = environment.supportsANSI(facts.IsStdoutTTY)
	reportProgress()

	facts.SupportsANSI16Colors = facts.SupportsANSI
	reportProgress()

	facts.SupportsANSI256Colors = environment.supportsANSI256Colors(facts.SupportsANSI)
	reportProgress()

	facts.SupportsTrueColor = environment.supportsTrueColor(facts.SupportsANSI)
	reportProgress()

	facts.CanAnimate = facts.SupportsANSI
	reportProgress()

	sizeFile := os.Stdout
	if !facts.IsStdoutTTY && facts.IsStdinTTY {
		sizeFile = os.Stdin
	} else if !facts.IsStdoutTTY && !facts.IsStdinTTY && facts.IsStderrTTY {
		sizeFile = os.Stderr
	}
	if term.IsTerminal(int(sizeFile.Fd())) {
		width, height, err := term.GetSize(int(sizeFile.Fd()))
		if err != nil {
			return facts, fmt.Errorf("get terminal size: %w", err)
		}
		facts.TerminalWidth = width
		facts.TerminalHeight = height
	}
	reportProgress()

	return facts, nil
}
