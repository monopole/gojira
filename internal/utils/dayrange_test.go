package utils

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_MakeRangeFromStringPair(t *testing.T) {
	type testCase struct {
		in string
	}
	tests := map[string]testCase{
		"t1": {
			"2025-Apr-08:2025-Apr-10",
		},
		"t2": {
			"2025-Apr-08:2025-May-10",
		},
		"t3": {
			"2025-Apr-08:2025-Dec-30",
		},
		"t4": {
			"2025-Apr-08:2026-Apr-10",
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			dr, err := MakeRangeFromStringPair(tc.in)
			if err != nil {
				t.Error(err.Error())
			}
			assert.Equal(t, tc.in, dr.String())
		})
	}
}

func Test_MakeRangeFromStringPairBad(t *testing.T) {
	type testCase struct {
		in string
	}
	tests := map[string]testCase{
		"t1": {
			"2025-Apr-10:2025-Apr-9",
		},
		"t2": {
			"2025-Apr-082025-May-10",
		},
		"t3": {
			"2025-Apr-08:2023-Dec-30",
		},
		"t4": {
			"2Apr-08:2026-Apr-10",
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			_, err := MakeRangeFromStringPair(tc.in)
			assert.Error(t, err)
		})
	}
}

func Test_dayCountInclusive(t *testing.T) {
	type testCase struct {
		dayStart        string
		dayEnd          string
		desiredDayCount int
	}
	tests := map[string]testCase{
		"t1": {
			dayStart:        "2020-Mar-18",
			dayEnd:          "2020-Mar-18",
			desiredDayCount: 1,
		},
		"t2": {
			dayStart:        "2020-Mar-18",
			dayEnd:          "2020-Mar-19",
			desiredDayCount: 2,
		},
		"t3": {
			dayStart:        "2023-May-31",
			dayEnd:          "2023-Jun-03",
			desiredDayCount: 4,
		},
		"t4": {
			dayStart:        "2023-Jan-01",
			dayEnd:          "2023-Jun-23",
			desiredDayCount: 174,
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t0, err := ParseDate(tc.dayStart)
			if err != nil {
				t.Error(err.Error())
			}
			t1, err := ParseDate(tc.dayEnd)
			if err != nil {
				t.Error(err.Error())
			}
			dc := t0.DayCount(t1)
			if dc != tc.desiredDayCount {
				t.Fatalf("expected dc=%d, got %d", tc.desiredDayCount, dc)
			}
		})
	}
}

func Test_RoundToMondayOrFriday(t *testing.T) {
	type testCase struct {
		r        string
		expected string
	}
	//   2025  Su  Mo  Tu  We  Th  Fr  Sa
	//    Mar  23  24  25  26  27  28 [29]
	//    Apr  30  31   1   2   3   4   5
	//          6   7   8   9  10  11  12
	//         13  14  15  16  17  18  19
	//         20  21  22  23  24  25  26
	//    May  27  28  29  30   1   2   3
	tests := map[string]testCase{
		"t1": {
			r:        "2025-Apr-08:2025-Apr-10",
			expected: "2025-Apr-07:2025-Apr-11",
		},
		"t2": {
			r:        "2025-Apr-07:2025-Apr-11",
			expected: "2025-Apr-07:2025-Apr-11",
		},
		"t3": {
			r:        "2025-Apr-06:2025-Apr-12",
			expected: "2025-Apr-07:2025-Apr-11",
		},
		"t4": {
			r:        "2025-Apr-05:2025-Apr-13",
			expected: "2025-Apr-07:2025-Apr-11",
		},
		"t5": {
			r:        "2025-Apr-09:2025-Apr-09",
			expected: "2025-Apr-07:2025-Apr-11",
		},
		"t6": {
			r:        "2025-Mar-30:2025-Apr-24",
			expected: "2025-Mar-31:2025-Apr-25",
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			dr, err := MakeRangeFromStringPair(tc.r)
			if err != nil {
				t.Fatal(err.Error())
			}
			assert.Equal(t, tc.expected, dr.RoundToMondayAndFriday().String())
		})
	}
}

