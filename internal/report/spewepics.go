package report

import (
	"fmt"
	"io"
	"log"
	"sort"

	"github.com/monopole/gojira/internal/myj"
)

// SpewEpics sends text describing epics to the writer.
func SpewEpics(
	w io.Writer,
	epicMap map[myj.MyKey]*myj.ResponseIssue,
	issueMap map[myj.MyKey]myj.IssueList,
	fnc func(*myj.ResponseIssue) myj.MyKey) {
	epicKeys := myj.GetSortedKeys(epicMap)
	if len(epicKeys) == 0 {
		return
	}
	epicKey := epicKeys[0]
	specOneEpic(
		w, epicKey, epicMap[epicKey.MyKey], issueMap[epicKey.MyKey], fnc)
	for i := 1; i < len(epicKeys); i++ {
		epicKey = epicKeys[i]
		if len(issueMap) > 0 {
			_, _ = fmt.Fprintln(w)
		}
		specOneEpic(
			w, epicKey, epicMap[epicKey.MyKey], issueMap[epicKey.MyKey], fnc)
	}
}

func specOneEpic(
	w io.Writer,
	epicKey myj.SrtKey,
	epic *myj.ResponseIssue,
	issues myj.IssueList,
	fnc func(*myj.ResponseIssue) myj.MyKey,
) {
	epic.SpewParsable(w, true, 0)
	sort.Sort(issues)
	for i := range issues {
		issue := issues[i]
		if fnc(issue) != epicKey.MyKey {
			// Sanity check
			log.Printf(
				"Issue %s has %s %s but is grouped into epic %s",
				issue.MyKey,
				myj.CustomFieldEpicLink, fnc(issue), epicKey)
		}
		issue.SpewParsable(w, true, 1)
	}
}
