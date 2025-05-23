package utils

import (
	"bytes"
	"fmt"
	"strings"
	"unicode"
)

// DayRange is a specific calendar start day (year, month, day number) paired
// with a dayCount.
// A dayCount less than one is illegal; there must be at least one day.
// Handy for GitHub or jira date range queries.
// https://docs.github.com/en/search-github/getting-started-with-searching-on-github/understanding-the-search-syntax#query-for-dates
type DayRange struct {
	date     Date
	dayCount int
}

// Start returns start date.
func (dr *DayRange) Start() Date {
	return dr.date
}

// End returns end date.
func (dr *DayRange) End() Date {
	return dr.date.AddDays(dr.dayCount - 1)
}

// String is DayRange as date pair splittable on a colon.
func (dr *DayRange) String() string {
	return dr.Start().String() + ":" + dr.End().String()
}

// MakeDayRangeSimple makes an instance of DayRange from the given arguments.
// Fails if dayCount < 1.
func MakeDayRangeSimple(start Date, dayCount int) (*DayRange, error) {
	if dayCount < 1 {
		return nil, fmt.Errorf("daycount of %d is not >= 1", dayCount)
	}
	return &DayRange{date: start, dayCount: dayCount}, nil
}

// MakeDayRangeGentle _always_ provides a valid date range (dayCount > 0),
// but its value only matches the arguments if they make sense.
// When the args don't make sense, the returned error won't be nil, but the
// returned DayRange is still usable as a correction to the arguments.
func MakeDayRangeGentle(start, end Date) (*DayRange, error) {
	const defaultDelta = 4 * 7 // ~one month
	var err error
	if start.IsDefined() {
		if end.IsDefined() {
			if end.Before(start) {
				err = fmt.Errorf(
					"end %s precedes start %s; using end minus ~%d days",
					end, start, defaultDelta)
				start, end = end, start
			} else {
				// This is the only no error path.
			}
		} else {
			err = fmt.Errorf("end undefined, using ~%d days", defaultDelta)
			end = start.AddDays(defaultDelta).SlideBeforeWeekend()
		}
	} else {
		if end.IsDefined() {
			err = fmt.Errorf(
				"bad start date %s; using end minus ~%d days", start, defaultDelta)
			start = end.AddDays(-defaultDelta).SlideOverWeekend()
		} else {
			err = fmt.Errorf(
				"both end and start dates undefined, pushed to one month from now")
			start = Today().AddDays(28).SlideOverWeekend()
			end = start.AddDays(defaultDelta).SlideOffWeekend()
		}
	}
	dayCount := start.DayCount(end)
	if dayCount < 1 {
		// This is an assertion.
		panic("you screwed up repairing the dates")
	}
	return &DayRange{date: start, dayCount: dayCount}, err
}

// MakeRangeFromStringPair makes an instance of DayRange
// from a string like 2006-01-02:2006-01-02
func MakeRangeFromStringPair(arg string) (*DayRange, error) {
	index := strings.Index(arg, ":")
	if index < 0 {
		return nil, fmt.Errorf("no colon")
	}
	start, err := ParseDate(arg[:index])
	if err != nil {
		return nil, err
	}
	end, err := ParseDate(arg[index+1:])
	if err != nil {
		return nil, err
	}
	return MakeDayRangeGentle(start, end)
}

func (dr *DayRange) Contains(d Date) bool {
	start := dr.Start()
	if d == start {
		return true
	}
	end := dr.End()
	if d == end {
		return true
	}
	return d.After(start) && d.Before(end)
}

// StartsBefore is true if the argument strictly starts before this
func (dr *DayRange) StartsBefore(other *DayRange) bool {
	return dr.Start().Before(other.Start())
}

// EndsAfter is true if the argument strictly ends after this
func (dr *DayRange) EndsAfter(other *DayRange) bool {
	return dr.End().After(other.End())
}

