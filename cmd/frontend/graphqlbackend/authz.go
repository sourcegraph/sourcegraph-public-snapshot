package graphqlbackend

import (
	"context"

	"github.com/graph-gophers/graphql-go"

	"github.com/sourcegraph/sourcegraph/internal/gqlutil"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

type AuthzResolver interface {
	// Mutations
	SetRepositoryPermissionsForUsers(ctx context.Context, args *RepoPermsArgs) (*EmptyResponse, error)
	SetRepositoryPermissionsUnrestricted(ctx context.Context, args *RepoUnrestrictedArgs) (*EmptyResponse, error)
	ScheduleRepositoryPermissionsSync(ctx context.Context, args *RepositoryIDArgs) (*EmptyResponse, error)
	ScheduleUserPermissionsSync(ctx context.Context, args *UserPermissionsSyncArgs) (*EmptyResponse, error)
	SetSubRepositoryPermissionsForUsers(ctx context.Context, args *SubRepoPermsArgs) (*EmptyResponse, error)
	SetRepositoryPermissionsForBitbucketProject(ctx context.Context, args *RepoPermsBitbucketProjectArgs) (*EmptyResponse, error)

	// Queries
	AuthorizedUserRepositories(ctx context.Context, args *AuthorizedRepoArgs) (RepositoryConnectionResolver, error)
	UsersWithPendingPermissions(ctx context.Context) ([]string, error)
	AuthorizedUsers(ctx context.Context, args *RepoAuthorizedUserArgs) (UserConnectionResolver, error)
	BitbucketProjectPermissionJobs(ctx context.Context, args *BitbucketProjectPermissionJobsArgs) (BitbucketProjectsPermissionJobsResolver, error)

	// Helpers
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
	UpdatedAt() gqlutil.DateTime
	Unrestricted() bool
}
