package epic

import (
	"fmt"
	"os"

	"github.com/monopole/gojira/internal/myj"
	"github.com/monopole/gojira/internal/report"
	"github.com/monopole/gojira/internal/utils"
	"github.com/spf13/cobra"
)

func newCalCmd(jb *myj.JiraBoss) *cobra.Command {
	var (
		calP        report.CalParams
		flagPrevVal string
	)
	const (
		flagPrevName    = "prev"
		flagPrevDefault = "1m"
		durationDefault = "5m"
	)
	c := &cobra.Command{
		Use:   "cal [duration]",
		Short: "Show epic calendar",
		Example: `
  The following all show the epic calendar for the coming ~6 months:

    cal 6m
    cal 24w
    cal 180d

  To show more of the past, use --` + flagPrevName + `

    cal 6m --` + flagPrevName + ` 2m
   
`,
		SilenceUsage: true,
		Args: func(_ *cobra.Command, args []string) (err error) {
			var (
				dayCount int
				prevDays int
			)
			if len(args) > 1 {
				return fmt.Errorf("just specify a duration")
			}
			duration := durationDefault
			if len(args) > 0 {
				duration = args[0]
			}
			dayCount, err = utils.ConvertToDayCount(duration)
			if err != nil {
				return err
			}
			prevDays, err = utils.ConvertToDayCount(flagPrevVal)
			if err != nil {
				return fmt.Errorf("invalid --%s %s: %w",
					flagPrevName, flagPrevVal, err)
			}
			start := utils.Today().SlideOverWeekend().AddDays(-prevDays)
			calP.Outer, err = utils.MakeDayRange2(start, dayCount+prevDays)
			return err
		},

		RunE: func(cmd *cobra.Command, args []string) error {
			orgEpicMap := jb.GetEpics()
			epicMap := make(map[myj.MyKey]*myj.ResponseIssue)
			for k, v := range orgEpicMap {
				if k.Num < myj.UnknownEpicBase {
					epicMap[k] = v
				}
			}
			err := report.DoCal(os.Stdout, epicMap, calP)
			if err != nil {
				utils.DoErr1(err.Error())
				utils.DoErr1("use the '" + fixDatesCmd + "' command to see and repair errors")
			}
			return nil
		},
	}
	c.Flags().BoolVar(&calP.UseColor, "color", true, "use colors")
	c.Flags().BoolVar(&calP.ShowHeaders, "header", true, "show date headers")
	c.Flags().IntVar(&calP.FieldSizeName, "name-size", 70, "size of name field")
	c.Flags().IntVar(&calP.LineSetSize, "line-set-size", 3, "number of lines in a set")
	c.Flags().StringVar(&flagPrevVal, flagPrevName, flagPrevDefault,
		"number of previous days, weeks, months to show")
	return c
}
