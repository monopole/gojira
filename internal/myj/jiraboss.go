package myj

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"

	"github.com/monopole/gojira/internal/utils"
)

// MyJiraArgs holds information needed to contact Jira
// (public or enterprise instance).
type MyJiraArgs struct {
	Host    string
	Project string
	Token   string
}

// JiraBoss manages requests to the jira api.
// Using the v2 version of the Jira API.
// https://developer.atlassian.com/cloud/jira/platform/rest/v2/intro/#version
//
// v3 is still in beta?
// https://developer.atlassian.com/cloud/jira/platform/rest/v3/intro/#version
type JiraBoss struct {
	htCl            *http.Client
	args            *MyJiraArgs
	placeholderEpic *ResponseIssue
}

func MakeJiraBoss(htCl *http.Client, args *MyJiraArgs) JiraBoss {
	return JiraBoss{
		htCl:            htCl,
		args:            args,
		placeholderEpic: makePlaceHolderEpic(UnknownEpicBase, args.Project),
	}
}

func (jb *JiraBoss) Project() string {
	return jb.args.Project
}

func (jb *JiraBoss) Key(issue int) MyKey {
	return MyKey{
		Proj: jb.Project(),
		Num:  issue,
	}
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
		http.MethodPut, req,
		endpointIssue+"/"+jb.Key(issue).String())
	return err
}

// LabelIssues adds or removes a label from the given issues.
func (jb *JiraBoss) LabelIssues(
	label string, issues []int, remove bool) error {
	for _, issue := range issues {
		debug1 := utils.Debug
		utils.Debug = false
		record, err := jb.GetOneIssue(issue)
		if err != nil {
			return err
		}
		utils.Debug = debug1
		labels := record.Fields.Labels
		found := false
		i := 0
		var newLabels []string
		for ; i < len(labels); i++ {
			if label == labels[i] {
				found = true
			} else {
				newLabels = append(newLabels, labels[i])
			}
		}
		if remove {
			if found {
				err = jb.writeLabels(issue, newLabels)
				if err != nil {
					return fmt.Errorf(
						"trouble removing label %q from issue %d; %w",
						label, issue, err)
				}
			}
		} else {
			if !found {
				newLabels = append(newLabels, label)
				err = jb.writeLabels(issue, newLabels)
				if err != nil {
					return fmt.Errorf(
						"trouble adding label %q to issue %d; %w",
						label, issue, err)
				}
			}
		}
	}
	return nil
}

func (jb *JiraBoss) writeLabels(issue int, labels []string) (err error) {
	type fieldsToWrite struct {
		Labels []string `json:"labels"` // Don't omitempty!
	}
	var req struct {
		Fields fieldsToWrite `json:"fields"`
	}
	req.Fields.Labels = labels
	_, err = jb.punchItChewie(
		http.MethodPut, req, endpointIssue+"/"+jb.Key(issue).String())
	return err
}

func (jb *JiraBoss) WriteEpics(em map[MyKey]*ResponseIssue) error {
	utils.DoErrF("Renaming %d epics.\n", len(em))
	for _, epic := range em {
		if err := jb.writeOneEpic(epic); err != nil {
			return fmt.Errorf("could not write epic %q %w", epic.MyKey, err)
		}
	}
	return nil
}

// writeOneEpic writes over data in an epic.
func (jb *JiraBoss) writeOneEpic(epic *ResponseIssue) (err error) {
	if epic.Fields.Summary == "" {
		return fmt.Errorf("bad data in issue write")
	}
	type fieldsToWrite struct {
		epicOnlyFields
		CommonIssueAndEpicFields
	}
	type requestPutIssue struct {
		Fields fieldsToWrite `json:"fields"`
	}
	req := requestPutIssue{
		Fields: fieldsToWrite{
			epicOnlyFields: epicOnlyFields{
				CustomEpicName: epic.Fields.Summary,
			},
			CommonIssueAndEpicFields: CommonIssueAndEpicFields{
				Summary:                    epic.Fields.Summary,
				CustomStartDate:            epic.Fields.CustomStartDate,
				CustomTargetCompletionDate: epic.Fields.CustomTargetCompletionDate,
			},
		},
	}
	if debug {
		utils.DoErrF("Would write epic %+v\n", req)
		return nil
	}
	_, err = jb.punchItChewie(
		http.MethodPut, req,
		endpointIssue+"/"+epic.MyKey.String())
	return err
}

