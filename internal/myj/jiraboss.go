package myj

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/monopole/gojira/internal/utils"
)

// MyJiraArgs holds information needed to contact Jira
// (public or enterprise instance).
type MyJiraArgs struct {
	Host    string
	Project string
	Token   string
}

type JiraBossIfc interface {
	Project() string
	GetOneIssue(int) (*ResponseIssue, error)
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

// RenameIssue renames an issue.
func (jb *JiraBoss) RenameIssue(n int, name string) error {
	issue, err := jb.GetOneIssue(n)
	if err != nil {
		return err
	}
	type requestPutIssue struct {
		Fields basicEpicFields `json:"fields"`
	}
	req := requestPutIssue{
		Fields: basicEpicFields{
			CommonIssueAndEpicFields: CommonIssueAndEpicFields{
				Summary: name,
			},
		},
	}
	if issue.IsEpic() {
		// For epics, always make the "short" name match the summary
		req.Fields.CustomEpicName = name
	}
	_, err = jb.punchItChewie(
		http.MethodPut, &req, endpointIssue+"/"+jb.Key(n).String())
	return err
}

// AssignIssues assigns or unassigns an issue.
func (jb *JiraBoss) AssignIssues(issues []int, ldap string) error {
	var req struct {
		Fields struct {
			Assignee struct {
				Name any `json:"name"` // Don't omitempty
			} `json:"assignee"`
		} `json:"fields"`
	}
	if ldap != "" {
		req.Fields.Assignee.Name = ldap
	}
	for _, n := range issues {
		_, err := jb.punchItChewie(
			http.MethodPut, &req, endpointIssue+"/"+jb.Key(n).String())
		if err != nil {
			return err
		}
	}
	return nil
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
	var req struct {
		Fields struct {
			Labels []string `json:"labels"` // Don't use omitempty
		} `json:"fields"`
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

// WriteDates actually writes new dates to jira.
func (jb *JiraBoss) WriteDates(doIt bool, nodes map[MyKey]*Node) error {
	var lastErr error
	proposedChangeCount := 0
	success := 0
	for key, node := range nodes {
		if node.originalStart != node.startD ||
			node.originalEnd != node.endD {
			proposedChangeCount++
			_, _ = fmt.Fprintf(
				os.Stderr,
				"%4d %s from%5d days duration starting on %s to%5d days duration starting on %s\n",
				key.Num,
				func() string {
					if doIt {
						return "changes"
					}
					return "should change"
				}(),
				node.originalStart.DayCount(node.originalEnd),
				node.originalStart.String(),
				node.startD.DayCount(node.endD),
				node.startD.String(),
			)
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
				FlagDoIt, proposedChangeCount)
		}
	} else {
		utils.DoErrF("No changes proposed.\n")
	}
	return lastErr
}

// SetDates sets the dates for an issue.
func (jb *JiraBoss) SetDates(
	issue int, start, end utils.Date) error {
	type requestPutIssue struct {
		Fields CommonIssueAndEpicFields `json:"fields"`
	}
	req := requestPutIssue{
		Fields: CommonIssueAndEpicFields{
			CustomStartDate:            start.JiraFormat(),
			CustomTargetCompletionDate: end.JiraFormat(),
		},
	}
	_, err := jb.punchItChewie(
		http.MethodPut, &req, endpointIssue+"/"+jb.Key(issue).String())
	return err
}

func (jb *JiraBoss) CheckIssues(im map[MyKey]IssueList) error {
	foundLookupError := false
	foundTypeError := false
	count := 0
	for epicKey, issueList := range im {
		count++
		utils.DoErrF("Checking list %d with %d issues.\n",
			count, len(issueList))
		var (
			resp *ResponseIssue
			err  error
		)
		resp, err = jb.GetOneIssue(epicKey.Num)
		if err != nil {
			utils.DoErrF("Could not find epic %s", epicKey.String())
			foundLookupError = true
		} else if !resp.IsEpic() {
			utils.DoErrF("Why is the non-epic %s in the issue keys?\n", epicKey)
			foundTypeError = true
		}
		for _, issue := range issueList {
			resp, err = jb.GetOneIssue(issue.MyKey.Num)
			if err != nil {
				utils.DoErrF("Could not find issue %s", issue.Key)
				foundLookupError = true
				continue
			}

			if !resp.IsOkayUnderEpic() {
				foundTypeError = true
				continue
			}
		}
	}
	if foundLookupError {
		return fmt.Errorf(
			`aborting write due to lookup errors;
 fix issue numbers or create placeholder issues`)
	}
	if foundTypeError {
		return fmt.Errorf(`aborting write due type errors;
 if you want to overwrite the types, delete this line`)
	}
	return nil
}

// GetOneIssue recovers info about the issue called {project}-{id}.
func (jb *JiraBoss) GetOneIssue(issue int) (*ResponseIssue, error) {
	return jb.GetOneIssueByKey(jb.Key(issue))
}

// GetOneIssueByKey recovers info about the issue.
func (jb *JiraBoss) GetOneIssueByKey(issue MyKey) (*ResponseIssue, error) {
	var (
		err  error
		resp ResponseIssue
		body []byte
	)
	body, err = jb.punchItChewie(
		http.MethodGet, nil,
		endpointIssue+"/"+issue.String())
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(body, &resp)
	if err != nil {
		return nil, fmt.Errorf("trouble unmarshaling issue; %w", err)
	}
	resp.SetMyKey()
	return &resp, nil
}

// GetTransitionId finds the id of some transition.
func (jb *JiraBoss) GetTransitionId(issue int, status IssueStatus) (string, error) {
	type transition struct {
		Id   string `json:"id"`
		Name string `json:"name"`
	}
	type tranReq struct {
		Transitions []transition `json:"transitions"`
	}
	var (
		err  error
		req  tranReq
		body []byte
	)
	body, err = jb.punchItChewie(
		http.MethodGet, nil,
		endpointIssue+"/"+jb.Key(issue).String()+"/transitions")
	if err != nil {
		return "", err
	}
	err = json.Unmarshal(body, &req)
	if err != nil {
		return "", fmt.Errorf("trouble unmarshaling issue; %w", err)
	}
	for _, t := range req.Transitions {
		if strings.HasPrefix(t.Name, status.String()) {
			return t.Id, nil
		}
	}
	return "", fmt.Errorf("unable to find transition to %q", status)
}

// BlockIssues makes the first issue block the others.
func (jb *JiraBoss) BlockIssues(blocker int, toBeBlocked []int, comment string) error {
	var req struct {
		Type struct {
			Name string `json:"name"`
		} `json:"type"`
		// InwardIssue is the issue doing the blocking
		InwardIssue struct {
			Key string `json:"key"`
		} `json:"inwardIssue"`
		// OutwardIssue is the issue being blocked.
		OutwardIssue struct {
			Key string `json:"key"`
		} `json:"outwardIssue"`
		Comment struct {
			Body string `json:"body,omitempty"`
		} `json:"comment,omitempty"`
	}
	req.Type.Name = LinkTypeBlocks
	req.InwardIssue.Key = jb.Key(blocker).String()
	if comment != "" {
		req.Comment.Body = comment
	}
	for _, dependent := range toBeBlocked {
		req.OutwardIssue.Key = jb.Key(dependent).String()
		_, err := jb.punchItChewie(http.MethodPost, &req, endpointIssueLink)
		if err != nil {
			return err
		}
	}
	utils.DoErrF("%d now blocks %v\n", blocker, toBeBlocked)
	return nil
}

// UnBlockIssues deletes the links created by BlockIssues.
func (jb *JiraBoss) UnBlockIssues(blocker int, blocked []int) error {
	var (
		err  error
		body []byte
		resp ResponseIssue
	)
	body, err = jb.punchItChewie(
		http.MethodGet, nil,
		endpointIssue+"/"+jb.Key(blocker).String()+"?expand=issuelinks")
	if err != nil {
		return err
	}
	if err = json.Unmarshal(body, &resp); err != nil {
		return fmt.Errorf("trouble unmarshaling issue links; %w", err)
	}
	count := 0
	for _, issue := range blocked {
		key := jb.Key(issue)
		for _, link := range resp.Fields.IssueLinks {
			if link.Type.Name == LinkTypeBlocks &&
				link.OutwardIssue.Key == key.String() {
				if err = jb.deleteLink(link.Id); err != nil {
					return err
				}
				count++
				utils.DoErrF("%d no longer blocks %v\n", blocker, issue)
				break
			}
		}
	}
	utils.DoErrF("Deleted %d blocking links.\n", count)
	return nil
}

func (jb *JiraBoss) deleteLink(id string) error {
	_, err := jb.punchItChewie(http.MethodDelete, nil, endpointIssueLink+"/"+id)
	return err
}

// MoveIssueToState moves an issue to a new state
func (jb *JiraBoss) MoveIssueToState(n int, stateId string) error {
	var req struct {
		Transition struct {
			Id string `json:"id"`
		} `json:"transition"`
	}
	req.Transition.Id = stateId
	_, err := jb.punchItChewie(
		http.MethodPost, &req, endpointIssue+"/"+jb.Key(n).String()+"/transitions")
	return err
}

// GetOneIssueEditMeta recovers metadata (field accessibility)
// about the issue called {project}-{id}.
func (jb *JiraBoss) GetOneIssueEditMeta(issue int) (*ResponseEditMeta, error) {
	var (
		err  error
		resp ResponseEditMeta
		body []byte
	)
	body, err = jb.punchItChewie(
		http.MethodGet, nil,
		endpointIssue+"/"+jb.Key(issue).String()+"/editmeta")
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(body, &resp)
	if err != nil {
		return nil, fmt.Errorf("trouble unmarshaling issue; %w", err)
	}
	return &resp, nil
}

type ResponseEditMeta struct {
	Fields map[string]any `json:"fields,omitempty"`
}
