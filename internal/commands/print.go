package commands

import (
	"os"

	"github.com/monopole/gojira/internal/myj"
	"github.com/monopole/gojira/internal/utils"
	"github.com/spf13/cobra"
)

func newPrintCmd(jb *myj.JiraBoss) *cobra.Command {
	var issues []int
	c := &cobra.Command{
		Use:   "print",
		Short: "Print information about the given issues",
		Args: func(_ *cobra.Command, args []string) (err error) {
			issues, err = utils.ConvertToInt(args)
			return err
		},
		SilenceUsage: true,
		RunE: func(_ *cobra.Command, _ []string) (err error) {
			var issue *myj.ResponseIssue
			for i := range issues {
				issue, err = jb.GetOneIssue(issues[i])
				if err != nil {
					return err
				}
				issue.SpewParsable(os.Stdout, false, 0)
			}
			return nil
		},
	}
	return c
}
