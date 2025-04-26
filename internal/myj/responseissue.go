package myj

import (
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/monopole/gojira/internal/utils"
)

// This file (and sibling files) have the endpoints and associated types
// for marshalling (serializing) from a request struct to wire data (JSON),
// and un-marshalling (de-serializing) wire data (JSON) into a response struct.

const (
	// https://developer.atlassian.com/server/jira/platform/rest/v10004/api-group-issue/#api-api-2-issue-issueidorkey-get
	endpointIssue = "rest/api/2/issue"
	// https://developer.atlassian.com/server/jira/platform/rest/v10004/api-group-issuetype/#api-group-issuetype
	endpointIssueType = "rest/api/2/issuetype"
	// https://developer.atlassian.com/server/jira/platform/rest/v10004/api-group-issuelink/#api-group-issuelink
	endpointIssueLink = "rest/api/2/issueLink"

	// The issue endpoint is sufficient (an epic is an issue), but this
	// other endpoint has some interesting things
	// https://developer.atlassian.com/server/jira/platform/rest/v10004/api-group-epic/#api-group-epic
	// const endpointEpic = "rest/agile/1.0/epic"
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

type MiscIssueFields struct {
	IssueType  IssueTypeR        `json:"issuetype,omitempty"`
	Resolution ResolutionDetails `json:"resolution,omitempty"`
	Status     StatusDetails     `json:"status,omitempty"`
	//Creator     humanUser              `json:"creator,omitempty"`
	//Description string            `json:"description,omitempty"`
	//Project     ProjectDetails    `json:"project,omitempty"`
	//Reporter    humanUser              `json:"reporter,omitempty"`
	Assignee humanUser `json:"assignee,omitempty"`
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

type humanUser struct {
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

// MySummary returns the issue's summary, tacking on the
// name field if it is different.
func (ri *ResponseIssue) MySummary() string {
	const nameFieldTruc = 30
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
		fieldSize := 14 - (len(indentVal) * depth)
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
	_, _ = fmt.Fprintf(w, "%11s", d2)
	_, _ = fmt.Fprintf(w, "%4dw", d1.WeekCount(d2))

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

func (ri *ResponseIssue) AssigneeName() string {
	if len(ri.Fields.Assignee.DisplayName) > 0 {
		return ri.Fields.Assignee.DisplayName
	}
	return ri.AssigneeLdap()
}

func (ri *ResponseIssue) AssigneeLdap() string {
	return ri.Fields.Assignee.Name
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
