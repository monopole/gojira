package set

import (
	"fmt"

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
  Set duration of issues 99 and 300 to two months:

    set duration 2m  99 300
    set duration 8w  99 300
    set duration 60d 99 300

  This sets the "Target Completion Date" of these issues to be
  two months after their start dates.

  If the start date isn't set, it will be initialized to today (` + utils.Today().String() + `).

  Add -` + deltaFlagShort + ` to treat the argument as a delta.

    set duration -` + deltaFlagShort + ` 1w  99      // add one week to existing duration
    set duration -` + deltaFlagShort + ` -- -2d  99  // subtract two days from existing duration
`,
		Args: func(_ *cobra.Command, args []string) (err error) {
			if len(args) < 2 {
				return fmt.Errorf("specify a date and issue number")
			}
			dayCount, err = utils.ConvertToDayCount(args[0])
			if err != nil {
				return err
			}
			if dayCount == 0 {
				return fmt.Errorf("duration must be non-zero")
			}
			if dayCount < 0 && !delta {
				return fmt.Errorf("duration must positive unless using --" + deltaFlag)
			}
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
					start = utils.Today().SlideOverWeekend()
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
	c.Flags().BoolVarP(
		&delta, deltaFlag, deltaFlagShort, false,
		"treat argument as a delta, not an absolute")

	return c
}
