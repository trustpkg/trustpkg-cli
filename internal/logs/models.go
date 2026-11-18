package logs

import "fmt"

type Severity string

type LoaderOption interface {
	applyLoaderOption(*loaderOptions) error
}

type LoaderLabel string

type PrintOption interface {
	applyPrintOption(*printOptions) error
}

type PrintTimestamp bool

const WithTimestamp PrintTimestamp = true

type SectionOption interface {
	applySectionOption(*sectionOptions) error
}

type SectionDashed bool

const DashedSection SectionDashed = true

type sectionOptions struct {
	severity    Severity
	severitySet bool
	dashed      bool
	dashedSet   bool
}

type printOptions struct {
	severity     Severity
	severitySet  bool
	timestamp    bool
	timestampSet bool
}

type loaderOptions struct {
	severity    Severity
	label       string
	severitySet bool
	labelSet    bool
}

func (severity Severity) applySectionOption(options *sectionOptions) error {
	if options.severitySet {
		return fmt.Errorf("section accepts at most one severity")
	}
	options.severity, options.severitySet = severity, true
	return nil
}

func (dashed SectionDashed) applySectionOption(options *sectionOptions) error {
	if options.dashedSet {
		return fmt.Errorf("section accepts at most one line style")
	}
	options.dashed, options.dashedSet = bool(dashed), true
	return nil
}

func (severity Severity) applyPrintOption(options *printOptions) error {
	if options.severitySet {
		return fmt.Errorf("print accepts at most one severity")
	}
	options.severity, options.severitySet = severity, true
	return nil
}

func (timestamp PrintTimestamp) applyPrintOption(options *printOptions) error {
	if options.timestampSet {
		return fmt.Errorf("print accepts at most one timestamp option")
	}
	options.timestamp, options.timestampSet = bool(timestamp), true
	return nil
}

func (severity Severity) applyLoaderOption(options *loaderOptions) error {
	if options.severitySet {
		return fmt.Errorf("loader accepts at most one severity")
	}
	options.severity, options.severitySet = severity, true
	return nil
}

func (label LoaderLabel) applyLoaderOption(options *loaderOptions) error {
	if options.labelSet {
		return fmt.Errorf("loader accepts at most one label")
	}
	options.label, options.labelSet = string(label), true
	return nil
}
