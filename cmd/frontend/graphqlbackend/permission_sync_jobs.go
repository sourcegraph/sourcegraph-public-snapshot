package graphqlbackend

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/internal/gqlutil"
)

// PermissionSyncJobResolver is used to resolve permission sync jobs.
//
// TODO(sashaostrikov) add PermissionSyncJobProvider when it is persisted in the
// db.
type PermissionSyncJobResolver interface {
	ID() graphql.ID
	State() string
	FailureMessage() *string
	Reason() PermissionSyncJobReasonResolver
	CancellationReason() *string
	TriggeredByUser(ctx context.Context) (*UserResolver, error)
	QueuedAt() gqlutil.DateTime
	StartedAt() *gqlutil.DateTime
	FinishedAt() *gqlutil.DateTime
	ProcessAfter() *gqlutil.DateTime
	RanForMs() *int32
	NumResets() *int32
	NumFailures() *int32
	LastHeartbeatAt() *gqlutil.DateTime
	WorkerHostname() string
	Cancel() bool
	Subject() PermissionSyncJobSubject
	Priority() string
	NoPerms() bool
	InvalidateCaches() bool
	PermissionsAdded() int32
	PermissionsRemoved() int32
	PermissionsFound() int32
	CodeHostStates() []CodeHostStateResolver
}

type PermissionSyncJobReasonResolver interface {
	Group() string
	Message() string
}

type CodeHostStateResolver interface {
	ProviderID() string
	ProviderType() string
	Status() string
	Message() string
}

type PermissionSyncJobSubject interface {
	ToRepository() (*RepositoryResolver, bool)
	ToUser() (*UserResolver, bool)
}

type ListPermissionSyncJobsArgs struct {
	graphqlutil.ConnectionResolverArgs
}
