package theme

func NoColorTheme() Theme {
	return Theme{
		Text:        NoColor,
		Heading:     NoColor,
		Description: NoColor,
		Muted:       NoColor,
		Accent:      NoColor,
		Link:        NoColor,
		Code:        NoColor,
		Key:         NoColor,
		Value:       NoColor,
		Border:      NoColor,
		Separator:   NoColor,
		Background:  NoColor,

		Log:        NoColor,
		Debug:      NoColor,
		Info:       NoColor,
		Warning:    NoColor,
		Error:      NoColor,
		Success:    NoColor,
		Pending:    NoColor,
		InProgress: NoColor,
		Done:       NoColor,

		Prompt:      NoColor,
		Input:       NoColor,
		Placeholder: NoColor,
		Selected:    NoColor,
		Disabled:    NoColor,

		ProgressFilled:    NoColor,
		ProgressRemaining: NoColor,
		Spinner:           NoColor,

		Reset: NoColor,
	}
}

func DefaultFullColorLightTheme() Theme {
	theme := themeFromPalette(palette{
		text:       foregroundDarkSlate,
		secondary:  foregroundSlateGray,
		accent:     foregroundForestGreen,
		info:       foregroundDeepBlue,
		warning:    foregroundBurntOrange,
		errorColor: foregroundCrimsonRed,
		success:    foregroundSuccessGreen,
	})
	theme.Heading = foregroundDeepGreen
	theme.Selected = foregroundDeepGreen
	theme.Code = foregroundDarkTeal
	theme.Muted = foregroundMutedSlate
	theme.Debug = foregroundMutedSlate
	theme.Placeholder = foregroundMutedSlate
	theme.Disabled = foregroundMutedSlate

	return theme
}

func DefaultFullColorDarkTheme() Theme {
	theme := themeFromPalette(palette{
		text:       foregroundSoftWhite,
		secondary:  foregroundCoolGray,
		accent:     foregroundLavender,
		info:       foregroundSkyBlue,
		warning:    foregroundAmber,
		errorColor: foregroundCoralRed,
		success:    foregroundLightGreen,
	})
	theme.Description = foregroundLightGray
	theme.Code = foregroundPaleCyan
	theme.Debug = foregroundSilverGray

	return theme
}

func DefaultLimitedSixteenColorLightTheme() Theme {
	return themeFromPalette(palette{
		text:       ANSIForegroundBlack,
		secondary:  ANSIForegroundBlack,
		accent:     ANSIForegroundBlue,
		info:       ANSIForegroundBlue,
		warning:    ANSIForegroundBlack,
		errorColor: ANSIForegroundRed,
		success:    ANSIForegroundGreen,
	})
}

func DefaultLimitedSixteenColorDarkTheme() Theme {
	theme := themeFromPalette(palette{
		text:       ANSIForegroundBrightWhite,
		secondary:  ANSIForegroundWhite,
		accent:     ANSIForegroundBrightMagenta,
		info:       ANSIForegroundBrightCyan,
		warning:    ANSIForegroundBrightYellow,
		errorColor: ANSIForegroundBrightRed,
		success:    ANSIForegroundBrightGreen,
	})
	theme.Code = ANSIForegroundCyan

	return theme
}
