package constants

// PhidType is the type of the PHID.
type PhidType string

const (
	// PhidTypeCommit is the PHID of a commit.
	PhidTypeCommit PhidType = "CMIT"

	// PhidTypeTask is the PHID of a task.
	PhidTypeTask PhidType = "TASK"

	// PhidTypeDifferentialRevision is the PHID of a differential revision.
	PhidTypeDifferentialRevision PhidType = "DREV"
)
