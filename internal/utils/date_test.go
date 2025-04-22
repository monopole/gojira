package utils

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_DateWeekCount(t *testing.T) {
	const dateStart = "2025-Apr-20"
	type testCase struct {
		end      string
		expected int
	}
	tests := map[string]testCase{
		"t1": {
			end:      "2025-Apr-21",
			expected: 1,
		},
		"t2": {
			end:      "2025-Apr-25",
			expected: 1,
		},
		"t3": {
			end:      "2025-Apr-27",
			expected: 1,
		},
		"t4": {
			end:      "2025-May-16",
			expected: 4,
		},
		"t5": {
			end:      "2025-May-31",
			expected: 6,
		},
		"t6": {
			end:      "2025-Jul-22",
			expected: 13,
		},
		"t7": {
			end:      "2025-Jul-23",
			expected: 14,
		},
		"t8": {
			end:      "2025-Jul-24",
			expected: 14,
		},
	}
	var start, end Date
	var err error
	start, err = ParseDate(dateStart)
	assert.NoError(t, err)
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			end, err = ParseDate(tc.end)
			assert.NoError(t, err)
			assert.Equal(t, tc.expected, start.WeekCount(end))
		})
	}
}

func Test_DateDayCount(t *testing.T) {
	const dateStart = "2025-Mar-04"
	type testCase struct {
		end      string
		expected int
	}
	tests := map[string]testCase{
		"t1": {
			end:      "2025-Mar-04",
			expected: 1,
		},
		"t2": {
			end:      "2025-Mar-05",
			expected: 2,
		},
		"t3": {
			end:      "2025-Mar-06",
			expected: 3,
		},
		"t4": {
			end:      "2025-Mar-16",
			expected: 13,
		},
		"t5": {
			end:      "2025-Mar-31",
			expected: 28,
		},
		"t6": {
			end:      "2025-Apr-21",
			expected: 49,
		},
	}
	var start, end Date
	var err error
	start, err = ParseDate(dateStart)
	assert.NoError(t, err)
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			end, err = ParseDate(tc.end)
			assert.NoError(t, err)
			assert.Equal(t, tc.expected, start.DayCount(end))
		})
	}
}

func Test_DateRoundTrip(t *testing.T) {
	start, err := ParseDate("2025-May-03")
	assert.NoError(t, err)
	end, err := ParseDate("2025-may-4")
	assert.NoError(t, err)
	assert.Equal(t, "2025-May-03", start.String())
	assert.Equal(t, "2025-May-04", end.String())
}

func Test_DateEqual(t *testing.T) {
	start, err := ParseDate("2025-May-03")
	assert.NoError(t, err)
	end, err := ParseDate("2025-May-03")
	assert.NoError(t, err)
	assert.Equal(t, start, end)
	assert.True(t, start == end)
}

func Test_DateAfter(t *testing.T) {
	start, err := ParseDate("2025-May-03")
	assert.NoError(t, err)
	end, err := ParseDate("2025-May-03")
	assert.NoError(t, err)
	assert.False(t, end.After(start))
	assert.False(t, start.After(end))
	end, err = ParseDate("2025-May-04")
	assert.NoError(t, err)
	assert.True(t, end.After(start))
	assert.False(t, start.After(end))
	end, err = ParseDate("2025-Jun-04")
	assert.NoError(t, err)
	assert.True(t, end.After(start))
	assert.False(t, start.After(end))
	end, err = ParseDate("2026-Jun-04")
	assert.NoError(t, err)
	assert.True(t, end.After(start))
	assert.False(t, start.After(end))
}

func Test_FromTimeRounded(t *testing.T) {
	type testCase struct {
		t        time.Time
		expected Date
	}
	tests := map[string]testCase{
		"t1": {
			t:        makeTime(2025, time.December, 20, 11, 59),
			expected: MakeDate(2025, time.December, 20),
		},
		"t2": {
			t:        makeTime(2025, time.December, 20, 12, 01),
			expected: MakeDate(2025, time.December, 21),
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			assert.True(t, tc.expected.Equal(FromTimeRounded(tc.t)))
		})
	}
}

func makeTime(year int, m time.Month, day, hour, min int) time.Time {
	return time.Date(year, m, day, hour, min, 0, 0, time.UTC)
}

func Test_AddDays(t *testing.T) {
	type testCase struct {
		d        Date
		days     int
		expected Date
	}
	tests := map[string]testCase{
		"t0": {
			d:        MakeDate(2025, time.December, 20),
			days:     0,
			expected: MakeDate(2025, time.December, 20),
		},
		"t1": {
			d:        MakeDate(2025, time.December, 20),
			days:     3,
			expected: MakeDate(2025, time.December, 23),
		},
		"leapYear": {
			d:        MakeDate(2024, time.February, 28),
			days:     2,
			expected: MakeDate(2024, time.March, 1),
		},
		"notLeapYear": {
			d:        MakeDate(2025, time.February, 28),
			days:     2,
			expected: MakeDate(2025, time.March, 2),
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			assert.True(t, tc.expected.Equal(tc.d.AddDays(tc.days)))
			// Works backward too.
			assert.True(t, tc.d.Equal(tc.expected.AddDays(-tc.days)))
		})
	}
}

func Test_SlideOffWeekend(t *testing.T) {
	type testCase struct {
		d        Date
		expected Date
	}
	tests := map[string]testCase{
		"mondayStays": {
			d:        MakeDate(2024, time.March, 4),
			expected: MakeDate(2024, time.March, 4),
		},
		"fridayStays": {
			d:        MakeDate(2024, time.March, 1),
			expected: MakeDate(2024, time.March, 1),
		},
		"leapYear": {
			d:        MakeDate(2024, time.March, 2),
			expected: MakeDate(2024, time.March, 1),
		},
		"notLeapYear": {
			d:        MakeDate(2025, time.March, 1),
			expected: MakeDate(2025, time.February, 28),
		},
		"forward": {
			d:        MakeDate(2025, time.March, 2),
			expected: MakeDate(2025, time.March, 3),
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			assert.True(t, tc.expected.Equal(tc.d.SlideOffWeekend()))
		})
	}
}

func Test_BackToMonday(t *testing.T) {
	type testCase struct {
		d        Date
		expected Date
	}
	tests := map[string]testCase{
		"mondayStays": {
			d:        MakeDate(2024, time.March, 4),
			expected: MakeDate(2024, time.March, 4),
		},
		"t2": {
			d:        MakeDate(2024, time.March, 1),
			expected: MakeDate(2024, time.February, 26),
		},
		"leapYear": {
			d:        MakeDate(2024, time.March, 5),
			expected: MakeDate(2024, time.March, 4),
		},
		"notLeapYear": {
			d:        MakeDate(2025, time.March, 2),
			expected: MakeDate(2025, time.February, 24),
		},
		"forward": {
			d:        MakeDate(2025, time.March, 3),
			expected: MakeDate(2025, time.March, 3),
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			assert.True(t, tc.expected.Equal(tc.d.BackToMonday()))
		})
	}
}
