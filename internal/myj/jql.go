package myj

import (
	"fmt"
	"strings"

	"github.com/monopole/gojira/internal/utils"
)

func (jb *JiraBoss) JqlEpics() string {
	return andTerms(
		termString("project", RelEqual, jb.Project()),
		termType(RelEqual, IssueTypeEpic),
	)
}

func (jb *JiraBoss) JqlIssues() string {
	return andTerms(
		termString("project", RelEqual, jb.Project()),
		termType(RelNotEqual, IssueTypeEpic),
		termStatus(RelNotEqual, IssueStatusDone),
		termStatus(RelNotEqual, IssueStatusClosedWoAction),
	)
}

func (jb *JiraBoss) JqlIssuesInEpic(epic string) string {
	return andTerms(
		termString("project", RelEqual, jb.Project()),
		termString(CustomFieldEpicLink, RelEqual, epic),
		termStatus(RelNotEqual, IssueStatusDone),
		termStatus(RelNotEqual, IssueStatusClosedWoAction),
	)
}

// Time range queries are tricky - must add one day to the end so that the query
// range counts up through midnight on the end-day.
//
// E.g. suppose https://issues.acmecorp.com/browse/MSFT-001
// was created at 2023-06-08T13:47
//
// The following queries capture that issue
//
//	created > '2023/06/07' and created < '2023/06/09'",
//	created > '2023/06/08' and created < '2023/06/09'",
//
// but this query fails:
//
//	created >= '2023/06/08' and created <= '2023/06/08'

func jqlIssuesCreated(user string, dayRange *utils.DayRange) string {
	// the creator cannot change, but the reporter can change.
	// So maybe use reporter instead of creator?
	// see :  https://support.atlassian.com/jira-software-cloud/docs/jql-fields/
	return fmt.Sprintf(
		"creator = %s and created >= '%s' and created < '%s'",
		user,
		dayRange.Start().JiraFormat(),
		dayRange.End().AddDays(1).JiraFormat())
}

func jqlIssuesCommented(user string, dayRange *utils.DayRange) string {
	return fmt.Sprintf(
		"creator != %s and issuefunction in commented (' by %s after %s') and issuefunction in commented ('by %s before %s')",
		user,
		user,
		dayRange.Start().JiraFormat(),
		user,
		dayRange.End().AddDays(1).JiraFormat(),
	)
}

func jqlIssuesClosed(user string, dayRange *utils.DayRange) string {
	return fmt.Sprintf(
		"status WAS 'Resolved' BY %s DURING ('%s','%s')",
		user,
		dayRange.Start().JiraFormat(),
		dayRange.End().AddDays(1).JiraFormat(),
	)
}

func andTerms(s ...string) string {
	return strings.Join(s, " AND ")
}

func termString(k string, r Rel, s string) string {
	return fmt.Sprintf("%q %s %q", k, r, s)
}

func termType(r Rel, t IssueType) string {
	return fmt.Sprintf("issuetype %s %q", r, t)
}

func termStatus(r Rel, s IssueStatus) string {
	return fmt.Sprintf("status %s %q", r, s)
}
