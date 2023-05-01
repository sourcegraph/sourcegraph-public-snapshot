package graphqlbackend

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gqlutil"
)

// PermissionsSyncJobResolver is used to resolve permission sync jobs.
//
// TODO(sashaostrikov) add PermissionsSyncJobProvider when it is persisted in the
// db.
type PermissionsSyncJobResolver interface {
	ID() graphql.ID
	State() string
	FailureMessage() *string
	Reason() PermissionsSyncJobReasonResolver
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
	Subject() PermissionsSyncJobSubject
	Priority() string
	NoPerms() bool
	InvalidateCaches() bool
	PermissionsAdded() int32
	PermissionsRemoved() int32
	PermissionsFound() int32
	CodeHostStates() []CodeHostStateResolver
	PartialSuccess() bool
	PlaceInQueue() *int32
}

type PermissionsSyncJobReasonResolver interface {
	Group() string
	Reason() *string
}

type CodeHostStateResolver interface {
	ProviderID() string
	ProviderType() string
	Status() database.CodeHostStatus
	Message() string
}

type PermissionsSyncJobSubject interface {
	ToRepository() (*RepositoryResolver, bool)
	ToUser() (*UserResolver, bool)
}

type ListPermissionsSyncJobsArgs struct {
	graphqlutil.ConnectionResolverArgs
	ReasonGroup *database.PermissionsSyncJobReasonGroup
	State       *database.PermissionsSyncJobState
	SearchType  *database.PermissionsSyncSearchType
	Query       *string
	UserID      *graphql.ID
	RepoID      *graphql.ID
	Partial     *bool
}
