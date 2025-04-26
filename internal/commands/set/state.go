package set

import (
	"fmt"
	"github.com/monopole/gojira/internal/myj"
	"github.com/monopole/gojira/internal/utils"
	"github.com/spf13/cobra"
)

func newStateCmd(jb *myj.JiraBoss) *cobra.Command {
	var issues []int
	var status myj.IssueStatus
	c := &cobra.Command{
		Use:   "state {state} {issueNum}...",
		Short: "Move the given issues to a new state",
		Example: `
   set status "In Queue" 12 33 45 

   set status Done 100 200 300
`,
		Args: func(_ *cobra.Command, args []string) (err error) {
			if len(args) < 2 {
				return fmt.Errorf(
					"specify new state in quotes and issue number(s)")
			}
			status, err = myj.IssueStatusString(args[0])
			if err != nil {
				return err
			}
			issues, err = utils.ConvertToInt(args[1:])
			return err
		},
		SilenceUsage: true,
		RunE: func(_ *cobra.Command, _ []string) error {
			for _, issue := range issues {
				id, err := jb.GetTransitionId(issue, status)
				if err != nil {
					return err
				}
				err = jb.MoveIssueToState(issue, id)
				if err != nil {
					return err
				}
			}
			return nil
		},
	}
	return c
}
