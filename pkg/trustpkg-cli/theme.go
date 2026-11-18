package trustpkgCli

import "github.com/trustpkg/trustpkg-cli/internal/theme"

type (
	Theme          = theme.Theme
	ThemeHandler   = theme.Handler
	ThemeInitInput = theme.ThemeInitInput
)

const (
	NoColor                     = theme.NoColor
	ANSIForegroundBlack         = theme.ANSIForegroundBlack
	ANSIForegroundRed           = theme.ANSIForegroundRed
	ANSIForegroundGreen         = theme.ANSIForegroundGreen
	ANSIForegroundYellow        = theme.ANSIForegroundYellow
	ANSIForegroundBlue          = theme.ANSIForegroundBlue
	ANSIForegroundMagenta       = theme.ANSIForegroundMagenta
	ANSIForegroundCyan          = theme.ANSIForegroundCyan
	ANSIForegroundWhite         = theme.ANSIForegroundWhite
	ANSIForegroundBrightBlack   = theme.ANSIForegroundBrightBlack
	ANSIForegroundBrightRed     = theme.ANSIForegroundBrightRed
	ANSIForegroundBrightGreen   = theme.ANSIForegroundBrightGreen
	ANSIForegroundBrightYellow  = theme.ANSIForegroundBrightYellow
	ANSIForegroundBrightBlue    = theme.ANSIForegroundBrightBlue
	ANSIForegroundBrightMagenta = theme.ANSIForegroundBrightMagenta
	ANSIForegroundBrightCyan    = theme.ANSIForegroundBrightCyan
	ANSIForegroundBrightWhite   = theme.ANSIForegroundBrightWhite
	ANSIBackgroundBlack         = theme.ANSIBackgroundBlack
	ANSIBackgroundRed           = theme.ANSIBackgroundRed
	ANSIBackgroundGreen         = theme.ANSIBackgroundGreen
	ANSIBackgroundYellow        = theme.ANSIBackgroundYellow
	ANSIBackgroundBlue          = theme.ANSIBackgroundBlue
	ANSIBackgroundMagenta       = theme.ANSIBackgroundMagenta
	ANSIBackgroundCyan          = theme.ANSIBackgroundCyan
	ANSIBackgroundWhite         = theme.ANSIBackgroundWhite
	ANSIBackgroundBrightBlack   = theme.ANSIBackgroundBrightBlack
	ANSIBackgroundBrightRed     = theme.ANSIBackgroundBrightRed
	ANSIBackgroundBrightGreen   = theme.ANSIBackgroundBrightGreen
	ANSIBackgroundBrightYellow  = theme.ANSIBackgroundBrightYellow
	ANSIBackgroundBrightBlue    = theme.ANSIBackgroundBrightBlue
	ANSIBackgroundBrightMagenta = theme.ANSIBackgroundBrightMagenta
	ANSIBackgroundBrightCyan    = theme.ANSIBackgroundBrightCyan
	ANSIBackgroundBrightWhite   = theme.ANSIBackgroundBrightWhite
)

func InitTheme(input ThemeInitInput) (*ThemeHandler, error) {
	return theme.Init(input)
}

func GenerateThemeFile(directory string, onProgress ProgressCallback) error {
	return theme.GenerateThemeFile(directory, onProgress)
}

func DefaultFullColorDarkTheme() Theme  { return theme.DefaultFullColorDarkTheme() }
func DefaultFullColorLightTheme() Theme { return theme.DefaultFullColorLightTheme() }
func DefaultLimitedSixteenColorDarkTheme() Theme {
	return theme.DefaultLimitedSixteenColorDarkTheme()
}
func DefaultLimitedSixteenColorLightTheme() Theme {
	return theme.DefaultLimitedSixteenColorLightTheme()
}
func NoColorTheme() Theme { return theme.NoColorTheme() }

func RGBForeground(red, green, blue uint8) string {
	return theme.RGBForeground(red, green, blue)
}

func RGBBackground(red, green, blue uint8) string {
	return theme.RGBBackground(red, green, blue)
}
