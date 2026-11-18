package theme

const (
	NoColor = ""
	reset   = "\x1b[0m"
)

const (
	foregroundDarkSlate    = "\x1b[38;2;28;35;43m"
	foregroundSlateGray    = "\x1b[38;2;75;85;99m"
	foregroundForestGreen  = "\x1b[38;2;21;128;61m"
	foregroundDeepGreen    = "\x1b[38;2;20;83;45m"
	foregroundSuccessGreen = "\x1b[38;2;22;101;52m"
	foregroundDarkTeal     = "\x1b[38;2;17;94;89m"
	foregroundMutedSlate   = "\x1b[38;2;89;98;115m"
	foregroundDeepBlue     = "\x1b[38;2;29;78;216m"
	foregroundBurntOrange  = "\x1b[38;2;194;65;12m"
	foregroundCrimsonRed   = "\x1b[38;2;185;28;28m"
	foregroundSoftWhite    = "\x1b[38;2;243;244;246m"
	foregroundLightGray    = "\x1b[38;2;209;213;219m"
	foregroundCoolGray     = "\x1b[38;2;156;163;175m"
	foregroundSilverGray   = "\x1b[38;2;181;189;204m"
	foregroundLavender     = "\x1b[38;2;193;165;255m"
	foregroundSkyBlue      = "\x1b[38;2;96;165;250m"
	foregroundAmber        = "\x1b[38;2;251;191;36m"
	foregroundCoralRed     = "\x1b[38;2;248;113;113m"
	foregroundLightGreen   = "\x1b[38;2;74;222;128m"
	foregroundPaleCyan     = "\x1b[38;2;165;243;252m"
)

// ANSI foreground and background colors use the terminal's configured palette.
const (
	ANSIForegroundBlack         = "\x1b[30m"
	ANSIForegroundRed           = "\x1b[31m"
	ANSIForegroundGreen         = "\x1b[32m"
	ANSIForegroundYellow        = "\x1b[33m"
	ANSIForegroundBlue          = "\x1b[34m"
	ANSIForegroundMagenta       = "\x1b[35m"
	ANSIForegroundCyan          = "\x1b[36m"
	ANSIForegroundWhite         = "\x1b[37m"
	ANSIForegroundBrightBlack   = "\x1b[90m"
	ANSIForegroundBrightRed     = "\x1b[91m"
	ANSIForegroundBrightGreen   = "\x1b[92m"
	ANSIForegroundBrightYellow  = "\x1b[93m"
	ANSIForegroundBrightBlue    = "\x1b[94m"
	ANSIForegroundBrightMagenta = "\x1b[95m"
	ANSIForegroundBrightCyan    = "\x1b[96m"
	ANSIForegroundBrightWhite   = "\x1b[97m"
)

const (
	ANSIBackgroundBlack         = "\x1b[40m"
	ANSIBackgroundRed           = "\x1b[41m"
	ANSIBackgroundGreen         = "\x1b[42m"
	ANSIBackgroundYellow        = "\x1b[43m"
	ANSIBackgroundBlue          = "\x1b[44m"
	ANSIBackgroundMagenta       = "\x1b[45m"
	ANSIBackgroundCyan          = "\x1b[46m"
	ANSIBackgroundWhite         = "\x1b[47m"
	ANSIBackgroundBrightBlack   = "\x1b[100m"
	ANSIBackgroundBrightRed     = "\x1b[101m"
	ANSIBackgroundBrightGreen   = "\x1b[102m"
	ANSIBackgroundBrightYellow  = "\x1b[103m"
	ANSIBackgroundBrightBlue    = "\x1b[104m"
	ANSIBackgroundBrightMagenta = "\x1b[105m"
	ANSIBackgroundBrightCyan    = "\x1b[106m"
	ANSIBackgroundBrightWhite   = "\x1b[107m"
)
