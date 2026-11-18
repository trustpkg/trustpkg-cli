package logger

import "github.com/trustpkg/trustpkg-cli-forge/internal/logs"

type (
	Logger         = logs.Logger
	NewLoggerInput = logs.NewLoggerInput
	Severity       = logs.Severity
	LoaderOption   = logs.LoaderOption
	LoaderLabel    = logs.LoaderLabel
	PrintOption    = logs.PrintOption
	PrintTimestamp = logs.PrintTimestamp
	SectionOption  = logs.SectionOption
	SectionDashed  = logs.SectionDashed
	InputConfig    = logs.InputConfig
	SelectInput    = logs.SelectInput
	SelectOption   = logs.SelectOption
	Window         = logs.Window
	WindowPanel    = logs.WindowPanel
)

const (
	SeverityLog     = logs.SeverityLog
	SeverityDebug   = logs.SeverityDebug
	SeverityInfo    = logs.SeverityInfo
	SeverityWarning = logs.SeverityWarning
	SeverityError   = logs.SeverityError
	SeveritySuccess = logs.SeveritySuccess
	SeverityStart   = logs.SeverityStart
	SeverityDone    = logs.SeverityDone
	WithTimestamp   = logs.WithTimestamp
	DashedSection   = logs.DashedSection
)

var (
	ErrFooterUnsupported = logs.ErrFooterUnsupported
	ErrPromptUnsupported = logs.ErrPromptUnsupported
	ErrPromptCancelled   = logs.ErrPromptCancelled
)

func NewLogger(input NewLoggerInput) *Logger {
	return logs.NewLogger(input)
}