func (jb *JiraBoss) WriteIssues(
	im map[MyKey]IssueList) error {
	utils.DoErrF("Writing %d issue lists.\n", len(im))
	for epic, list := range im {
		utils.DoErrF("  Writing %d issues.\n", len(list))
		for _, issue := range list {
			if err := jb.writeOneIssue(issue, epic); err != nil {
				return fmt.Errorf("could not write issue %q %w", issue.MyKey, err)
			}
		}
	}
	return nil
}

// writeOneIssue writes an issue to jira with the given epic link
func (jb *JiraBoss) writeOneIssue(issue *ResponseIssue, epic MyKey) (err error) {
	if issue.Fields.Summary == "" || issue.Status() == IssueStatusUnknown {
		return fmt.Errorf("bad data in issue write")
	}
	type fieldsToWrite struct {
		issueOnlyFields
		CommonIssueAndEpicFields
	}
	type requestPutIssue struct {
		Fields fieldsToWrite `json:"fields"`
	}
	// Not writing all fields, e.g. not overwriting status or type.
	// That's forbidden by the api, except maybe in a leap year.
	req := requestPutIssue{
		Fields: fieldsToWrite{
			issueOnlyFields: issueOnlyFields{
				CustomEpicLink: epic.String(),
			},
			CommonIssueAndEpicFields: CommonIssueAndEpicFields{
				Summary:                    issue.Fields.Summary,
				CustomStartDate:            issue.Fields.CustomStartDate,
				CustomTargetCompletionDate: issue.Fields.CustomTargetCompletionDate,
			},
		},
	}
	if debug {
		utils.DoErrF("Would write issue %+v\n", req)
		return nil
	}
	_, err = jb.punchItChewie(
		http.MethodPut, req,
		endpointIssue+"/"+issue.MyKey.String())
	return err
}

const debug = false

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
		http.MethodPut, req,
		endpointIssue+"/"+jb.Key(issue).String())
	return err
}

// GetCustomFieldId recovers information about field names that one
// needs to get what one wants from the API.
func (jb *JiraBoss) GetCustomFieldId(name string) string {
	fields, err := jb.DoOneFieldRequest()
	if err != nil {
		utils.DoErrF("no luck with field request")
		log.Fatal(err)
	}
	for _, f := range fields {
		if f.Name == name {
			return f.Id
		}
	}
	log.Fatalf("custom field '%s' not found", name)
	return ""
}

