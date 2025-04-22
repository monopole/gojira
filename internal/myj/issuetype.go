package myj

//go:generate go run github.com/dmarkham/enumer -linecomment -type=IssueType
type IssueType int

const (
	IssueTypeUnknown    IssueType = iota
	IssueTypeInitiative           // Initiative
	IssueTypeEpic                 // Epic
	IssueTypeStory                // Story
	IssueTypeTask                 // Task
	IssueTypeSubTask              // Sub-task
	IssueTypeBug                  // Bug
)
