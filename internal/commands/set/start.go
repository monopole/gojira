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
		issues []int
		start  utils.Date
	)
	const defaultWeeks = 4
	c := &cobra.Command{
		Use:   "start {date} {issueNum}...",
		Short: `Set the work start date for a set of issues`,
		Example: `
  To start issue (story, epic, etc.) 99 and 300 on April 1:

    set start apr-1 99 300

  Put the year in front if you want a year other than the current year, e.g.

    set start 2026-jan-1 99 300 

  The existing end date will be shifted to keep the duration the same.
  If there is no existing end date, it will be set to establish
  a duration of ` + strconv.Itoa(defaultWeeks) + ` weeks.
  One can change this with 'set duration'.
`,
		Args: func(_ *cobra.Command, args []string) (err error) {
			if len(args) < 2 {
				return fmt.Errorf("specify a date and issue number")
			}
			start, err = utils.ParseDate(args[0])
			if err != nil {
				// Try prepending the current year
				d := strconv.Itoa(time.Now().Year()) + "-" + args[0]
				start, err = utils.ParseDate(d)
				if err != nil {
					return fmt.Errorf("bad date %q", d)
				}
			}
			start = start.SlideOverWeekend()
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
				durationDays := func() int {
					oldStart := record.DateStart()
					oldEnd := record.DateEnd()
					if oldStart.IsDefined() && oldEnd.IsDefined() {
						return oldStart.DayCount(oldEnd)
					}
					return defaultWeeks * 7
				}()
				err = jb.SetDates(
					issue,
					start,
					start.AddDays(durationDays).SlideOffWeekend())
				if err != nil {
					return err
				}
			}
			return nil
		},
	}
	return c
}
