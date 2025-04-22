package epic

import (
	"fmt"
	"github.com/monopole/gojira/internal/myj"
	"github.com/monopole/gojira/internal/troper"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

const (
	importHelp = `Import a file written by the '` + exportCmd + `' command, but perhaps edited by user`
	importCmd  = "import"
	flagDoIt   = "go"
)

func newImportCmd(jb *myj.JiraBoss) *cobra.Command {
	var doIt bool
	c := &cobra.Command{
		Use:   importCmd + " {fileName}",
		Short: importHelp,
		Long: importHelp + `.

Allows bulk retitling/re-dating/epic-re-organization of issues.

 - Pipe the output of '` + exportCmd + `' command to a file.

 - Manually edit the file.

   - change titles (aka summaries)
   - change types (e.g. Task -> Story)
   - change dates
   - re-arrange story grouping by epic (move indented lines around)

   - at this time one cannot change
       - state (e.g. Backlog -> Done) (use the set state command)
       - labels (use the label command)

 - Use '` + importCmd + `' to apply the changes.

The '` + importCmd + `' checks for errors and won't write without
the additional flag --` + flagDoIt + `.
`,
		Args: func(_ *cobra.Command, args []string) error {
			if len(args) != 1 {
				return fmt.Errorf("specify file path from which to load")
			}
			return nil
		},
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			fs := afero.NewOsFs()
			rawIssues, err := troper.UnSpewEpics(fs, args[0])
			if err != nil {
				return err
			}
			em, im := troper.Convert(rawIssues)
			fmt.Printf("Checking %d epics.\n", len(em))
			if err = jb.CheckEpics(em); err != nil {
				return err
			}
			fmt.Println("Epics look good.")
			fmt.Printf("Checking %d issue lists.\n", len(im))
			if err = jb.CheckIssues(im); err != nil {
				return err
			}
			fmt.Println("Issues look good.")
			if !doIt {
				return fmt.Errorf("add --" + flagDoIt + " to actually perform the write")
			}
			if err = jb.WriteEpics(em); err != nil {
				return err
			}
			if err = jb.WriteIssues(im); err != nil {
				return err
			}
			return nil
		},
	}
	c.Flags().BoolVar(&doIt, flagDoIt, false, "actually write the data")
	return c
}
