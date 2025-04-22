package myj

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"

	"github.com/monopole/gojira/internal/utils"
)

// This file (and sibling files) have the endpoints and associated types
// for marshalling (serializing) from a request struct to wire data (JSON),
// and un-marshalling (de-serializing) wire data (JSON) into a response struct.

type JiraBossIfc interface {
	Project() string
	GetOneIssue(int) (*ResponseIssue, error)
}

const (
	// https://developer.atlassian.com/server/jira/platform/rest/v10004/api-group-issue/#api-api-2-issue-issueidorkey-get
	endpointIssue = "rest/api/2/issue"
	// https://developer.atlassian.com/server/jira/platform/rest/v10004/api-group-issuetype/#api-group-issuetype
	endpointIssueType = "rest/api/2/issuetype"
	// https://developer.atlassian.com/server/jira/platform/rest/v10004/api-group-issuelink/#api-group-issuelink
	endpointIssueLink = "rest/api/2/issueLink"
)

const (
	LinkTypeBlocks = "Blocks"
)

// ResponseIssue holds responses from endpointIssue
type ResponseIssue struct {
	Fields AllIssueFields `json:"fields"`
	Id     string         `json:"id,omitempty"`
	Key    string         `json:"key,omitempty"`
	MyKey  MyKey          `json:"myKey,omitempty"`
}

type basicEpicFields struct {
	epicOnlyFields
	CommonIssueAndEpicFields
}

type AllIssueFields struct {
	epicOnlyFields
	issueOnlyFields
	CommonIssueAndEpicFields
	MiscIssueFields
}

// CustomFieldEpicName is called "customfield_12004"
const CustomFieldEpicName = "Epic Name"

type epicOnlyFields struct {
	// CustomEpicName is the 'Epic Name' field, used only in epics, but
	// not shown on the epic page.  It's shown only in an issue that's part
	// of the epic, as the text in the anchor link to that epic.
	// It's best to have this match the epic's summary field to avoid confusion.
	CustomEpicName string `json:"customfield_12004,omitempty"`
}

// CustomFieldEpicLink is called "customfield_12003"
const CustomFieldEpicLink = "Epic Link"

type issueOnlyFields struct {
	// CustomEpicLink is the 'Epic Link' field shown in an issue to
	// link to the epic that contains it (if any).
	// In an epic, this link is always nil.
	// In an issue if this is nil, then the issue is not part of an epic.
	// If it's a string, it's value is something like "BOB-7", where
	// issue 7 in the BOB project is an epic.
	// Making this "any" to allow use of nil to
	// clear the field in Jira, and NOT allowing omitempty
	CustomEpicLink any `json:"customfield_12003"`
}

const (
	// CustomFieldStartDate is called customfield_12134
	CustomFieldStartDate = "Start Date"
	// CustomFieldTargetCompletionDate is called customfield_11203
	CustomFieldTargetCompletionDate = "Target Completion Date"
)

// CommonIssueAndEpicFields holds fields common to issues and epics
type CommonIssueAndEpicFields struct {
	// Summary is the primary field for describing an issue or epic.
	Summary string `json:"summary,omitempty"`

	// CustomStartDate is the 'Start Date'.
	CustomStartDate string `json:"customfield_12134,omitempty"`

	// CustomTargetCompletionDate is the 'Target Completion Date'.
	CustomTargetCompletionDate string `json:"customfield_11203,omitempty"`
}

// MySummary returns the issue's summary, tacking on the
// name field if it is different.
func (ri *ResponseIssue) MySummary() string {
	if ri.Fields.CustomEpicName == "" {
		return ri.Fields.Summary
	}
	if ri.Fields.CustomEpicName == ri.Fields.Summary {
		return ri.Fields.Summary
	}
	return ri.Fields.Summary +
		" [" + utils.Ellipsis(ri.Fields.CustomEpicName, nameFieldTruc) + "]"
}

const indentVal = "  "

func doIndent(w io.Writer, depth int) {
	for i := 0; i < depth; i++ {
		_, _ = fmt.Fprint(w, indentVal)
	}
}

