package epic

import (
	"fmt"
	"os"

	"github.com/monopole/gojira/internal/myj"
	"github.com/spf13/cobra"
)

const (
	dotCmd = "dot"
)

func newDotCmd(jb *myj.JiraBoss) *cobra.Command {
	var flagFlip bool
	c := &cobra.Command{
		Use:   dotCmd,
		Short: "Emit dot program instructions to make a digraph of epic dependencies",
		Example: `
  epic ` + dotCmd + ` >k.dot; dot -Tsvg k.dot >k.svg; display k.svg

or

  epic ` + dotCmd + ` | dot -Tsvg | display -

Learn the language at https://graphviz.org/doc/info/lang.html
`,
		Args: func(_ *cobra.Command, args []string) error {
			if len(args) > 0 {
				return fmt.Errorf("this command takes no arguments")
			}
			return nil
		},
		SilenceUsage: true,
		RunE: func(_ *cobra.Command, _ []string) error {
			g, err := jb.CreateDiGraph()
			if err != nil {
				return err
			}
			g.WriteDigraph(os.Stdout, flagFlip)
			g.ReportMisOrdering(os.Stderr)
			g.ReportWeekends(os.Stderr)
			return nil
		},
	}
	c.Flags().BoolVar(&flagFlip, "flip", false,
		"flip the diagram (put end goal at top)")
	return c
}