// PrettyRange returns a simplified date range as a string.
func (dr *DayRange) PrettyRange() string {
	start := dr.Start()
	end := dr.End()
	weeks := dr.Start().WeekCount(end)
	if start.Year() != end.Year() {
		const f = "Jan 2, 2006"
		return fmt.Sprintf(
			"%s - %s (%dw)",
			start.Format(f), end.Format(f), weeks)
	}
	thisYear := Today().Year()
	if start.Month() != end.Month() {
		const f = "Jan 2"
		if thisYear == start.Year() {
			return fmt.Sprintf(
				"%s - %s (%dw)",
				start.Format(f), end.Format(f), weeks)
		}
		return fmt.Sprintf(
			"%s - %s %d (%dw)",
			start.Format(f), end.Format(f), end.Year(), weeks)
	}
	if start.Day() != end.Day() {
		const f = "Jan 2"
		if thisYear == start.Year() {
			return fmt.Sprintf(
				"%s-%d (%dw)",
				start.Format(f), end.Day(), weeks)
		}
		return fmt.Sprintf(
			"%s-%d %d (%dw)",
			start.Format(f), end.Day(), end.Year(), weeks)
	}
	return fmt.Sprintf("%s (one day)", start.Format("Jan 2, 2006"))
}

// RoundToMondayAndFriday moves the start date to the nearest Monday
// and the end date to the nearest Friday.
func (dr *DayRange) RoundToMondayAndFriday() *DayRange {
	d2, _ := MakeDayRangeGentle(
		dr.Start().SlideOverWeekend().BackToMonday(),
		dr.End().SlideBeforeWeekend().ForwardToFriday())
	return d2
}

const (
	plusSign        = '+' // '⊳'  '⮞'  '►'
	arrowRightSmall = '→' // '▷' '⊳'  '⮞'  '►'
	arrowLeftSmall  = '←' //  '⮜'  '◄'
	arrowRightBig   = '▷' // '⊳'  '⮞'  '►'
	arrowLeftBig    = '◁' // '◁' '◅' '⮜'  '◄'
	heroSpacer      = '⁃' // '‧'
	circleOpen      = '○' // '‧', '○' '⬤'
	circleClosed    = '⬤' // '‧'
	vertBar         = '│'
	hyphen          = '-' // '∙' '─' '-'
	emptySpace      = ' '
	zeroPlaceholder = '_' // Because using 0 is confusing if a digit precedes it
)

type MyBuff struct {
	useColors bool
	tIndex    int
	hero      []rune
	color     ColorString
	bytes.Buffer
}

func (b *MyBuff) incHeroIndex() {
	b.tIndex = (b.tIndex + 1) % len(b.hero)
}

func (b *MyBuff) writeTodaySymbol() {
	if b.useColors {
		b.WriteString(TerminalColorGray)
		b.WriteRune(circleClosed)
		b.WriteString(TerminalReset)
	} else {
		b.WriteRune(plusSign)
	}
}

func (b *MyBuff) writeDaySymbol() {
	if b.useColors {
		b.WriteString(string(b.color))
		b.WriteRune(b.hero[b.tIndex])
		b.WriteString(TerminalReset)
		b.incHeroIndex()
	} else {
		b.WriteRune(hyphen)
	}
}
func (b *MyBuff) writeStartsEarlierSymbol() {
	if b.useColors {
		b.WriteString(string(b.color))
		b.WriteRune(arrowLeftBig)
		b.WriteString(TerminalReset)
	} else {
		b.WriteRune(arrowLeftSmall)
	}
}

func (b *MyBuff) writeEndsLaterSymbol() {
	if b.useColors {
		b.WriteString(string(b.color))
		b.WriteRune(arrowRightBig)
		b.WriteString(TerminalReset)
	} else {
		b.WriteRune(arrowRightSmall)
	}
}

func (b *MyBuff) writeDaySeparator() {
	if b.useColors {
		b.WriteString(TerminalColorGray)
		b.WriteRune(vertBar)
		b.WriteString(TerminalReset)
	} else {
		b.WriteRune(vertBar)
	}
}

func massageHero(s string) []rune {
	const minimum = 5
	result := make([]rune, 0, len(s))
	for _, r := range s {
		if !unicode.IsSpace(r) {
			result = append(result, r)
		}
	}
	padding := func() rune {
		if len(result) == 0 {
			return circleOpen
		}
		return '_'
	}()
	//result = append(result, padding)
	for len(result) < minimum {
		result = append(result, padding)
	}
	return result
}

