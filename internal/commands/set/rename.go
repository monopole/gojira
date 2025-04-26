package set

import (
	"fmt"
	"github.com/monopole/gojira/internal/myj"
	"github.com/spf13/cobra"
	"strconv"
)

const setNameHelp = `Set a new name (a.k.a. 'summary') for an issue`

func newNameCmd(jb *myj.JiraBoss) *cobra.Command {
	var issue int
	var name string
	c := &cobra.Command{
		Use:   "name \"The new name\" {issueNum} ",
		Short: setNameHelp,
		Long: setNameHelp + `
Use quotes around the new name.`,
		Args: func(_ *cobra.Command, args []string) (err error) {
			if len(args) != 2 {
				return fmt.Errorf(
					"specify the name in quotas and the issue number")
			}
			name = args[0]
			issue, err = strconv.Atoi(args[1])
			return err
		},
		SilenceUsage: true,
		RunE: func(_ *cobra.Command, _ []string) error {
			return jb.RenameIssue(issue, name)
		},
	}
	return c
}
