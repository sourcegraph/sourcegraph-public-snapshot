package types

import (
	"time"

	"github.com/sourcegraph/sourcegraph/internal/workerutil"
)

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
	ExecutionLogs   []workerutil.ExecutionLogEntry
	WorkerHostname  string

	// Name of the Bitbucket Project
	ProjectKey string
	// ID of the external service that contains the Bitbucket address and credentials
	ExternalServiceID int64
	// List of user permissions for the Bitbucket Project
	Permissions []UserPermission
	// Whether all of the repos of the project are unrestricted
	Unrestricted bool
}

type BitbucketProjectPermissionsJobModel struct {
	ID              int `gorm:"primarykey"`
	State           string
	FailureMessage  *string
	QueuedAt        time.Time
	StartedAt       *time.Time
	FinishedAt      *time.Time
	ProcessAfter    *time.Time
	NumResets       int
	NumFailures     int
	LastHeartbeatAt time.Time

	WorkerHostname string

	// Name of the Bitbucket Project
	ProjectKey string
	// ID of the external service that contains the Bitbucket address and credentials
	ExternalServiceID int64

	// Whether all of the repos of the project are unrestricted
	Unrestricted bool
}

func (BitbucketProjectPermissionsJobModel) TableName() string {
	return "explicit_permissions_bitbucket_projects_jobs"
}

// RecordID implements workerutil.Record.
func (g *BitbucketProjectPermissionJob) RecordID() int {
	return g.ID
}

type UserPermission struct {
	BindID     string `json:"bindID"`
	Permission string `json:"permission"`
}
