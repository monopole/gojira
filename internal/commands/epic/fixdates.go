package epic

import (
	"fmt"
	"github.com/monopole/gojira/internal/myj"
	"github.com/spf13/cobra"
	"os"
)

const (
	fixDatesCmd = "fix-dates"
)

func newFixDatesCmd(jb *myj.JiraBoss) *cobra.Command {
	var tighten, doIt bool
	c := &cobra.Command{
		Use:   fixDatesCmd,
		Short: "Fix epic dates",
		Long: `Fix epic dates so that

  - if epic B depends on A, B doesn't start before A ends
  - an epic starts the day after it's tardiest dependency ends

If a new start date lands on a weekend, it slides forward to Monday.

If a new end date lands on a weekend, it slides back to the preceding Friday.
`,
		Args: func(_ *cobra.Command, args []string) error {
			if len(args) > 0 {
				return fmt.Errorf("this command takes no arguments")
			}
			return nil
		},
		SilenceUsage: true,
		RunE: func(_ *cobra.Command, _ []string) error {
			g, err := jb.CreateDiGraph()
			if err != nil {
				return err
			}
			g.ReportNodes(os.Stderr)
			g.ReportMisOrdering(os.Stderr)
			g.ReportWeekends(os.Stderr)
			g.MaybeChangeInMemoryDates(tighten)
			return jb.WriteDates(doIt, g.Nodes())
		},
	}
	c.Flags().BoolVar(&doIt, flagDoIt, false,
		"actually write new dates, rather than just report")
	c.Flags().BoolVar(&tighten, "tighten", false,
		"look for gaps and tighten them")
	return c
}
