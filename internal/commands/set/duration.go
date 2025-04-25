package set

import (
	"fmt"
	"github.com/monopole/gojira/internal/myj"
	"github.com/monopole/gojira/internal/utils"
	"github.com/spf13/cobra"
	"strconv"
)

func newDurationCmd(jb *myj.JiraBoss) *cobra.Command {
	var (
		issue, dayCount int
	)
	c := &cobra.Command{
		Use:   "duration {issue} {duration}",
		Short: `Set the work duration for an issue in days, weeks or months`,
		Example: `
  The following commands are equivalent for setting the 
  duration of issue 99 to 2 months.

    set duration 99 2m
    set duration 99 8w
    set duration 99 60d

  The effect of this is to set the "Target Completion Date" of issue
  99 to be 2 months after its start date.

  If the start date isn't set, it will be initialized to today (` + utils.Today().String() + `).
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
			if len(args) != 1 {
				return fmt.Errorf("need an issue number and a duration")
			}
			dayCount, err = utils.ConvertToDayCount(args[0])
			if err != nil {
				return err
			}
			args = args[1:]
			return nil
		},
		SilenceUsage: true,
		RunE: func(_ *cobra.Command, _ []string) error {
			record, err := jb.GetOneIssue(issue)
			if err != nil {
				return err
			}
			start := record.DateStart()
			if !start.IsDefined() {
				start = utils.Today()
			}
			start = start.SlideOverWeekend()
			return jb.SetDates(
				issue,
				start,
				start.AddDays(dayCount).SlideOffWeekend())
		},
	}
	return c
}
