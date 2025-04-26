package myj

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// https://developer.atlassian.com/server/jira/platform/rest/v10004/api-group-field/#api-group-field
const endpointField = "rest/api/2/field"

// FieldFields is used in requests for field names.
type FieldFields struct {
	ClauseNames []string `json:"clauseNames,omitempty"`
	Custom      bool     `json:"custom,omitempty"`
	Id          string   `json:"id,omitempty"`
	Name        string   `json:"name,omitempty"`
	Navigable   bool     `json:"navigable,omitempty"`
	Orderable   bool     `json:"orderable,omitempty"`
	Schema      schema   `json:"schema,omitempty"`
	Searchable  bool     `json:"searchable,omitempty"`
	DueDate     string   `json:"duedate,omitempty"`
}

type schema struct {
	Custom   string `json:"custom,omitempty"`
	CustomId int    `json:"customId,omitempty"`
	Items    string `json:"items,omitempty"`
	System   string `json:"system,omitempty"`
	Type     string `json:"type,omitempty"`
}

func (jb *JiraBoss) DoOneFieldRequest() ([]FieldFields, error) {
	body, err := jb.punchItChewie(http.MethodGet, nil, endpointField)
	if err != nil {
		return nil, err
	}
	var resp []FieldFields
	err = json.Unmarshal(body, &resp)
	if err != nil {
		return nil, fmt.Errorf("trouble unmarshaling response; %w", err)
	}
	return resp, nil
}
