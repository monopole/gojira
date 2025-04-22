package epic

import (
	"fmt"
	"os"

	"github.com/monopole/gojira/internal/myj"
	"github.com/monopole/gojira/internal/report"
	"github.com/monopole/gojira/internal/utils"
	"github.com/spf13/cobra"
)

const (
	exportHelp = "Print all epics, optionally with their stories"
	exportCmd  = "export"
)

func newExportCmd(jb *myj.JiraBoss) *cobra.Command {
	var (
		epics      []int
		storiesToo bool
	)
	c := &cobra.Command{
		Use:   exportCmd + " [{epicNum}...]",
		Short: exportHelp,
		Long: exportHelp + `.

Issues (stories, tasks, etc.) that lack an epic, and maybe issues that
refer to an epic outside the given project, are put into a false epic
starting at project-9000.

The output of '` + exportCmd + `' can be read by '` + importCmd + `' to perform
bulk title edits, or bulk re-arrangement of which stories go into which epics.
`,
		SilenceUsage: true,
		Args: func(_ *cobra.Command, args []string) (err error) {
			epics, err = utils.ConvertToInt(args)
			return err
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			var epicMap map[myj.MyKey]*myj.ResponseIssue
			var issueMap map[myj.MyKey]myj.IssueList
			if len(epics) > 0 {
				epicMap = make(map[myj.MyKey]*myj.ResponseIssue)
				for i := range epics {
					issue, err := jb.GetOneIssue(epics[i])
					if err != nil {
						return err
					}
					if issue.Type() != myj.IssueTypeEpic {
						return fmt.Errorf("%d is not an epic", epics[i])
					}
					epicMap[issue.MakeMyKey()] = issue
				}
			} else {
				epicMap = jb.GetEpics()
			}
			if storiesToo {
				issueMap = jb.GetIssuesGroupedByEpic(epicMap)
			}
			report.SpewEpics(
				os.Stdout, epicMap, issueMap, jb.DetermineEpicLink)
			return nil
		},
	}
	c.Flags().BoolVar(&storiesToo, "stories", false, "show stories in epic")
	return c
}