// AsIntersect accepts an "outer" DayRange and returns a string like "  ---".
//
// Each character in the return represents one business day (Mon-Fri).
//
// The string is roughly the width of the passed in date-range argument, rounded
// down to the nearest Monday and up to the nearest Friday, thus it's width will
// be at least five and will always be a multiple of five, plus the as many
// weekend characters as needed.
//
// If the character is not ' ', then that day lies in both DayRanges.
// If the character is a ' ', then that day is missing from one of the ranges.
//
// If the first character is '<', then there's an intersection to the left
// that's not shown.
//
// If the last character is a '>' then there's an intersection to the right
// that's not shown.
//
// If the character is a '+' (or something special), it's _today_.
func (dr *DayRange) AsIntersect(
	today Date, assignee string,
	outer *DayRange, useColor bool, clr ColorString) string {
	outer = outer.RoundToMondayAndFriday()
	var b MyBuff
	b.hero = massageHero(assignee)
	b.useColors = useColor
	b.color = clr
	outDay := outer.Start().AddDays(-1)
	if dr.StartsBefore(outer) {
		b.writeStartsEarlierSymbol()
	} else {
		b.writeDaySeparator()
	}
	var newWeekend = false
	for i := 0; i < outer.dayCount; i++ {
		outDay = outDay.AddDays(1)
		if outDay.IsWeekend() {
			// This starts false, turns true on saturday,
			// and false again on sunday.
			newWeekend = !newWeekend
			if newWeekend {
				saturday := outDay
				sunday := outDay.AddDays(1)
				if today == saturday || today == sunday {
					b.writeTodaySymbol()
				} else {
					b.writeDaySeparator()
				}
			}
			continue
		}
		if dr.Contains(outDay) {
			b.writeDaySymbol()
		} else {
			if outDay == today {
				b.writeTodaySymbol()
			} else {
				b.WriteByte(emptySpace)
			}
		}
		// b.incHeroIndex()
	}
	if dr.EndsAfter(outer) {
		b.writeEndsLaterSymbol()
	} else {
		b.writeDaySeparator()
	}
	return b.String()
}

func (dr *DayRange) MonthHeader() string {
	outer := dr.RoundToMondayAndFriday()
	var b bytes.Buffer
	b.WriteByte(emptySpace)
	outDay := outer.Start().AddDays(-1)
	prevMonth := outDay.Month()
	var newWeekend = false
	i := 0
	for i < outer.dayCount {
		outDay = outDay.AddDays(1)
		if prevMonth != outDay.Month() {
			monthName := outDay.Month().String()
			b.WriteString(monthName)
			for k := 0; k < len(monthName); k++ {
				i++
				if outDay.IsWeekend() {
					newWeekend = !newWeekend // only 2 days in a weekend
				}
				outDay = outDay.AddDays(1)
			}
			prevMonth = outDay.Month()
		} else {
			if outDay.IsWeekend() {
				newWeekend = !newWeekend // only 2 days in a weekend
				if newWeekend {
					b.WriteByte(emptySpace)
				}
			} else {
				b.WriteByte(emptySpace)
			}
			i++
		}
	}
	return b.String()
}

func (dr *DayRange) DayHeaders() (string, string) {
	outer := dr.RoundToMondayAndFriday()
	var b1, b2 bytes.Buffer
	outDay := outer.Start().AddDays(-1)
	var newWeekend = false
	b1.WriteByte(emptySpace)
	b2.WriteByte(emptySpace)
	var d0 byte
	d0 = 'x'
	for i := 0; i < outer.dayCount; i++ {
		outDay = outDay.AddDays(1)
		if outDay.IsWeekend() {
			newWeekend = !newWeekend // only 2 days in a weekend
			if newWeekend {
				b1.WriteByte(emptySpace)
				b2.WriteByte(emptySpace)
			}
		} else {
			d := fmt.Sprintf("%02d", outDay.Day())
			if d0 != d[0] {
				d0 = d[0]
				if d0 == '0' {
					b1.WriteByte(zeroPlaceholder)
				} else {
					b1.WriteByte(d0)
				}
			} else {
				b1.WriteByte(emptySpace)

			}
			b2.WriteByte(d[1])
		}
	}
	b1.WriteByte(emptySpace)
	b2.WriteByte(emptySpace)
	return b1.String(), b2.String()
}
