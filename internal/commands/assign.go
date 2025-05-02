package commands

import (
	"fmt"
	"github.com/monopole/gojira/internal/myj"
	"github.com/monopole/gojira/internal/utils"
	"github.com/spf13/cobra"
)

func newAssignCmd(jb *myj.JiraBoss) *cobra.Command {
	var (
		ldap   string
		issues []int
		remove bool
	)
	c := &cobra.Command{
		Use:   "assign {userName} {issueNum}...",
		Short: "Assign or unassign a user to/from given issues",
		Example: `
To assign issues 100, 200 and 300 to bob:

   assign bob 100 200 300

To remove assignments from those issues:

   assign -r 100 200 300
`,
		Args: func(_ *cobra.Command, args []string) (err error) {
			if remove {
				if len(args) < 1 {
					return fmt.Errorf("specify at least one issue")
				}
			} else {
				if len(args) < 2 {
					return fmt.Errorf("specify a user and at least one issue")
				}
				ldap = args[0]
				args = args[1:]
			}
			issues, err = utils.ConvertToInt(args)
			return err
		},
		SilenceUsage: true,
		RunE: func(_ *cobra.Command, _ []string) (err error) {
			return jb.AssignIssues(issues, ldap)
		},
	}
	c.Flags().BoolVarP(
		&remove, "remove", "r", false, "remove the label instead of add the label")
	return c
}
