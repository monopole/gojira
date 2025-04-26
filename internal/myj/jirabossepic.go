package myj

import (
	"fmt"
	"github.com/monopole/gojira/internal/utils"
	"log"
	"net/http"
	"sort"
)

// GetEpics returns a map of issue keys (e.g. project-100) to issue structs,
// where all the issues happen to be epics.
func (jb *JiraBoss) GetEpics() (result map[MyKey]*ResponseIssue) {
	epics, err := jb.DoPagedSearch(jb.jqlEpics())
	if err != nil {
		log.Fatal(err)
	}
	return jb.makeEpicMap(epics, false)
}

// GetEpicsWithPlaceholder returns GetEpics plus a known placeholder
// to accumulate orphan stories.
func (jb *JiraBoss) GetEpicsWithPlaceholder() (result map[MyKey]*ResponseIssue) {
	epics, err := jb.DoPagedSearch(jb.jqlEpics())
	if err != nil {
		log.Fatal(err)
	}
	return jb.makeEpicMap(epics, true)
}

func (jb *JiraBoss) makeEpicMap(
	found []ResponseIssue, placeHold bool) (result map[MyKey]*ResponseIssue) {
	result = make(map[MyKey]*ResponseIssue)
	if placeHold {
		result[jb.placeholderEpic.MakeMyKey()] = jb.placeholderEpic
	}
	for i := range found {
		key := found[i].MakeMyKey()
		_, ok := result[key]
		if ok {
			log.Fatal(fmt.Errorf("epic %s appears more than once", key))
		}
		result[key] = &found[i]
	}
	return
}

func (jb *JiraBoss) DetermineEpicLink(ir *ResponseIssue) (result MyKey) {
	str, ok := ir.Fields.CustomEpicLink.(string)
	if ok && str != "" {
		return ParseMyKey(str)
	}
	return jb.placeholderEpic.MyKey
}

// GetIssuesGroupedByEpic returns a map of epic keys
// (issues that happen to be epics),
// to lists of issues that are in that epic.
func (jb *JiraBoss) GetIssuesGroupedByEpic(epics map[MyKey]*ResponseIssue) (
	result map[MyKey]IssueList) {
	issues, err := jb.DoPagedSearch(jb.jqlIssues())
	if err != nil {
		log.Fatal(err)
	}
	result = make(map[MyKey]IssueList)
	for i := range issues {
		issue := issues[i]
		epicKey := jb.DetermineEpicLink(&issue)
		if _, ok := epics[epicKey]; !ok {
			// Found an issue that points to an unknown epic.
			// Most likely outside the project.
			// Look it up so we can print it.
			var epic *ResponseIssue
			epic, err = jb.GetOneIssue(epicKey.Num)
			if err != nil {
				epic = jb.incrementUnknownEpic()
			}
			epic.MyKey = epicKey
			epics[epicKey] = epic
		}
		result[epicKey] = append(result[epicKey], &issue)
	}
	for _, v := range result {
		sort.Sort(v)
	}
	return
}

// SetEpicLink PUTs an issue to modify the epic link.
func (jb *JiraBoss) SetEpicLink(issue int, epic int) (err error) {
	type fieldsToWrite struct {
		issueOnlyFields
	}
	type requestPutIssue struct {
		Fields fieldsToWrite `json:"fields"`
	}
	var req requestPutIssue
	req.Fields.CustomEpicLink = jb.Key(epic).String()
	_, err = jb.punchItChewie(
		http.MethodPut, req, endpointIssue+"/"+jb.Key(issue).String())
	return err
}

// ClearEpicLink PUTS an issue to clear the CustomFieldEpicLink.
func (jb *JiraBoss) ClearEpicLink(issue int) (err error) {
	type fieldsToWrite struct {
		issueOnlyFields
		CommonIssueAndEpicFields
	}
	type requestPutIssue struct {
		Fields fieldsToWrite `json:"fields"`
	}
	var req requestPutIssue
	req.Fields.CustomEpicLink = nil
	_, err = jb.punchItChewie(
		http.MethodPut, req, endpointIssue+"/"+jb.Key(issue).String())
	return err
}

// FixEpicName gets the epic, reads the string value in the summary field
// and writes that value to the custom name field so that they match.
func (jb *JiraBoss) FixEpicName(epic int) error {
	r, err := jb.GetOneIssue(epic)
	if err != nil {
		return err
	}
	if r.Fields.CustomEpicName == r.Fields.Summary {
		// _, _ = fmt.Fprintln(os.Stderr, "names match")
		// Nothing to do
		return nil
	}
	type requestPutIssue struct {
		Fields basicEpicFields `json:"fields"`
	}
	req := requestPutIssue{
		Fields: basicEpicFields{
			epicOnlyFields: epicOnlyFields{
				CustomEpicName: r.Fields.Summary,
			},
		},
	}
	_, err = jb.punchItChewie(
		http.MethodPut, &req, endpointIssue+"/"+jb.Key(epic).String())
	return err
}

