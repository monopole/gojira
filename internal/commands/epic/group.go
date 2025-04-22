package epic

import (
	"fmt"

	"github.com/monopole/gojira/internal/myj"
	"github.com/monopole/gojira/internal/utils"
	"github.com/spf13/cobra"
)

func newGroupCmd(jb *myj.JiraBoss) *cobra.Command {
	var (
		issues []int
		epic   int
	)
	c := &cobra.Command{
		Use: "group",
		Short: `Set the '` + myj.CustomFieldEpicLink +
			`' field of multiple issues to a common epic number`,
		Example: `
To specify epic 33 as the epic for the issues 111 and 118 enter:

    epic group 33 111 118
`,
		Args: func(_ *cobra.Command, args []string) (err error) {
			if len(args) < 2 {
				return fmt.Errorf(
					"specify an epic number and at least one issue number")
			}
			issues, err = utils.ConvertToInt(args)
			if err != nil {
				return err
			}
			epic = issues[0]
			issues = issues[1:]
			return nil
		},
		SilenceUsage: true,
		RunE: func(_ *cobra.Command, _ []string) (err error) {
			for i := range issues {
				if err = jb.SetEpicLink(issues[i], epic); err != nil {
					return err
				}
			}
			return nil
		},
	}
	return c
}

func newUnGroupCmd(jb *myj.JiraBoss) *cobra.Command {
	var issues []int
	c := &cobra.Command{
		Use: "ungroup",
		Short: `Clear the '` + myj.CustomFieldEpicLink +
			`' field of multiple issues`,
		Example: `
This undoes the work of the 'group' command.

To clear the ` + myj.CustomFieldEpicLink + ` field for issues 111 and 118 enter:

    epic ungroup 111 118
`,
		Args: func(_ *cobra.Command, args []string) (err error) {
			issues, err = utils.ConvertToInt(args)
			return err
		},
		SilenceUsage: true,
		RunE: func(_ *cobra.Command, _ []string) error {
			for i := range issues {
				if err := jb.ClearEpicLink(issues[i]); err != nil {
					return err
				}
			}
			return nil
		},
	}
	return c
}
