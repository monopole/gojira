package myj

// CustomFieldEpicName is the human name for "customfield_12004".
const CustomFieldEpicName = "Epic Name"

type epicOnlyFields struct {
	// CustomEpicName is the 'Epic Name' field, used only in epics, but
	// not shown on the epic page.  It's shown only in an issue that's part
	// of the epic, as the text in the anchor link to that epic.
	// It's best to have this match the epic's summary field to avoid confusion.
	CustomEpicName string `json:"customfield_12004,omitempty"`
}

// CustomFieldEpicLink is the human name for "customfield_12003"
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
	// CustomFieldStartDate is the human name for customfield_12134
	CustomFieldStartDate = "Start Date"
	// CustomFieldTargetCompletionDate is the human name for customfield_11203
	CustomFieldTargetCompletionDate = "Target Completion Date"
)

// CommonIssueAndEpicFields holds fields common to issues and epics
type CommonIssueAndEpicFields struct {
	// Summary is the primary field for describing an issue or epic.
	Summary string `json:"summary,omitempty"`

	// CustomStartDate is the 'Start Date', a string like "2006-01-02"
	CustomStartDate string `json:"customfield_12134,omitempty"`

	// CustomTargetCompletionDate is the 'Target Completion Date'.
	CustomTargetCompletionDate string `json:"customfield_11203,omitempty"`
}