func Test_AsIntersect(t *testing.T) {
	type testCase struct {
		outer    string
		inner    string
		header0  string
		header1  string
		header2  string
		expected string
	}
	//   2025  Su  Mo  Tu  We  Th  Fr  Sa
	//    Mar  23  24  25  26  27  28 [29]
	//    Apr  30  31   1   2   3   4   5   *
	//          6   7   8   9  10  11  12   *
	//         13  14  15  16  17  18  19   *
	//         20  21  22  23  24  25  26
	//    May  27  28  29  30   1   2   3
	tests := map[string]testCase{
		"t1": {
			outer:    "2025-Mar-30:2025-Apr-24",
			inner:    "2025-Apr-06:2025-Apr-12",
			header0:  "  April                 ",
			header1:  " 3_       1        2     ",
			header2:  " 11234 78901 45678 12345 ",
			expected: "│     │-----│     │     │",
		},
		"t2": {
			outer:    "2025-Mar-30:2025-Apr-24",
			inner:    "2025-Apr-01:2025-Apr-12",
			header0:  "  April                 ",
			header1:  " 3_       1        2     ",
			header2:  " 11234 78901 45678 12345 ",
			expected: "│ ----│-----│     │     │",
		},
		"t3": {
			outer:    "2025-Mar-30:2025-Apr-24",
			inner:    "2025-Apr-14:2025-Apr-25",
			header0:  "  April                 ",
			header1:  " 3_       1        2     ",
			header2:  " 11234 78901 45678 12345 ",
			expected: "│     │     │-----│-----│",
		},
		"t4": {
			outer:    "2025-Mar-30:2025-Apr-24",
			inner:    "2025-Apr-01:2025-Apr-24",
			header0:  "  April                 ",
			header1:  " 3_       1        2     ",
			header2:  " 11234 78901 45678 12345 ",
			expected: "│ ----│-----│-----│---- │",
		},
		"t5": {
			outer:    "2025-Mar-30:2025-Apr-24",
			inner:    "2025-Mar-15:2025-Apr-12",
			header0:  "  April                 ",
			header1:  " 3_       1        2     ",
			header2:  " 11234 78901 45678 12345 ",
			expected: "←-----│-----│     │     │",
		},
		"t6": {
			outer:    "2025-Mar-30:2025-Apr-24",
			inner:    "2025-Apr-14:2025-Apr-30",
			header0:  "  April                 ",
			header1:  " 3_       1        2     ",
			header2:  " 11234 78901 45678 12345 ",
			expected: "│     │     │-----│-----→",
		},
		"t7": {
			outer:    "2025-Mar-30:2025-May-01",
			inner:    "2025-Apr-14:2025-Apr-25",
			header0:  "  April                     May",
			header1:  " 3_       1        2       3_  ",
			header2:  " 11234 78901 45678 12345 89012 ",
			expected: "│     │     │-----│-----│     │",
		},
		"t8": {
			outer:    "2025-Mar-23:2025-May-01",
			inner:    "2025-Apr-14:2025-Apr-24",
			header0:  "        April                     May",
			header1:  " 2     3_       1        2       3_  ",
			header2:  " 45678 11234 78901 45678 12345 89012 ",
			expected: "│     │     │     │-----│---- │     │",
		},
	}
	today, err := ParseDate("2024-Apr-10")
	if err != nil {
		t.Fatal(err.Error())
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			outer, err := MakeRangeFromStringPair(tc.outer)
			if err != nil {
				t.Fatal(err.Error())
			}
			inner, err := MakeRangeFromStringPair(tc.inner)
			if err != nil {
				t.Fatal(err.Error())
			}
			x := inner.AsIntersect(today, "", outer, false, "")
			h1, h2 := outer.DayHeaders()
			assert.Equal(t, tc.expected, x)
			assert.Equal(t, tc.header0, outer.MonthHeader())
			assert.Equal(t, tc.header1, h1)
			assert.Equal(t, tc.header2, h2)
		})
	}
}

