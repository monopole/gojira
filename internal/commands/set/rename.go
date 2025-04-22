package set

import (
	"fmt"
	"strconv"

	"github.com/monopole/gojira/internal/myj"
	"github.com/spf13/cobra"
)

func newNameCmd(jb *myj.JiraBoss) *cobra.Command {
	var issue int
	var name string
	c := &cobra.Command{
		Use:   "name {number} \"the new name\"",
		Short: `Set a new name for an issue`,
		Args: func(_ *cobra.Command, args []string) (err error) {
			if len(args) < 2 {
				return fmt.Errorf(
					"specify issue number and its new name")
			}
			issue, err = strconv.Atoi(args[0])
			if err != nil {
				return err
			}
			name = args[1]
			return nil
		},
		SilenceUsage: true,
		RunE: func(_ *cobra.Command, _ []string) error {
			return jb.RenameIssue(issue, name)
		},
	}
	return c
}
