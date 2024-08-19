package graphqlbackend

import (
	"context"

	"github.com/graph-gophers/graphql-go"

	"github.com/sourcegraph/sourcegraph/internal/gqlutil"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

type AuthzResolver interface {
	// SetRepositoryPermissionsForUsers and functions below are GraphQL Mutations.
	SetRepositoryPermissionsForUsers(ctx context.Context, args *RepoPermsArgs) (*EmptyResponse, error)
	SetRepositoryPermissionsUnrestricted(ctx context.Context, args *RepoUnrestrictedArgs) (*EmptyResponse, error)
	ScheduleRepositoryPermissionsSync(ctx context.Context, args *RepositoryIDArgs) (*EmptyResponse, error)
	ScheduleUserPermissionsSync(ctx context.Context, args *UserPermissionsSyncArgs) (*EmptyResponse, error)
	SetSubRepositoryPermissionsForUsers(ctx context.Context, args *SubRepoPermsArgs) (*EmptyResponse, error)
	SetRepositoryPermissionsForBitbucketProject(ctx context.Context, args *RepoPermsBitbucketProjectArgs) (*EmptyResponse, error)
	CancelPermissionsSyncJob(ctx context.Context, args *CancelPermissionsSyncJobArgs) (CancelPermissionsSyncJobResultMessage, error)

	// AuthorizedUserRepositories and functions below are GraphQL Queries.
	AuthorizedUserRepositories(ctx context.Context, args *AuthorizedRepoArgs) (RepositoryConnectionResolver, error)
	UsersWithPendingPermissions(ctx context.Context) ([]string, error)
	AuthorizedUsers(ctx context.Context, args *RepoAuthorizedUserArgs) (UserConnectionResolver, error)
	BitbucketProjectPermissionJobs(ctx context.Context, args *BitbucketProjectPermissionJobsArgs) (BitbucketProjectsPermissionJobsResolver, error)
	PermissionsSyncJobs(ctx context.Context, args ListPermissionsSyncJobsArgs) (*gqlutil.ConnectionResolver[PermissionsSyncJobResolver], error)
	PermissionsSyncingStats(ctx context.Context) (PermissionsSyncingStatsResolver, error)

	// RepositoryPermissionsInfo and UserPermissionsInfo are helpers functions.
	RepositoryPermissionsInfo(ctx context.Context, repoID graphql.ID) (PermissionsInfoResolver, error)
	UserPermissionsInfo(ctx context.Context, userID graphql.ID) (PermissionsInfoResolver, error)
}

type RepositoryIDArgs struct {
	Repository graphql.ID
}

type UserPermissionsSyncArgs struct {
	User    graphql.ID
	Options *struct {
		InvalidateCaches *bool
	}
}

type RepoPermsArgs struct {
	Repository      graphql.ID
	UserPermissions []struct {
		BindID     string
		Permission string
	}
}

type RepoUnrestrictedArgs struct {
	Repositories []graphql.ID
	Unrestricted bool
}

type SubRepoPermsArgs struct {
	Repository      graphql.ID
	UserPermissions []struct {
		BindID       string
		PathIncludes *[]string
		PathExcludes *[]string
		Paths        *[]string
	}
}

type AuthorizedRepoArgs struct {
	Username *string
	Email    *string
	Perm     string
	First    int32
	After    *string
}

type RepoPermsBitbucketProjectArgs struct {
	ProjectKey      string
	CodeHost        graphql.ID
	UserPermissions []types.UserPermission
	Unrestricted    *bool
}

type CancelPermissionsSyncJobArgs struct {
	Job    graphql.ID
	Reason *string
}

type BitbucketProjectPermissionJobsArgs struct {
	ProjectKeys *[]string
	Status      *string
	Count       *int32
}

type BitbucketProjectsPermissionJobsResolver interface {
	TotalCount() int32
	Nodes() ([]BitbucketProjectsPermissionJobResolver, error)
}

type BitbucketProjectsPermissionJobResolver interface {
	InternalJobID() int32
	State() string
	FailureMessage() *string
	QueuedAt() gqlutil.DateTime
	StartedAt() *gqlutil.DateTime
	FinishedAt() *gqlutil.DateTime
	ProcessAfter() *gqlutil.DateTime
	NumResets() int32
	NumFailures() int32
	ProjectKey() string
	ExternalServiceID() graphql.ID
	Permissions() []UserPermissionResolver
	Unrestricted() bool
}

type UserPermissionResolver interface {
	BindID() string
	Permission() string
}

type PermissionsInfoResolver interface {
	Permissions() []string
	SyncedAt() *gqlutil.DateTime
	UpdatedAt() *gqlutil.DateTime
	Source() *string
	Unrestricted(ctx context.Context) bool
	Repositories(ctx context.Context, args PermissionsInfoRepositoriesArgs) (*gqlutil.ConnectionResolver[PermissionsInfoRepositoryResolver], error)
	Users(ctx context.Context, args PermissionsInfoUsersArgs) (*gqlutil.ConnectionResolver[PermissionsInfoUserResolver], error)
}

type PermissionsInfoRepositoryResolver interface {
	ID() graphql.ID
	Repository(ctx context.Context) (*RepositoryResolver, error)
	Reason() string
	UpdatedAt() *gqlutil.DateTime
}

type PermissionsInfoRepositoriesArgs struct {
	gqlutil.ConnectionResolverArgs
	Query *string
}

type PermissionsInfoUserResolver interface {
	ID() graphql.ID
	User(context.Context) *UserResolver
	Reason() string
	UpdatedAt() *gqlutil.DateTime
}

type PermissionsInfoUsersArgs struct {
	gqlutil.ConnectionResolverArgs
	Query *string
}

const (
	CancelPermissionsSyncJobResultMessageSuccess  CancelPermissionsSyncJobResultMessage = "SUCCESS"
	CancelPermissionsSyncJobResultMessageNotFound CancelPermissionsSyncJobResultMessage = "NOT_FOUND"
	CancelPermissionsSyncJobResultMessageError    CancelPermissionsSyncJobResultMessage = "ERROR"
)

type CancelPermissionsSyncJobResultMessage string

type PermissionsSyncingStatsResolver interface {
	QueueSize(ctx context.Context) (int32, error)
	UsersWithLatestJobFailing(ctx context.Context) (int32, error)
	ReposWithLatestJobFailing(ctx context.Context) (int32, error)
	UsersWithNoPermissions(ctx context.Context) (int32, error)
	ReposWithNoPermissions(ctx context.Context) (int32, error)
	UsersWithStalePermissions(ctx context.Context) (int32, error)
	ReposWithStalePermissions(ctx context.Context) (int32, error)
}
