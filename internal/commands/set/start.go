package set

import (
	"fmt"
	"strconv"
	"time"

	"github.com/monopole/gojira/internal/myj"
	"github.com/monopole/gojira/internal/utils"
	"github.com/spf13/cobra"
)

func newStartCmd(jb *myj.JiraBoss) *cobra.Command {
	var (
		issue int
		start utils.Date
	)
	const defaultWeeks = 4
	c := &cobra.Command{
		Use:   "start {issue} {date}",
		Short: `Set the work start date for some issue`,
		Example: `
  To start issue (story, epic, etc.) 99 on april 1, 2025:

    set start 99 2025-apr-1

  An omitted year defaults to current year, e.g 'apr-1'' is sufficient.
  If the date is entirely omitted, today (` + utils.Today().String() + `) is used.

  The existing _end_ date will be shifted to keep the duration the same.

  If there is no existing end date, it will be set to establish
  a duration of ` + strconv.Itoa(defaultWeeks) + ` weeks.
  One can change this with 'set duration'.
`,
		Args: func(_ *cobra.Command, args []string) (err error) {
			if len(args) < 1 {
				return fmt.Errorf("specify issue number")
			}
			issue, err = strconv.Atoi(args[0])
			if err != nil {
				return err
			}
			args = args[1:]
			if len(args) > 1 {
				return fmt.Errorf("just want an issue number and a date")
			}
			if len(args) > 0 {
				start, err = utils.ParseDate(args[0])
				if err != nil {
					// Try prepending the current year
					d := strconv.Itoa(time.Now().Year()) + "-" + args[0]
					start, err = utils.ParseDate(d)
					if err != nil {
						return fmt.Errorf("bad date %q", d)
					}
				}
				args = args[1:]
			} else {
				start = utils.Today()
			}
			start = start.SlideOverWeekend()
			return nil
		},
		SilenceUsage: true,
		RunE: func(_ *cobra.Command, _ []string) error {
			record, err := jb.GetOneIssue(issue)
			if err != nil {
				return err
			}
			oldStart := record.DateStart()
			oldEnd := record.DateEnd()
			if !oldStart.IsDefined() || !oldEnd.IsDefined() {
				// If either of the old dates is bad, we don't have a valid
				// duration, so just use the default duration.
				return jb.SetDates(
					issue,
					start,
					start.AddDays(defaultWeeks*7).SlideOffWeekend())
			}
			// Roughly preserve the existing duration.
			return jb.SetDates(
				issue,
				start,
				start.AddDays(oldStart.DayCount(oldEnd)).SlideOffWeekend())
		},
	}
	return c
}
