package epic

import (
	"fmt"
	"github.com/monopole/gojira/internal/myj"
	"github.com/monopole/gojira/internal/utils"
	"github.com/spf13/cobra"
)

func newFixNameCmd(jb *myj.JiraBoss) *cobra.Command {
	var epics []int
	c := &cobra.Command{
		Use:   "fix-name",
		Short: `Copy the value in an epic's summary field to its name field`,
		// The name field is stupid, only shows in stories
		// grouped into that epic.
		Args: func(_ *cobra.Command, args []string) (err error) {
			if len(args) == 0 {
				return fmt.Errorf("specify at least one epic number")
			}
			epics, err = utils.ConvertToInt(args)
			return err
		},
		SilenceUsage: true,
		RunE: func(_ *cobra.Command, _ []string) error {
			for i := range epics {
				if err := jb.FixEpicName(epics[i]); err != nil {
					return err
				}
			}
			return nil
		},
	}
	return c
}
