package theme

import (
	"github.com/trustpkg/trustpkg-cli/internal/progress"
	terminalParameters "github.com/trustpkg/trustpkg-cli/internal/terminal-parameters"
)

type Theme struct {
	Text        string `toml:"text"`
	Heading     string `toml:"heading"`
	Description string `toml:"description"`
	Muted       string `toml:"muted"`
	Accent      string `toml:"accent"`
	Link        string `toml:"link"`
	Code        string `toml:"code"`
	Key         string `toml:"key"`
	Value       string `toml:"value"`
	Border      string `toml:"border"`
	Separator   string `toml:"separator"`
	Background  string `toml:"background"`

	Log        string `toml:"log"`
	Debug      string `toml:"debug"`
	Info       string `toml:"info"`
	Warning    string `toml:"warning"`
	Error      string `toml:"error"`
	Success    string `toml:"success"`
	Pending    string `toml:"pending"`
	InProgress string `toml:"in_progress"`
	Done       string `toml:"done"`

	Prompt      string `toml:"prompt"`
	Input       string `toml:"input"`
	Placeholder string `toml:"placeholder"`
	Selected    string `toml:"selected"`
	Disabled    string `toml:"disabled"`

	ProgressFilled    string `toml:"progress_filled"`
	ProgressRemaining string `toml:"progress_remaining"`
	Spinner           string `toml:"spinner"`

	Reset string `toml:"-"`
}

type palette struct {
	text       string
	secondary  string
	accent     string
	info       string
	warning    string
	errorColor string
	success    string
}

type GatherThemeInput struct {
	terminalFacts terminalParameters.Facts
	themeFilePath *string
	onProgress    progress.Callback
}