// LineRegExp parses line written by SpewParsable.
var LineRegExp = regexp.MustCompile(
	`(?P<proj>[A-Z]+)-(?P<num>\d+)\s+` + // match 1 and 2
		`\[(?P<type>[a-zA-Z\s]*)\]\s+` + // match 3
		`\((?P<status>[a-zA-Z\s]*)\)\s+` + // match 4
		`(?P<start>[a-zA-Z\-\d]*)\s+` + // match 5
		`(?P<end>[a-zA-Z\-\d]*)\s+` + // match 6
		`(?P<dayCount>[wd\d]*)\s+` + // match 7 (ignored)
		`\<(?P<labels>[\w\-,\s]*)\>\s+` + // match 8
		`(?P<summary>.*)$`) // match 9

// SpewParsable issue that can be parsed by LineRegExp.
func (ri *ResponseIssue) SpewParsable(w io.Writer, brief bool, depth int) {
	doIndent(w, depth)
	{
		fieldSize := 12 - (len(indentVal) * depth)
		f := fmt.Sprintf("%%-%ds", fieldSize)
		_, _ = fmt.Fprintf(w, f, ri.Key)
	}
	typ := ri.Type().String()
	if utils.Debug {
		str, ok := ri.Fields.CustomEpicLink.(string)
		if ok && str != "" {
			typ += " " + str
		}
	}

	// e.g. Story, Task
	_, _ = fmt.Fprintf(w, "%-12s ", "["+utils.Ellipsis(typ, 10)+"]")

	// e.g. Done, Backlog, In Progress
	_, _ = fmt.Fprintf(
		w, "%-16s ", "("+utils.Ellipsis(ri.Status().String(), 14)+")")

	d1 := ri.DateStart()
	d2 := ri.DateEnd()
	_, _ = fmt.Fprintf(w, "%11s ", d1)
	_, _ = fmt.Fprintf(w, "%11s ", d2)
	_, _ = fmt.Fprintf(w, "%3dw", d1.WeekCount(d2))

	_, _ = fmt.Fprintf(
		w, " %s", "<"+strings.Join(ri.Fields.Labels, ",")+"> ")
	//_, _ = fmt.Fprintf(
	//	w, " %s", "<"+utils.Ellipsis(strings.Join(ri.Fields.Labels, ","), 20)+">")

	_, _ = fmt.Fprintln(w, ri.MySummary())
	if brief {
		return
	}
	var blocks []MyKey
	var blockedBy []MyKey
	for _, link := range ri.Fields.IssueLinks {
		if link.Type.Name == LinkTypeBlocks {
			if link.InwardIssue.Key != "" {
				blockedBy = append(blockedBy, ParseMyKey(link.InwardIssue.Key))
			}
			if link.OutwardIssue.Key != "" {
				blocks = append(blocks, ParseMyKey(link.OutwardIssue.Key))
			}
		}
	}
	if len(blockedBy) > 0 {
		doIndent(w, depth+1)
		_, _ = fmt.Fprintln(w, "is blocked by")
		for _, k := range blockedBy {
			doIndent(w, depth+2)
			_, _ = fmt.Fprintln(w, k)
		}
	}
	if len(blocks) > 0 {
		doIndent(w, depth+1)
		_, _ = fmt.Fprintln(w, "blocks")
		for _, k := range blocks {
			doIndent(w, depth+2)
			_, _ = fmt.Fprintln(w, k)
		}
	}
}

func (ri *ResponseIssue) MakeMyKey() (result MyKey) {
	return ParseMyKey(ri.Key)
}

func (ri *ResponseIssue) SetMyKey() {
	ri.MyKey = ri.MakeMyKey()
}

func (ri *ResponseIssue) Status() IssueStatus {
	s, err := IssueStatusString(ri.StatusRaw())
	if err == nil {
		return s
	}
	return IssueStatusUnknown
}

func (ri *ResponseIssue) StatusRaw() string {
	return ri.Fields.Status.Name
}