// GetEpics returns a map of issue keys (e.g. project-100) to issue structs,
// where all the issues happen to be epics.
func (jb *JiraBoss) GetEpics() (result map[MyKey]*ResponseIssue) {
	epics, err := jb.DoPagedSearch(jb.JqlEpics())
	if err != nil {
		log.Fatal(err)
	}
	result = make(map[MyKey]*ResponseIssue)
	// Make a placeholder for unknown epics
	result[jb.placeholderEpic.MakeMyKey()] = jb.placeholderEpic
	for i := range epics {
		key := epics[i].MakeMyKey()
		_, ok := result[key]
		if ok {
			log.Fatal(fmt.Errorf("epic %s appears more than once", key))
		}
		result[key] = &epics[i]
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
	issues, err := jb.DoPagedSearch(jb.JqlIssues())
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

const (
	// Pagination control
	maxResult    = 50
	maxMaxResult = 10000
)

func (jb *JiraBoss) DoPagedSearch(
	jql string) (result []ResponseIssue, err error) {
	req := makeSearchRequest(jql)
	for {
		var resp *ResponseSearch
		resp, err = jb.doOneSearchRequest(req)
		if err != nil {
			return nil, err
		}
		if len(resp.Issues) == 0 {
			break
		}
		result = append(result, resp.Issues...)
		req.StartAt += len(resp.Issues)
		if req.StartAt > maxMaxResult {
			break
		}
	}
	return
}

func (jb *JiraBoss) doOneSearchRequest(
	req RequestSearch) (*ResponseSearch, error) {
	var (
		err  error
		body []byte
	)
	body, err = jb.punchItChewie(http.MethodPost, req, endpointSearch)
	if err != nil {
		return nil, err
	}
	resp := &ResponseSearch{}
	err = json.Unmarshal(body, resp)
	if err != nil {
		return nil, fmt.Errorf("trouble unmarshaling response; %w", err)
	}
	for i := range resp.Issues {
		// Make these easy to sort.
		resp.Issues[i].SetMyKey()
	}
	return resp, nil
}

type IssueTypeFields struct {
	Description string `json:"description,omitempty"`
	Name        string `json:"name,omitempty"`
	Id          string `json:"id,omitempty"`
}

// DoOneIssueTypeRequest returns the type id associated with field names.
func (jb *JiraBoss) DoOneIssueTypeRequest() ([]IssueTypeFields, error) {
	body, err := jb.punchItChewie(http.MethodGet, nil, endpointIssueType)
	if err != nil {
		return nil, err
	}

	var resp []IssueTypeFields
	err = json.Unmarshal(body, &resp)
	if err != nil {
		return nil, fmt.Errorf("trouble unmarshaling response; %w", err)
	}
	return resp, nil
}

// UnknownEpicBase is a big number, out of range of the actual issue numbers.
// It serves as a fake epic to gather all epic-less issues.
const UnknownEpicBase = 9000

var unknownEpicOffset = 0

func (jb *JiraBoss) incrementUnknownEpic() *ResponseIssue {
	unknownEpicOffset = unknownEpicOffset + 1
	return makePlaceHolderEpic(UnknownEpicBase+unknownEpicOffset, jb.Project())
}

func makePlaceHolderEpic(i int, project string) *ResponseIssue {
	key := project + "-" + strconv.Itoa(i)
	name := "PlaceHolder Epic " + key
	result := &ResponseIssue{
		Fields: AllIssueFields{
			CommonIssueAndEpicFields: CommonIssueAndEpicFields{
				Summary:                    name,
				CustomStartDate:            "", // TODO: put in today?
				CustomTargetCompletionDate: "",
			},
			epicOnlyFields: epicOnlyFields{
				CustomEpicName: name,
			},
		},
		Id:  "",
		Key: key,
		MyKey: MyKey{
			Proj: project,
			Num:  i,
		},
	}
	return result
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
		jb.ConsiderEpic(issue, nodes, edges)
	}
	return MakeGraph(nodes, edges), nil
}

// ConsiderEpic adds the incoming epic to a graph (if not already seen),
// then looks for epics that block it.
func (jb *JiraBoss) ConsiderEpic(
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
			jb.ConsiderEpic(issue, visited, edges)
		}
	}
}

// WriteDates actually writes new dates to jira.
func (jb *JiraBoss) WriteDates(doIt bool, nodes map[MyKey]*Node) error {
	var lastErr error
	proposedChangeCount := 0
	success := 0
	for key, node := range nodes {
		if node.originalStart != node.startD {
			proposedChangeCount++
			_, _ = fmt.Fprintf(
				os.Stderr,
				"Start of %3d %s from %s to %s (%4d days)\n",
				key.Num,
				func() string {
					if doIt {
						return "moves"
					}
					return "could move"
				}(),
				node.originalStart.String(),
				node.startD.String(),
				node.originalStart.DayCount(node.startD))
			if doIt {
				if err := jb.SetDates(
					key.Num, node.startD, node.endD); err == nil {
					success++
				} else {
					lastErr = err
					utils.DoErr1(err.Error())
				}
			}
		}
	}
	if proposedChangeCount > 0 {
		if doIt {
			_, _ = fmt.Fprintf(
				os.Stderr, "Succeeded in %d of %d date moves.\n",
				success, proposedChangeCount)
		} else {
			utils.DoErrF(`
Redo this with --%s to move these %d dates,
or do them individually with 'set start' and 'set duration'.`,
				FlagFixDatesName, proposedChangeCount)
		}
	} else {
		utils.DoErrF("No changes proposed.\n")
	}
	return lastErr
}
