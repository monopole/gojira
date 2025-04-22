package set

import (
	"github.com/monopole/gojira/internal/myj"
	"github.com/spf13/cobra"
)

func NewSetCmd(jb *myj.JiraBoss) *cobra.Command {
	c := &cobra.Command{
		Use:          "set",
		Short:        "Set properties of issues",
		SilenceUsage: true,
	}
	c.AddCommand(
		newDurationCmd(jb),
		newStateCmd(jb),
		newNameCmd(jb),
		newStartCmd(jb),
	)
	return c
}
