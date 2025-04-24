package myj

//go:generate go run github.com/dmarkham/enumer -linecomment -type=IssueStatus
type IssueStatus int

const (
	IssueStatusUnknown          IssueStatus = iota
	IssueStatusBacklog                      // Backlog
	IssueStatusDone                         // Done
	IssueStatusClosed                       // Closed
	IssueStatusClosedWoAction               // Closed Without Action
	IssueStatusInProgress                   // In Progress
	IssueStatusInQueue                      // In Queue
	IssueStatusReleaseCandidate             // Release Candidate
	IssueStatusInValidation                 // In Validation
	IssueStatusValidation                   // Validation
	IssueStatusReadyForDev                  // Ready for Development
	IssueStatusReadyForReview               // Ready for Review
)
