package utils

import (
	"fmt"
	"strconv"
	"strings"
)

// ConvertToDayCount returns a day count, after parsing a string that
// might have units d (days), w (weeks) or m (months) with WEEKS being the
// default.  Yes, some months don't have 30 days.  This is meant to
// be used in length approximations where weeks or months are the
// appropriate unit, and one day shaved off or added doesn't matter.
// Days come and go as start or end dates are 'rolled off' weekends
// to a workday.
func ConvertToDayCount(s string) (int, error) {
	if strings.HasSuffix(s, "m") {
		m := strings.TrimSuffix(s, "m")
		numMonths, err := strconv.Atoi(m)
		if err != nil {
			return 0, fmt.Errorf("unable to parse %q as months", m)
		}
		return numMonths * 30, nil
	}
	if strings.HasSuffix(s, "d") {
		d := strings.TrimSuffix(s, "d")
		numDays, err := strconv.Atoi(d)
		if err != nil {
			return 0, fmt.Errorf("unable to parse %q as days", d)
		}
		return numDays, nil
	}
	w := strings.TrimSuffix(s, "w") // don't complain if not there.
	numWeeks, err := strconv.Atoi(w)
	if err != nil {
		return 0, fmt.Errorf("unable to parse %q as weeks", s)
	}
	return numWeeks * 7, nil
}
