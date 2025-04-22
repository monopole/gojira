package troper

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"unicode"

	"github.com/monopole/gojira/internal/utils"

	"github.com/monopole/gojira/internal/myj"
	"github.com/spf13/afero"
)

type ParsedIssue struct {
	Proj      string
	Num       int
	Type      myj.IssueType
	Status    myj.IssueStatus
	RawLabels []string
	Start     utils.Date
	End       utils.Date
	Summary   string
}

type ParsedJiraLine struct {
	IsEpic bool
	ParsedIssue
}

// UnSpewEpics reads a file created by SpewEpics.
func UnSpewEpics(fs afero.Fs, path string) (
	result []*ParsedJiraLine,
	err error) {
	var data []byte
	data, err = afero.ReadFile(fs, path)
	if err != nil {
		return
	}
	for _, line := range bytes.Split(data, []byte("\n")) {
		var issue *ParsedJiraLine
		issue, err = parseLine(line)
		if err != nil {
			return // TODO: return err
		}
		if issue != nil {
			result = append(result, issue)
		}
	}
	return
}

func parseLine(line []byte) (*ParsedJiraLine, error) {
	remainder := bytes.TrimLeftFunc(line, unicode.IsSpace)
	// Ignore empty lines.
	if len(remainder) < 1 {
		// ignore empty line
		return nil, nil
	}

	// No indent means it's an epic.
	isEpic := len(remainder) == len(line)
	match := myj.LineRegExp.FindStringSubmatch(string(remainder))
	if len(match) != 10 {
		return nil, fmt.Errorf("unable to parse %q (len(match)=%d)",
			string(line), len(match))
	}

	var arg string

	arg = match[2]
	num, err := strconv.Atoi(arg)
	if err != nil {
		return nil, fme("issue number", arg, line)
	}

	var iType myj.IssueType
	arg = match[3]
	iType, err = myj.IssueTypeString(arg)
	if err != nil {
		return nil, fme("type", arg, line)
	}

	arg = match[4]
	var status myj.IssueStatus
	status, err = myj.IssueStatusString(arg)
	if err != nil {
		return nil, fme("status", arg, line)
	}

	arg = match[5]
	var start utils.Date
	start, err = utils.ParseDate(arg)
	if err != nil {
		return nil, fme("start date", arg, line)
	}

	arg = match[6]
	var end utils.Date
	end, err = utils.ParseDate(arg)
	if err != nil {
		return nil, fme("end date", arg, line)
	}

	result := ParsedJiraLine{
		IsEpic: isEpic,
		ParsedIssue: ParsedIssue{
			Proj:    match[1],
			Num:     num,
			Type:    iType,
			Status:  status,
			Start:   start,
			End:     end,
			Summary: match[9],
		},
	}
	if len(match[8]) > 0 {
		result.RawLabels = strings.Split(match[8], ",")
	}
	return &result, nil
}

func fme(f, arg string, line []byte) error {
	return fmt.Errorf(
		"unable to parse %s from %q in line %q", f, arg, string(line))
}

// Convert the args to structs used to communicate with the Jira API.
// LABELS not processed at time of writing
func Convert(lines []*ParsedJiraLine) (
	epicMap map[myj.MyKey]*myj.ResponseIssue,
	issueMap map[myj.MyKey]myj.IssueList,
) {
	epicMap = make(map[myj.MyKey]*myj.ResponseIssue)
	issueMap = make(map[myj.MyKey]myj.IssueList)

	if len(lines) == 0 {
		return
	}

	var issues myj.IssueList

	line := lines[0]
	if !line.IsEpic {
		panic("the first line should be an epic")
	}
	epic := myj.MyKey{
		Proj: line.Proj,
		Num:  line.Num,
	}
	epicMap[epic] = makeIssueRecord(epic, &line.ParsedIssue)
	for i := 1; i < len(lines); i++ {
		line = lines[i]
		key := myj.MyKey{
			Proj: line.Proj,
			Num:  line.Num,
		}
		if line.IsEpic {
			// Close previous epic.
			issueMap[epic] = issues
			// Start new epic.
			epic = key
			issues = nil
			epicMap[epic] = makeIssueRecord(epic, &line.ParsedIssue)
		} else {
			issue := makeIssueRecord(key, &line.ParsedIssue)
			issue.Fields.CustomEpicLink = epic.String()
			issues = append(issues, issue)
		}
	}
	// Close previous epic.
	issueMap[epic] = issues
	return epicMap, issueMap
}

func makeIssueRecord(key myj.MyKey, issue *ParsedIssue) *myj.ResponseIssue {
	res := myj.ResponseIssue{
		Fields: myj.AllIssueFields{
			CommonIssueAndEpicFields: myj.CommonIssueAndEpicFields{
				Summary: issue.Summary,
			},
			MiscIssueFields: myj.MiscIssueFields{
				IssueType:  myj.IssueTypeR{Name: issue.Type.String()},
				Resolution: myj.ResolutionDetails{},
				Status:     myj.StatusDetails{Name: issue.Status.String()},
			},
		},
		Key:   key.String(),
		MyKey: key,
	}
	if len(issue.RawLabels) > 0 {
		res.Fields.Labels = issue.RawLabels
	}
	if issue.Start.IsGood() {
		res.Fields.CustomStartDate = issue.Start.JiraFormat()
	}
	if issue.End.IsGood() {
		res.Fields.CustomTargetCompletionDate = issue.End.JiraFormat()
	}
	return &res
}