func (ri *ResponseIssue) Type() IssueType {
	s, err := IssueTypeString(ri.TypeRaw())
	if err == nil {
		return s
	}
	return IssueTypeUnknown
}

func (ri *ResponseIssue) IsEpic() bool {
	return ri.Type() == IssueTypeEpic
}

// IsOkayUnderEpic kinda makes sense?  Not sure if we want this.
// Maybe we merely demand that it's not an epic.
func (ri *ResponseIssue) IsOkayUnderEpic() bool {
	t := ri.Type()
	return t == IssueTypeBug ||
		t == IssueTypeTask ||
		t == IssueTypeStory ||
		t == IssueTypeSubTask
}

func (ri *ResponseIssue) TypeRaw() string {
	return ri.Fields.IssueType.Name
}

func (ri *ResponseIssue) DateStart() utils.Date {
	return utils.FromJiraOrDie(ri.Fields.CustomStartDate)
}

func (ri *ResponseIssue) DateEnd() utils.Date {
	return utils.FromJiraOrDie(ri.Fields.CustomTargetCompletionDate)
}

type MiscIssueFields struct {
	IssueType  IssueTypeR        `json:"issuetype,omitempty"`
	Resolution ResolutionDetails `json:"resolution,omitempty"`
	Status     StatusDetails     `json:"status,omitempty"`
	//Creator     user              `json:"creator,omitempty"`
	//Description string            `json:"description,omitempty"`
	//Project     ProjectDetails    `json:"project,omitempty"`
	//Reporter    user              `json:"reporter,omitempty"`
	//Assignee    user              `json:"assignee,omitempty"`
	//Updated     string            `json:"updated,omitempty"`
	Labels     []string    `json:"labels,omitempty"`
	IssueLinks []IssueLink `json:"issueLinks,omitempty"`
}

type IssueLink struct {
	Id           string          `json:"id"`
	Type         IssueTypeR      `json:"type"`
	InwardIssue  IssueIdentifier `json:"inwardIssue"`
	OutwardIssue IssueIdentifier `json:"outwardIssue"`
}

type IssueTypeR struct {
	Id   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

type IssueIdentifier struct {
	Id   string `json:"id"`
	Self string `json:"self,omitempty"`
	Key  string `json:"key"`
}

type ProjectDetails struct {
	Id   string `json:"id,omitempty"`
	Self string `json:"self,omitempty"`
	Key  string `json:"key,omitempty"`
	Name string `json:"name,omitempty"`
}

type user struct {
	Id           string `json:"id,omitempty"`
	Self         string `json:"self,omitempty"`
	Key          string `json:"key,omitempty"`
	Name         string `json:"name,omitempty"`
	EmailAddress string `json:"emailAddress,omitempty"`
	DisplayName  string `json:"displayName,omitempty"`
}

type ResolutionDetails struct {
	Description string `json:"description,omitempty"`
	Id          string `json:"is,omitempty"`
	Name        string `json:"name,omitempty"`
}

type StatusDetails struct {
	Name string `json:"name,omitempty"`
}

func (jb *JiraBoss) CheckIssues(im map[MyKey]IssueList) error {
	foundLookupError := false
	foundTypeError := false
	count := 0
	for epicKey, issueList := range im {
		count++
		doErrF("Checking list %d with %d issues.\n",
			count, len(issueList))
		var (
			resp *ResponseIssue
			err  error
		)
		resp, err = jb.GetOneIssue(epicKey.Num)
		if err != nil {
			doErrF("Could not find epic %s", epicKey.String())
			foundLookupError = true
		} else if !resp.IsEpic() {
			doErrF("Why is the non-epic %s in the issue keys?\n", epicKey)
			foundTypeError = true
		}
		for _, issue := range issueList {
			resp, err = jb.GetOneIssue(issue.MyKey.Num)
			if err != nil {
				doErrF("Could not find issue %s", issue.Key)
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
	doErrF("%d now blocks %v\n", blocker, toBeBlocked)
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
				doErrF("%d no longer blocks %v\n", blocker, issue)
				break
			}
		}
	}
	doErrF("Deleted %d blocking links.\n", count)
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
