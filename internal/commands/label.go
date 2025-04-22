package commands

import (
	"fmt"
	"github.com/monopole/gojira/internal/myj"
	"github.com/monopole/gojira/internal/utils"
	"github.com/spf13/cobra"
)

func newLabelCmd(jb *myj.JiraBoss) *cobra.Command {
	var (
		label  string
		issues []int
		remove bool
	)
	c := &cobra.Command{
		Use:   "label {label} {issueNum}...",
		Short: "Add/remove a label to/from the given issues",
		Example: `
To add the label 'critical' to issues 100, 200 and 300:

   label critical 100 200 300

To remove it:

   label --remove critical 100 200 300
`,
		Args: func(_ *cobra.Command, args []string) (err error) {
			if len(args) < 2 {
				return fmt.Errorf("specify at least a label and one issue")
			}
			label = args[0]
			issues, err = utils.ConvertToInt(args[1:])
			return err
		},
		SilenceUsage: true,
		RunE: func(_ *cobra.Command, _ []string) (err error) {
			return jb.LabelIssues(label, issues, remove)
		},
	}
	c.Flags().BoolVar(
		&remove, "remove", false, "remove the label instead of add the label")
	return c
}
