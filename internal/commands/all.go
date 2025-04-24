package commands

import (
	"fmt"
	"os"

	"github.com/monopole/gojira/internal/commands/epic"
	"github.com/monopole/gojira/internal/commands/set"
	"github.com/monopole/gojira/internal/myhttp"
	"github.com/monopole/gojira/internal/myj"
	"github.com/monopole/gojira/internal/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

const (
	envJiraToken   = "JIRA_API_TOKEN"
	envJiraHost    = "JIRA_HOST"
	envJiraProject = "JIRA_PROJECT"
	flagJiraToken  = "jira-token"
)

func NewGoJiraCommand() *cobra.Command {
	var (
		// caPath holds the part to a CA cert file for server authentication.
		caPath   string
		jiraArgs myj.MyJiraArgs
		jb       myj.JiraBoss
	)
	c := &cobra.Command{
		Use:          "gojira",
		Short:        "Manipulate jira issues - read, write, create reports, etc.",
		SilenceUsage: true,
		Example: `
  View epics as a directed graph (via https://graphviz.org/docs/layouts/dot):

    gojira epic dot | dot -Tsvg | display -

  View epics as a six month wide calendar on the terminal

    gojira epic cal 6m

  Export, edit and import epics and stories:

    gojira epic export --stories > file.txt

    // Edit epic and story titles and dates as desired.
    // Move story lines to re-arrange stories between epics.
    vi file.text

    // Apply your changes (it does error checking first):
    gojira epic import file.txt`,
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			if cmd.Name() == "help" {
				return nil
			}
			if err := validateJiraArgs(&jiraArgs); err != nil {
				return err
			}
			htCl, err := myhttp.MakeHttpClient(caPath)
			if err != nil {
				return err
			}
			jb = myj.MakeJiraBoss(htCl, &jiraArgs)
			return nil
		},
	}
	c.AddCommand(
		set.NewSetCmd(&jb),
		newLabelCmd(&jb),
		newFieldCmd(&jb),
		epic.NewEpicCmd(&jb),
		newPrintCmd(&jb),
		newBlockCmd(&jb),
	)
	func(set *pflag.FlagSet) {
		set.StringVar(&jiraArgs.Project, "jira-project", "",
			fmt.Sprintf("jira project (overrides $%s)", envJiraProject))
		set.StringVar(&jiraArgs.Host, "jira-host", "",
			fmt.Sprintf("jira host (overrides $%s)", envJiraHost))
		set.StringVar(&jiraArgs.Token, flagJiraToken, "",
			fmt.Sprintf("access token for the given jira host (overrides $%s)",
				envJiraToken))
	}(c.PersistentFlags())

	utils.FlagsAddDebug(c.PersistentFlags())
	c.PersistentFlags().StringVar(
		&caPath, "ca-path", "", "local path to CA cert file for TLS checking")
	return c

}

func validateJiraArgs(args *myj.MyJiraArgs) error {
	if args.Host == "" {
		args.Host = os.Getenv(envJiraHost)
	}
	if args.Host == "" {
		return fmt.Errorf(
			"set env var %q to point to a jira API host", envJiraHost)
	}
	if args.Project == "" {
		args.Project = os.Getenv(envJiraProject)
	}
	if args.Project == "" {
		return fmt.Errorf(
			"set env var %q to specify a jira project", envJiraProject)
	}
	if args.Token == "" {
		args.Token = os.Getenv(envJiraToken)
	}
	if args.Token == "" {
		return fmt.Errorf(`
Set env var %s to a personal access token value obtained from
  https://%s/secure/ViewProfile.jspa?%s
`,
			envJiraToken,
			args.Host,
			"selectedTab=com.atlassian.pats.pats-plugin:myjira-user-personal-access-tokens",
		)
	}
	return nil
}
