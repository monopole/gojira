package myj

import "github.com/monopole/gojira/internal/utils"

type ColorKind int

const (
	ColorKindUnknown ColorKind = iota
	ColorKindDot
	ColorKindTerminal
)

func StatusColor(status IssueStatus, kind ColorKind) utils.ColorString {
	switch status {
	case IssueStatusDone, IssueStatusClosed:
		switch kind {
		case ColorKindDot:
			return "lightgreen"
		default:
			return utils.TerminalColorGreen
		}
	case IssueStatusClosedWoAction:
		switch kind {
		case ColorKindDot:
			return "purple"
		default:
			return utils.TerminalColorPurple
		}
	case IssueStatusInProgress:
		switch kind {
		case ColorKindDot:
			return "pink"
		default:
			return utils.TerminalColorRed
		}
	case IssueStatusInQueue:
		switch kind {
		case ColorKindDot:
			return "yellow"
		default:
			return utils.TerminalColorYellow
		}
	default:
		switch kind {
		case ColorKindDot:
			return "white"
		default:
			return utils.TerminalColorLightGray
		}
	}
}
