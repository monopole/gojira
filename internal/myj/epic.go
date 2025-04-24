package myj

import (
	"encoding/json"
	"fmt"
	"github.com/monopole/gojira/internal/utils"
	"net/http"
)

// The issue endpoint is sufficient (an epic is an issue), but this has some
// interesting things
// https://developer.atlassian.com/server/jira/platform/rest/v10004/api-group-epic/#api-group-epic
// const endpointEpic = "rest/agile/1.0/epic"

type ResponseGetEpic struct {
	Key    string          `json:"key,omitempty"`
	Fields basicEpicFields `json:"fields"`
}

const nameFieldTruc = 30

// MySummary returns the issue's summary, tacking on the
// name field if it is different.
func (ri *ResponseGetEpic) MySummary() string {
	if ri.Fields.CustomEpicName == "" {
		return ri.Fields.Summary
	}
	if ri.Fields.CustomEpicName == ri.Fields.Summary {
		return ri.Fields.Summary
	}
	tmp := ri.Fields.CustomEpicName
	if len(tmp) > nameFieldTruc {
		tmp = tmp[:nameFieldTruc] + "..."
	}
	return ri.Fields.Summary + " [" + tmp + "]"
}

// GetOneEpic gets some data on an epic.
func (jb *JiraBoss) GetOneEpic(epic int) (*ResponseGetEpic, error) {
	body, err := jb.punchItChewie(
		http.MethodGet, nil, endpointIssue+"/"+jb.Key(epic).String())
	if err != nil {
		return nil, err
	}
	var resp ResponseGetEpic
	err = json.Unmarshal(body, &resp)
	if err != nil {
		return nil, fmt.Errorf("unmarshaling get epic response; %w", err)
	}
	return &resp, nil
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

// FixEpicName gets the epic, reads the string value in the summary field
// and writes that value to the custom name field so that they match.
func (jb *JiraBoss) FixEpicName(epic int) error {
	r, err := jb.GetOneEpic(epic)
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

func (jb *JiraBoss) CheckEpics(
	em map[MyKey]*ResponseIssue) error {
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
