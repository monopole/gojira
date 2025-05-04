package set

import (
	"fmt"
	"strings"

	"github.com/monopole/gojira/internal/myj"
	"github.com/monopole/gojira/internal/utils"
	"github.com/spf13/cobra"
)

func newDurationCmd(jb *myj.JiraBoss) *cobra.Command {
	const (
		deltaFlag      = "delta"
		deltaFlagShort = "d"
	)
	var (
		issues   []int
		dayCount int
		delta    bool
	)
	c := &cobra.Command{
		Use:   "duration {duration} {issueNum}...",
		Short: `Set the work duration for a set of issues in days, weeks or months`,
		Example: `
  Set duration of issues 99 and 300 to ~two months:

    set duration 2m  99 300
    set duration 8w  99 300
    set duration 60d 99 300

  This sets the '` + myj.CustomFieldTargetCompletionDate +
			`' of these issues to be two
  months after their start dates.  If a start date isn't already
  set, it will be initialized to the next business day after today.

  Prefix with a plus or minus sign to treat the duration as
  a delta to the existing duration:

    set duration +1m  99       // add one month to existing duration
    set duration -- -2w  99    // subtract two weeks from existing duration
`,
		Args: func(_ *cobra.Command, args []string) (err error) {
			if len(args) < 2 {
				return fmt.Errorf("specify a date and issue number")
			}
			argZero := strings.TrimPrefix(args[0], "-")
			sign := 1
			if len(argZero) < len(args[0]) {
				delta = true
				sign = -1
			} else {
				argZero = strings.TrimPrefix(args[0], "+")
				delta = len(argZero) < len(args[0])
			}
			dayCount, err = utils.ConvertToDayCount(argZero)
			if err != nil {
				return err
			}
			if dayCount == 0 {
				return fmt.Errorf("duration must be non-zero")
			}
			dayCount *= sign
			issues, err = utils.ConvertToInt(args[1:])
			return err
		},
		SilenceUsage: true,
		RunE: func(_ *cobra.Command, _ []string) error {
			for _, issue := range issues {
				record, err := jb.GetOneIssue(issue)
				if err != nil {
					return err
				}
				start := record.DateStart()
				if !start.IsDefined() {
					start = utils.Today().AddDays(1).SlideOverWeekend()
				}
				end := record.DateEnd()
				if !end.IsDefined() {
					end = start
				}
				if delta {
					end = end.AddDays(dayCount).SlideOverWeekend()
				} else {
					end = start.AddDays(dayCount).SlideOverWeekend()
				}
				if err = jb.SetDates(issue, start, end); err != nil {
					return err
				}
			}
			return nil
		},
	}
	return c
}
