package commands

import (
	"fmt"

	"github.com/monopole/gojira/internal/myj"
	"github.com/monopole/gojira/internal/utils"
	"github.com/spf13/cobra"
)

func newBlockCmd(jb *myj.JiraBoss) *cobra.Command {
	var (
		issues  []int
		comment string
		remove  bool
	)
	c := &cobra.Command{
		Use:   "block {blocker} {blocked} {alsoBlocked}...",
		Short: "Indicate that an issue 'blocks' other issues",
		Example: `
To indicate that issue 99 blocks issues 200, 201 and 202,
i.e. that issue 99 must be completed before the other issues
can be started, enter:

   block 99 200 201 202

To indicate that issue 99 DOES NOT block issues 200, 201 and 202:

   block --remove 99 200 201 202

`,
		Args: func(_ *cobra.Command, args []string) (err error) {
			if len(args) < 2 {
				return fmt.Errorf("specify at least two issues")
			}
			issues, err = utils.ConvertToInt(args)
			return err
		},
		SilenceUsage: true,
		RunE: func(_ *cobra.Command, _ []string) (err error) {
			if remove {
				return jb.UnBlockIssues(issues[0], issues[1:])
			}
			return jb.BlockIssues(issues[0], issues[1:], comment)
		},
	}
	c.Flags().StringVar(&comment, "comment", "", "comment on the blockage")
	c.Flags().BoolVar(&remove, "remove", false, "remove the block instead of add it")
	return c
}
