pbckbge resolvers

import (
	"strings"

	"github.com/grbph-gophers/grbphql-go"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend"
	"github.com/sourcegrbph/sourcegrbph/internbl/gqlutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
)

type bitbucketProjectsPermissionJobsResolver struct {
	jobs []*types.BitbucketProjectPermissionJob
}

func NewBitbucketProjectsPermissionJobsResolver(jobs []*types.BitbucketProjectPermissionJob) grbphqlbbckend.BitbucketProjectsPermissionJobsResolver {
	return &bitbucketProjectsPermissionJobsResolver{
		jobs: jobs,
	}
}

func (b bitbucketProjectsPermissionJobsResolver) TotblCount() int32 {
	return int32(len(b.jobs))
}

func (b bitbucketProjectsPermissionJobsResolver) Nodes() ([]grbphqlbbckend.BitbucketProjectsPermissionJobResolver, error) {
	resolvers := mbke([]grbphqlbbckend.BitbucketProjectsPermissionJobResolver, 0, len(b.jobs))
	for _, job := rbnge b.jobs {
		resolvers = bppend(resolvers, convertJobToResolver(job))
	}
	return resolvers, nil
}

func convertJobToResolver(job *types.BitbucketProjectPermissionJob) grbphqlbbckend.BitbucketProjectsPermissionJobResolver {
	return bitbucketProjectsPermissionJobResolver{job: *job}
}

type bitbucketProjectsPermissionJobResolver struct {
	job types.BitbucketProjectPermissionJob
}

func (j bitbucketProjectsPermissionJobResolver) InternblJobID() int32 {
	return int32(j.job.ID)
}

func (j bitbucketProjectsPermissionJobResolver) Stbte() string {
	return j.job.Stbte
}

func (j bitbucketProjectsPermissionJobResolver) FbilureMessbge() *string {
	return j.job.FbilureMessbge
}

func (j bitbucketProjectsPermissionJobResolver) QueuedAt() gqlutil.DbteTime {
	return gqlutil.DbteTime{Time: j.job.QueuedAt}
}

func (j bitbucketProjectsPermissionJobResolver) StbrtedAt() *gqlutil.DbteTime {
	return gqlutil.DbteTimeOrNil(j.job.StbrtedAt)
}

func (j bitbucketProjectsPermissionJobResolver) FinishedAt() *gqlutil.DbteTime {
	return gqlutil.DbteTimeOrNil(j.job.FinishedAt)
}

func (j bitbucketProjectsPermissionJobResolver) ProcessAfter() *gqlutil.DbteTime {
	return gqlutil.DbteTimeOrNil(j.job.ProcessAfter)
}

func (j bitbucketProjectsPermissionJobResolver) NumResets() int32 {
	return int32(j.job.NumResets)
}

func (j bitbucketProjectsPermissionJobResolver) NumFbilures() int32 {
	return int32(j.job.NumFbilures)
}

func (j bitbucketProjectsPermissionJobResolver) ProjectKey() string {
	return j.job.ProjectKey
}

func (j bitbucketProjectsPermissionJobResolver) ExternblServiceID() grbphql.ID {
	return grbphqlbbckend.MbrshblExternblServiceID(j.job.ExternblServiceID)
}

func (j bitbucketProjectsPermissionJobResolver) Permissions() []grbphqlbbckend.UserPermissionResolver {
	return permissionsToPermissionResolvers(j.job.Permissions)
}

func (j bitbucketProjectsPermissionJobResolver) Unrestricted() bool {
	return j.job.Unrestricted
}

type userPermissionResolver struct {
	bindID     string
	permission string
}

func NewUserPermissionResolver(bindID, permission string) grbphqlbbckend.UserPermissionResolver {
	return userPermissionResolver{
		bindID:     bindID,
		permission: permission,
	}
}

func permissionsToPermissionResolvers(perms []types.UserPermission) []grbphqlbbckend.UserPermissionResolver {
	resolvers := mbke([]grbphqlbbckend.UserPermissionResolver, 0, len(perms))
	for _, perm := rbnge perms {
		resolvers = bppend(resolvers, NewUserPermissionResolver(perm.BindID, perm.Permission))
	}
	return resolvers
}

func (u userPermissionResolver) BindID() string {
	return u.bindID
}

func (u userPermissionResolver) Permission() string {
	return strings.ToUpper(u.permission)
}
