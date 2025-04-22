package commands

import (
	"fmt"

	"github.com/monopole/gojira/internal/myj"
	"github.com/spf13/cobra"
)

func newFieldCmd(jb *myj.JiraBoss) *cobra.Command {
	c := &cobra.Command{
		Use:   "field <name1> <name2>",
		Short: "Discover internal names of the given jira API fields",
		Long: `Discover internal names of the given jira API fields

This is here to document how to find the names of custom fields via the API.
e.g. one cannot write the field ` + myj.CustomFieldEpicLink + ` without first
discovering that its name inside Jira is "customfield_12003".
`,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			args = append(
				// examples
				[]string{
					myj.CustomFieldEpicLink,
					myj.CustomFieldEpicName,
					myj.CustomFieldStartDate,
					myj.CustomFieldTargetCompletionDate,
					"Resolution",
					"Status"},
				args...)
			for i := range args {
				reportCustomField(jb, args[i])
			}
			return nil
		},
	}
	return c
}

func reportCustomField(jb *myj.JiraBoss, name string) {
	fmt.Printf("Custom field  %30s = %s\n", name, jb.GetCustomFieldId(name))
}
