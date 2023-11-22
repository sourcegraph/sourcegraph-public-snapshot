package types

import (
	"database/sql/driver"
	"strconv"
	"time"
)

type ExecutionLogEntry interface {
	Scan(value any) error
	Value() (driver.Value, error)
}

// BitbucketProjectPermissionJob represents a task to apply a set of permissions
// to all the repos of the given Bitbucket project.
type BitbucketProjectPermissionJob struct {
	ID              int
	State           string
	FailureMessage  *string
	QueuedAt        time.Time
	StartedAt       *time.Time
	FinishedAt      *time.Time
	ProcessAfter    *time.Time
	NumResets       int
	NumFailures     int
	LastHeartbeatAt time.Time
	ExecutionLogs   []ExecutionLogEntry
	WorkerHostname  string

	// Name of the Bitbucket Project
	ProjectKey string
	// ID of the external service that contains the Bitbucket address and credentials
	ExternalServiceID int64
	// List of user permissions for the Bitbucket Project
	Permissions []UserPermission
	// Whether all the repos of the project are unrestricted
	Unrestricted bool
}

// RecordID implements workerutil.Record.
func (g *BitbucketProjectPermissionJob) RecordID() int {
	return g.ID
}

func (g *BitbucketProjectPermissionJob) RecordUID() string {
	return strconv.Itoa(g.ID)
}

type UserPermission struct {
	BindID     string `json:"bindID"`
	Permission string `json:"permission"`
}
