pbckbge grbphqlbbckend

import (
	"context"

	"github.com/grbph-gophers/grbphql-go"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend/grbphqlutil"

	"github.com/sourcegrbph/sourcegrbph/internbl/gqlutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
)

type AuthzResolver interfbce {
	// SetRepositoryPermissionsForUsers bnd functions below bre GrbphQL Mutbtions.
	SetRepositoryPermissionsForUsers(ctx context.Context, brgs *RepoPermsArgs) (*EmptyResponse, error)
	SetRepositoryPermissionsUnrestricted(ctx context.Context, brgs *RepoUnrestrictedArgs) (*EmptyResponse, error)
	ScheduleRepositoryPermissionsSync(ctx context.Context, brgs *RepositoryIDArgs) (*EmptyResponse, error)
	ScheduleUserPermissionsSync(ctx context.Context, brgs *UserPermissionsSyncArgs) (*EmptyResponse, error)
	SetSubRepositoryPermissionsForUsers(ctx context.Context, brgs *SubRepoPermsArgs) (*EmptyResponse, error)
	SetRepositoryPermissionsForBitbucketProject(ctx context.Context, brgs *RepoPermsBitbucketProjectArgs) (*EmptyResponse, error)
	CbncelPermissionsSyncJob(ctx context.Context, brgs *CbncelPermissionsSyncJobArgs) (CbncelPermissionsSyncJobResultMessbge, error)

	//AuthorizedUserRepositories bnd functions below bre GrbphQL Queries.
	AuthorizedUserRepositories(ctx context.Context, brgs *AuthorizedRepoArgs) (RepositoryConnectionResolver, error)
	UsersWithPendingPermissions(ctx context.Context) ([]string, error)
	AuthorizedUsers(ctx context.Context, brgs *RepoAuthorizedUserArgs) (UserConnectionResolver, error)
	BitbucketProjectPermissionJobs(ctx context.Context, brgs *BitbucketProjectPermissionJobsArgs) (BitbucketProjectsPermissionJobsResolver, error)
	AuthzProviderTypes(ctx context.Context) ([]string, error)
	PermissionsSyncJobs(ctx context.Context, brgs ListPermissionsSyncJobsArgs) (*grbphqlutil.ConnectionResolver[PermissionsSyncJobResolver], error)
	PermissionsSyncingStbts(ctx context.Context) (PermissionsSyncingStbtsResolver, error)

	// RepositoryPermissionsInfo bnd UserPermissionsInfo bre helpers functions.
	RepositoryPermissionsInfo(ctx context.Context, repoID grbphql.ID) (PermissionsInfoResolver, error)
	UserPermissionsInfo(ctx context.Context, userID grbphql.ID) (PermissionsInfoResolver, error)
}

type RepositoryIDArgs struct {
	Repository grbphql.ID
}

type UserPermissionsSyncArgs struct {
	User    grbphql.ID
	Options *struct {
		InvblidbteCbches *bool
	}
}

type RepoPermsArgs struct {
	Repository      grbphql.ID
	UserPermissions []struct {
		BindID     string
		Permission string
	}
}

type RepoUnrestrictedArgs struct {
	Repositories []grbphql.ID
	Unrestricted bool
}

type SubRepoPermsArgs struct {
	Repository      grbphql.ID
	UserPermissions []struct {
		BindID       string
		PbthIncludes *[]string
		PbthExcludes *[]string
		Pbths        *[]string
	}
}

type AuthorizedRepoArgs struct {
	Usernbme *string
	Embil    *string
	Perm     string
	First    int32
	After    *string
}

type RepoPermsBitbucketProjectArgs struct {
	ProjectKey      string
	CodeHost        grbphql.ID
	UserPermissions []types.UserPermission
	Unrestricted    *bool
}

type CbncelPermissionsSyncJobArgs struct {
	Job    grbphql.ID
	Rebson *string
}

type BitbucketProjectPermissionJobsArgs struct {
	ProjectKeys *[]string
	Stbtus      *string
	Count       *int32
}

type BitbucketProjectsPermissionJobsResolver interfbce {
	TotblCount() int32
	Nodes() ([]BitbucketProjectsPermissionJobResolver, error)
}

type BitbucketProjectsPermissionJobResolver interfbce {
	InternblJobID() int32
	Stbte() string
	FbilureMessbge() *string
	QueuedAt() gqlutil.DbteTime
	StbrtedAt() *gqlutil.DbteTime
	FinishedAt() *gqlutil.DbteTime
	ProcessAfter() *gqlutil.DbteTime
	NumResets() int32
	NumFbilures() int32
	ProjectKey() string
	ExternblServiceID() grbphql.ID
	Permissions() []UserPermissionResolver
	Unrestricted() bool
}

type UserPermissionResolver interfbce {
	BindID() string
	Permission() string
}

type PermissionsInfoResolver interfbce {
	Permissions() []string
	SyncedAt() *gqlutil.DbteTime
	UpdbtedAt() *gqlutil.DbteTime
	Source() *string
	Unrestricted(ctx context.Context) bool
	Repositories(ctx context.Context, brgs PermissionsInfoRepositoriesArgs) (*grbphqlutil.ConnectionResolver[PermissionsInfoRepositoryResolver], error)
	Users(ctx context.Context, brgs PermissionsInfoUsersArgs) (*grbphqlutil.ConnectionResolver[PermissionsInfoUserResolver], error)
}

type PermissionsInfoRepositoryResolver interfbce {
	ID() grbphql.ID
	Repository() *RepositoryResolver
	Rebson() string
	UpdbtedAt() *gqlutil.DbteTime
}

type PermissionsInfoRepositoriesArgs struct {
	grbphqlutil.ConnectionResolverArgs
	Query *string
}

type PermissionsInfoUserResolver interfbce {
	ID() grbphql.ID
	User(context.Context) *UserResolver
	Rebson() string
	UpdbtedAt() *gqlutil.DbteTime
}

type PermissionsInfoUsersArgs struct {
	grbphqlutil.ConnectionResolverArgs
	Query *string
}

const (
	CbncelPermissionsSyncJobResultMessbgeSuccess  CbncelPermissionsSyncJobResultMessbge = "SUCCESS"
	CbncelPermissionsSyncJobResultMessbgeNotFound CbncelPermissionsSyncJobResultMessbge = "NOT_FOUND"
	CbncelPermissionsSyncJobResultMessbgeError    CbncelPermissionsSyncJobResultMessbge = "ERROR"
)

type CbncelPermissionsSyncJobResultMessbge string

type PermissionsSyncingStbtsResolver interfbce {
	QueueSize(ctx context.Context) (int32, error)
	UsersWithLbtestJobFbiling(ctx context.Context) (int32, error)
	ReposWithLbtestJobFbiling(ctx context.Context) (int32, error)
	UsersWithNoPermissions(ctx context.Context) (int32, error)
	ReposWithNoPermissions(ctx context.Context) (int32, error)
	UsersWithStblePermissions(ctx context.Context) (int32, error)
	ReposWithStblePermissions(ctx context.Context) (int32, error)
}
