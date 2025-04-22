package utils

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
)

const (
	oneDay = 24 * time.Hour

	DayFormatGitHub = "2006-01-02"
	DayFormatHuman  = "2006-Jan-02"
	DayFormatJira   = "2006-01-02"
	DayFormatHuman2 = "2006-Jan-2"
)

func AllDateFormats() []string {
	return []string{
		DayFormatJira, DayFormatHuman, DayFormatHuman2, DayFormatGitHub}
}

func ParseDate(v string) (Date, error) {
	for _, f := range AllDateFormats() {
		if t, err := time.Parse(f, v); err == nil {
			return fromTimeTrunc(t), nil
		}
	}
	// Try prepending the current year
	v = strconv.Itoa(time.Now().Year()) + "-" + v
	for _, f := range AllDateFormats() {
		if t, err := time.Parse(f, v); err == nil {
			return fromTimeTrunc(t), nil
		}
	}
	return Today(), fmt.Errorf(
		"bad date value %q, use formats %s", v, DateOptions())
}

func DateOptions() string {
	opts := AllDateFormats()
	return strings.Join(opts[0:len(opts)-1], ", ") + " or " + opts[len(opts)-1]
}

type Date struct {
	ts time.Time
}

// Today returns a timestamp matching today.
func Today() Date {
	return fromTimeTrunc(time.Now())
}

// GoEpicDate is the Go language zero for dates - easy to recognize in output.
var GoEpicDate = MakeDate(2006, 1, 2)

func MakeDate(y int, m time.Month, d int) Date {
	return Date{ts: time.Date(y, m, d, 0, 0, 0, 0, time.UTC)}
}

// ZeroDate is an invalid empty date used as a placeholder.
var ZeroDate = MakeDate(0, 0, 0)

func (d Date) IsEmpty() bool {
	return d.Equal(ZeroDate)
}

func (d Date) IsGoEpic() bool {
	return d.Equal(GoEpicDate)
}

func (d Date) IsGood() bool {
	return !d.IsEmpty() && !d.IsGoEpic()
}

func (d Date) Year() int {
	return d.ts.Year()
}

func (d Date) Month() time.Month {
	return d.ts.Month()
}

func (d Date) Day() int {
	return d.ts.Day()
}

func (d Date) Weekday() time.Weekday {
	return d.ts.Weekday()
}

// AddDays just adds a day count to the date.  Argument can be negative.
func (d Date) AddDays(dayCount int) Date {
	if dayCount == 0 {
		return d
	}
	return fromTimeTrunc(d.ts.Add(time.Duration(dayCount) * oneDay))
}

// SlideOffWeekend moves this off the weekend if on weekend, either to the
// preceding Friday or the following Monday.
func (d Date) SlideOffWeekend() Date {
	switch d.Weekday() {
	case time.Saturday:
		return d.AddDays(-1)
	case time.Sunday:
		return d.AddDays(1)
	default:
		return d
	}
}

// SlideOverWeekend moves this forward if on weekend over the
// weekend to the following Monday.
func (d Date) SlideOverWeekend() Date {
	switch d.Weekday() {
	case time.Saturday:
		return d.AddDays(2)
	case time.Sunday:
		return d.AddDays(1)
	default:
		return d
	}
}

// SlideBeforeWeekend moves this backward if on weekend to prev Friday.
func (d Date) SlideBeforeWeekend() Date {
	switch d.Weekday() {
	case time.Sunday:
		return d.AddDays(-2)
	case time.Saturday:
		return d.AddDays(-1)
	default:
		return d
	}
}

func (d Date) IsWeekend() bool {
	day := d.Weekday()
	return day == time.Saturday || day == time.Sunday
}

// BackToMonday moves back to nearest Monday, not moving if date already Monday.
func (d Date) BackToMonday() Date {
	day := d.ts
	count := 0
	for day.Weekday() != time.Monday {
		count++
		day = day.Add(-oneDay)
	}
	return d.AddDays(-count)
}

// ForwardToFriday moves forward to nearest Friday.
func (d Date) ForwardToFriday() Date {
	day := d.ts
	count := 0
	for day.Weekday() != time.Friday {
		count++
		day = day.Add(oneDay)
	}
	return d.AddDays(count)
}

// FromTimeRounded moves anything before noon to be the start of that day,
// and anything after noon to be the start of the next day.
func FromTimeRounded(t time.Time) Date {
	return fromTimeTrunc(t.Round(oneDay))
}

// fromTimeTrunc ignores everything but year/mon/day.
func fromTimeTrunc(t time.Time) Date {
	return MakeDate(t.Year(), t.Month(), t.Day())
}

// FromJiraOrDie returns a Date parsed from jira, or dies.
func FromJiraOrDie(f string) Date {
	if f == "" {
		return GoEpicDate
	}
	d, err := ParseDate(f)
	if err != nil {
		panic(err)
	}
	return d
}

// Format uses a time format.
func (d Date) Format(f string) string {
	return d.ts.Format(f)
}

func (d Date) String() string {
	return fmt.Sprintf("%d-%s-%02d", d.ts.Year(), d.ts.Month().String()[:3], d.ts.Day())
}

func (d Date) Brief() string {
	if d.Year() == Today().Year() {
		return fmt.Sprintf("%s-%02d", d.ts.Month().String()[:3], d.ts.Day())
	}
	return fmt.Sprintf("%d-%s-%02d", d.ts.Year(), d.ts.Month().String()[:3], d.ts.Day())
}

func (d Date) JiraFormat() string {
	return fmt.Sprintf(
		"%4d-%02d-%02d", d.ts.Year(), d.ts.Month(), d.ts.Day())
}

// Equal does date level equality.
func (d Date) Equal(other Date) bool {
	return d.Day() == other.Day() &&
		d.Month() == other.Month() &&
		d.Year() == other.Year()
}

// After is true if self is strictly after the argument.
func (d Date) After(other Date) bool {
	if d.Year() == other.Year() {
		if d.Month() == other.Month() {
			return d.Day() > other.Day()
		}
		return d.Month() > other.Month()
	}
	return d.Year() > other.Year()
}

func (d Date) Before(other Date) bool {
	return !d.After(other) && !d.Equal(other)
}

func (d Date) hourCount(end Date) float64 {
	return end.ts.Sub(d.ts).Hours()
}

// DayCount is the number of days from this to end, inclusive
func (d Date) DayCount(end Date) int {
	// Because of a daylight saving time switch,
	// this might not be divisible by 24, so we round it.
	days := int(math.Round(d.hourCount(end) / float64(24)))

	// Add 1 because the dayCount between some day
	// and itself is defined to be 1
	return days + 1
}

// WeekCount is a rounded estimate of weeks;
// 3.4 is 3, 3.5 is 4, but 0 becomes 1.
func (d Date) WeekCount(end Date) int {
	weeks := int(math.Round(float64(d.DayCount(end)) / float64(7)))
	if weeks == 0 {
		return 1
	}
	return weeks
}
