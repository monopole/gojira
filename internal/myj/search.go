package myj

const endpointSearch = "rest/api/2/search"

// https://developer.atlassian.com/server/jira/platform/rest/v10004/api-group-search/#api-api-2-search-post
func makeSearchRequest(jql string) RequestSearch {
	return RequestSearch{
		// Jql is the actual query
		Jql: jql,

		// MaxResults is when to stop pagination (one we have this many).
		MaxResults: maxResult,

		// StartAt is used in pagination
		StartAt: 0,

		// Fields is what we want in the response for each issue.
		// Send Fields:nil to get all fields, but be mindful that you'll lose
		// them when marshalling from JSON if you've not specified fields for
		// them.
		Fields: []string{
			// id is a jira internal number with seven or so digits.
			"id",

			// key is something like BOB-25038.
			"key",

			// summary is the issue summary,
			// e.g. "users wants this blue thing to be red".
			"summary",

			// resolution is a struct describing the conditions of resolution.
			"resolution",

			// status is ?
			"status",

			// labels is a string array of labels.
			"labels",

			// assignee is a struct describing a user
			// - name, email, displayName, etc.
			"assignee",

			// reporter is a struct describing a user
			"reporter",

			// creator is a struct describing a user
			"creator",

			// The epic name field (often same as summary)
			"customfield_12004",

			// The "Epic Link" field
			"customfield_12003",

			// Start Date
			"customfield_12134",

			// Target Completion Date
			"customfield_11203",

			// project is a struct with a key like "MSFT",
			// a name like "microsoft developers", and avatar urls.
			// A project url takes the form:
			// https://issues.acmecorp.com/projects/PLM/issues
			"project",

			// description is the long textual description of the issue.
			"description",

			// updated is the timestamp associated with the most recent update.
			"updated",

			// issuetype is the type of issue
			// (task, bug, story, epic, initiative)
			"issuetype",

			// epicLink is the one epic with which this issue is associated
			"epicLink",
		},
		Expand: []string{"renderedFields", "names"},
	}
}

type RequestSearch struct {
	Expand     []string `json:"expand,omitempty"`
	Jql        string   `json:"jql,omitempty"`
	MaxResults int      `json:"maxResults,omitempty"`
	Fields     []string `json:"fields,omitempty"`
	StartAt    int      `json:"startAt"`
}

type ResponseSearch struct {
	Expand     string          `json:"expand,omitempty"`
	Issues     []ResponseIssue `json:"issues,omitempty"`
	MaxResults int             `json:"maxResults,omitempty"`
	StartAt    int             `json:"startAt,omitempty"`
	Total      int             `json:"total,omitempty"`
}