func TestDayRange_MakeDayRange(t *testing.T) {
	type testCase struct {
		dayStart string
		dayEnd   string
		dayCount int
		expEnd   Date
		want     string
	}
	tests := map[string]testCase{
		"t1": {
			dayStart: "2020-Mar-18",
			dayCount: 1,
			// A dayCount of one means that start == end.
			expEnd: MakeDate(2020, time.March, 18),
			want:   "Mar 18, 2020 (one day)",
		},
		"t2": {
			dayStart: "2020-Mar-01",
			dayCount: 1,
			expEnd:   MakeDate(2020, time.March, 1),
			want:     "Mar 1, 2020 (one day)",
		},
		"t3": {
			dayStart: "2020-03-01",
			dayCount: 1,
			expEnd:   MakeDate(2020, time.March, 1),
			want:     "Mar 1, 2020 (one day)",
		},
		"t3a": {
			dayStart: "2020-03-01",
			dayCount: 4,
			expEnd:   MakeDate(2020, time.March, 4),
			want:     "Mar 1-4 2020 (1w)",
		},
		"t4": {
			dayStart: "2020-Mar-30",
			dayCount: 5, // March has 31 days
			expEnd:   MakeDate(2020, time.April, 3),
			want:     "Mar 30 - Apr 3 2020 (1w)",
		},
		"t5": {
			dayStart: "2020-Mar-30",
			dayEnd:   "2020-Apr-03",
			expEnd:   MakeDate(2020, time.April, 3),
			want:     "Mar 30 - Apr 3 2020 (1w)",
		},
		"t6": {
			dayStart: "2020-Dec-30",
			dayCount: 5, // December has 31 days
			expEnd:   MakeDate(2021, time.January, 3),
			want:     "Dec 30, 2020 - Jan 3, 2021 (1w)",
		},
		"t7": {
			dayStart: "2020-Dec-30",
			dayEnd:   "2021-Jan-03",
			expEnd:   MakeDate(2021, time.January, 3),
			want:     "Dec 30, 2020 - Jan 3, 2021 (1w)",
		},
		"t8": {
			dayStart: "2023-Jan-01",
			dayEnd:   "2023-Jun-23",
			expEnd:   MakeDate(2023, time.June, 23),
			want:     "Jan 1 - Jun 23 2023 (25w)",
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			var (
				dr  *DayRange
				err error
			)
			if tc.dayEnd != "" {
				dr, err = MakeRangeFromStringPair(tc.dayStart + ":" + tc.dayEnd)
				assert.NoError(t, err)
			} else {
				var start Date
				start, err = ParseDate(tc.dayStart)
				assert.NoError(t, err)
				dr, err = MakeDayRangeSimple(start, tc.dayCount)
				assert.NoError(t, err)
			}
			got := dr.PrettyRange()
			if got != tc.want {
				t.Errorf("MakeDayRange() = %v, want %v", got, tc.want)
			}
			assert.True(t, tc.expEnd.Equal(dr.End()))
		})
	}
}

func TestDayRange_StartAsTime(t *testing.T) {
	type testCase struct {
		y    int
		m    time.Month
		d    int
		want string
	}
	tests := map[string]testCase{
		"t1": {
			y:    2020,
			m:    3,
			d:    18,
			want: "2020-03-18",
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			dr := &DayRange{
				date: MakeDate(tc.y, tc.m, tc.d),
			}
			if got := dr.Start().JiraFormat(); got != tc.want {
				t.Errorf("StartAsTime() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestDayRange_EndAsTime(t *testing.T) {
	type testCase struct {
		y        int
		m        time.Month
		d        int
		dayCount int
		want     string
	}
	tests := map[string]testCase{
		"t1": {
			y:        2020,
			m:        3,
			d:        18,
			dayCount: 1,
			want:     "2020-03-18",
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			dr := &DayRange{
				date:     MakeDate(tc.y, tc.m, tc.d),
				dayCount: tc.dayCount,
			}
			if got := dr.Start().JiraFormat(); got != tc.want {
				t.Errorf("EndAsTime() = %v, want %v", got, tc.want)
			}
		})
	}
}
