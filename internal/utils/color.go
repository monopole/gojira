package utils

type ColorString string

// Ansi (American National Standards Institute) terminal colors
// see https://gist.github.com/JBlond/2fea43a3049b38287e5e9cefc87b2124
const (
	TerminalReset = "\033[0m"

	//TerminalColorBlack  = "\033[30m"
	TerminalColorRed    = "\033[31m"
	TerminalColorGreen  = "\033[92m"
	TerminalColorYellow = "\033[33m"
	//TerminalColorBlue   = "\033[34m"
	TerminalColorPurple    = "\033[35m"
	TerminalColorCyan      = "\033[36m"
	TerminalColorLightGray = "\033[37m"
	TerminalColorGray      = "\033[1;30m"
	TerminalColorWhite     = "\033[97m"

	//TerminalColorBlackBackground  = "\033[40m"
	//TerminalColorRedBackground    = "\033[41m"
	//TerminalColorGreenBackground  = "\033[42m"
	//TerminalColorYellowBackground = "\033[43m"
	//TerminalColorBlueBackground   = "\033[44m"
	//TerminalColorPurpleBackground = "\033[45m"
	//TerminalColorCyanBackground   = "\033[46m"
	//TerminalColorGrayBackground   = "\033[47m"
	//TerminalColorWhiteBackground  = "\033[107m"
)