func (jb *JiraBoss) CheckEpics(em map[MyKey]*ResponseIssue) error {
	foundLookupError := false
	foundEpicError := false
	for epicKey, epic := range em {
		var (
			resp *ResponseIssue
			err  error
		)
		resp, err = jb.GetOneIssue(epicKey.Num)
		if err != nil {
			utils.DoErrF("Could not find epic %s", epicKey.String())
			foundLookupError = true
			continue
		}
		if resp.Key != epic.Key {
			utils.DoErrF("Key mismatch %s != %s\n", resp.Key, epic.Key)
			foundLookupError = true
			continue
		}
		{
			str, ok := epic.Fields.CustomEpicLink.(string)
			if ok && str != "" {
				utils.DoErrF("Epic %s should not have an epic link (its %s)\n",
					epicKey.String(), str)
				foundEpicError = true
			}
		}
		if epic.Fields.Summary == "" {
			utils.DoErrF("Epic %s should have a summary\n", epicKey.String())
			foundEpicError = true
		}
		if !epic.IsEpic() {
			utils.DoErrF("Epic %s should have type %q, not %q\n",
				epicKey.String(), IssueTypeEpic, epic.TypeRaw())
			foundEpicError = true
		}
		if epic.Status() == IssueStatusUnknown {
			utils.DoErrF("Epic %s has bad status %q\n",
				epicKey.String(), epic.StatusRaw())
			foundEpicError = true
		}
		// epic contains the data to write
	}
	if foundLookupError {
		return fmt.Errorf(
			`aborting write due to lookup errors;
 fix issue numbers or create placeholder issues`)
	}
	if foundEpicError {
		return fmt.Errorf(`aborting write due to epic errors;
 if you want to overwrite the types, delete this line`)
	}
	return nil
}

// CreateDiGraph makes a digraph that includes _all_ epics in a project.
// It's time-consuming.
func (jb *JiraBoss) CreateDiGraph() (*Graph, error) {
	epicMap := jb.GetEpics()
	var nodes = make(map[MyKey]*Node)
	var edges = make(map[Edge]bool)
	for k := range epicMap {
		if k.Num >= UnknownEpicBase {
			continue
		}
		utils.DoErr1("Considering epic " + k.String())
		issue, err := jb.GetOneIssueByKey(k)
		if err != nil {
			return nil, err
		}
		if !issue.IsEpic() {
			err = fmt.Errorf(
				"GetEpics returned %s which is not an Epic",
				issue.Key)
			return nil, err
		}
		jb.considerEpic(issue, nodes, edges)
	}
	return MakeGraph(nodes, edges), nil
}

// considerEpic adds the incoming epic to a graph (if not already seen),
// then looks for epics that block it.
func (jb *JiraBoss) considerEpic(
	epic *ResponseIssue, visited map[MyKey]*Node, edges map[Edge]bool) {
	if _, seen := visited[epic.MyKey]; seen {
		return
	}
	epicKey := epic.MyKey
	visited[epicKey] = MakeNode(epic)
	if epicKey.Proj != jb.Project() {
		// don't recurse into issues from other projects
		return
	}
	for _, link := range epic.Fields.IssueLinks {
		if link.Type.Name == LinkTypeBlocks && link.InwardIssue.Key != "" {
			// The incoming epic is blocked by the other
			other := ParseMyKey(link.InwardIssue.Key)
			issue, err := jb.GetOneIssueByKey(other)
			if err != nil {
				err = fmt.Errorf(
					"in epic %s, unable to look up blocker %s; %w",
					epicKey, other, err)
				utils.DoErr1(err.Error())
				continue
			}
			if other != issue.MyKey {
				panic(fmt.Errorf(
					"looked up %s, got %s",
					other.String(), issue.MyKey.String()))
			}
			if !issue.IsEpic() {
				// Don't include non-epics in the graph, even if they are
				// blockers, because the graph might feed into other functions
				// like fixing dates, and we cannot expect date fields on
				// non-epics to be meaningful. Perhaps control this with flag.
				utils.DoErrF(
					"in epic %s, ignoring blockage by (non-epic) %s %s (%s)\n",
					epicKey, issue.Type(), issue.MyKey, issue.Status())
				continue
			}
			edges[MakeEdge(epicKey, issue.MyKey)] = true
			jb.considerEpic(issue, visited, edges)
		}
	}
}
