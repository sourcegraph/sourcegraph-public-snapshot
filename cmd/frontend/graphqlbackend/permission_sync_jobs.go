package graphqlbackend

import (
	"context"
	"time"

	"github.com/graph-gophers/graphql-go"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/executor"
)

// PermissionSyncJobsResolver is a main interface for all GraphQL operations with
// permission sync jobs.
type PermissionSyncJobsResolver interface {
	PermissionSyncJobs(ctx context.Context, args ListPermissionSyncJobsArgs) (*graphqlutil.ConnectionResolver[PermissionSyncJobResolver], error)
}

type PermissionSyncJobResolver interface {
	ID() graphql.ID
	State() string
	FailureMessage() *string
	Reason() database.PermissionSyncJobReason
	CancellationReason() string
	TriggeredByUserID() int32
	QueuedAt() time.Time
	StartedAt() time.Time
	FinishedAt() time.Time
	ProcessAfter() time.Time
	NumResets() int
	NumFailures() int
	LastHeartbeatAt() time.Time
	ExecutionLogs() []executor.ExecutionLogEntry
	WorkerHostname() string
	Cancel() bool
	RepositoryID() graphql.ID
	UserID() graphql.ID
	Priority() database.PermissionSyncJobPriority
	NoPerms() bool
	InvalidateCaches() bool
	PermissionsAdded() int
	PermissionsRemoved() int
	PermissionsFound() int
}

type ListPermissionSyncJobsArgs struct {
	graphqlutil.ConnectionResolverArgs
}
