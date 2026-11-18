package theme

import "fmt"

func RGBForeground(red, green, blue uint8) string {
	return fmt.Sprintf("\x1b[38;2;%d;%d;%dm", red, green, blue)
}

func RGBBackground(red, green, blue uint8) string {
	return fmt.Sprintf("\x1b[48;2;%d;%d;%dm", red, green, blue)
}

func themeFromPalette(p palette) Theme {
	return Theme{
		Text:              p.text,
		Heading:           p.text,
		Description:       p.secondary,
		Muted:             p.secondary,
		Accent:            p.accent,
		Link:              p.accent,
		Code:              p.text,
		Key:               p.accent,
		Value:             p.text,
		Border:            p.secondary,
		Separator:         p.secondary,
		Background:        NoColor,
		Log:               p.text,
		Debug:             p.secondary,
		Info:              p.info,
		Warning:           p.warning,
		Error:             p.errorColor,
		Success:           p.success,
		Pending:           p.secondary,
		InProgress:        p.info,
		Done:              p.success,
		Prompt:            p.accent,
		Input:             p.text,
		Placeholder:       p.secondary,
		Selected:          p.accent,
		Disabled:          p.secondary,
		ProgressFilled:    p.accent,
		ProgressRemaining: p.secondary,
		Spinner:           p.accent,
		Reset:             reset,
	}
}
