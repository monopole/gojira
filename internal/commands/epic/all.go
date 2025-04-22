package epic

import (
	"github.com/monopole/gojira/internal/myj"
	"github.com/spf13/cobra"
)

func NewEpicCmd(jb *myj.JiraBoss) *cobra.Command {
	c := &cobra.Command{
		Use:          "epic",
		Short:        "Perform operations involving epics",
		SilenceUsage: true,
	}
	c.AddCommand(
		newFixNameCmd(jb),
		newFixDatesCmd(jb),
		newGroupCmd(jb),
		newUnGroupCmd(jb),
		newExportCmd(jb),
		newCalCmd(jb),
		newImportCmd(jb),
		newDotCmd(jb),
	)
	return c
}
