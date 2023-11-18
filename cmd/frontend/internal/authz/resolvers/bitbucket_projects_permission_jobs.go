package resolvers

import (
	"strings"

	"github.com/graph-gophers/graphql-go"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/gqlutil"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

type bitbucketProjectsPermissionJobsResolver struct {
	jobs []*types.BitbucketProjectPermissionJob
}

func NewBitbucketProjectsPermissionJobsResolver(jobs []*types.BitbucketProjectPermissionJob) graphqlbackend.BitbucketProjectsPermissionJobsResolver {
	return &bitbucketProjectsPermissionJobsResolver{
		jobs: jobs,
	}
}

func (b bitbucketProjectsPermissionJobsResolver) TotalCount() int32 {
	return int32(len(b.jobs))
}

func (b bitbucketProjectsPermissionJobsResolver) Nodes() ([]graphqlbackend.BitbucketProjectsPermissionJobResolver, error) {
	resolvers := make([]graphqlbackend.BitbucketProjectsPermissionJobResolver, 0, len(b.jobs))
	for _, job := range b.jobs {
		resolvers = append(resolvers, convertJobToResolver(job))
	}
	return resolvers, nil
}

func convertJobToResolver(job *types.BitbucketProjectPermissionJob) graphqlbackend.BitbucketProjectsPermissionJobResolver {
	return bitbucketProjectsPermissionJobResolver{job: *job}
}

type bitbucketProjectsPermissionJobResolver struct {
	job types.BitbucketProjectPermissionJob
}

func (j bitbucketProjectsPermissionJobResolver) InternalJobID() int32 {
	return int32(j.job.ID)
}

func (j bitbucketProjectsPermissionJobResolver) State() string {
	return j.job.State
}

func (j bitbucketProjectsPermissionJobResolver) FailureMessage() *string {
	return j.job.FailureMessage
}

func (j bitbucketProjectsPermissionJobResolver) QueuedAt() gqlutil.DateTime {
	return gqlutil.DateTime{Time: j.job.QueuedAt}
}

func (j bitbucketProjectsPermissionJobResolver) StartedAt() *gqlutil.DateTime {
	return gqlutil.DateTimeOrNil(j.job.StartedAt)
}

func (j bitbucketProjectsPermissionJobResolver) FinishedAt() *gqlutil.DateTime {
	return gqlutil.DateTimeOrNil(j.job.FinishedAt)
}

func (j bitbucketProjectsPermissionJobResolver) ProcessAfter() *gqlutil.DateTime {
	return gqlutil.DateTimeOrNil(j.job.ProcessAfter)
}

func (j bitbucketProjectsPermissionJobResolver) NumResets() int32 {
	return int32(j.job.NumResets)
}

func (j bitbucketProjectsPermissionJobResolver) NumFailures() int32 {
	return int32(j.job.NumFailures)
}

func (j bitbucketProjectsPermissionJobResolver) ProjectKey() string {
	return j.job.ProjectKey
}

func (j bitbucketProjectsPermissionJobResolver) ExternalServiceID() graphql.ID {
	return graphqlbackend.MarshalExternalServiceID(j.job.ExternalServiceID)
}

func (j bitbucketProjectsPermissionJobResolver) Permissions() []graphqlbackend.UserPermissionResolver {
	return permissionsToPermissionResolvers(j.job.Permissions)
}

func (j bitbucketProjectsPermissionJobResolver) Unrestricted() bool {
	return j.job.Unrestricted
}

type userPermissionResolver struct {
	bindID     string
	permission string
}

func NewUserPermissionResolver(bindID, permission string) graphqlbackend.UserPermissionResolver {
	return userPermissionResolver{
		bindID:     bindID,
		permission: permission,
	}
}

func permissionsToPermissionResolvers(perms []types.UserPermission) []graphqlbackend.UserPermissionResolver {
	resolvers := make([]graphqlbackend.UserPermissionResolver, 0, len(perms))
	for _, perm := range perms {
		resolvers = append(resolvers, NewUserPermissionResolver(perm.BindID, perm.Permission))
	}
	return resolvers
}

func (u userPermissionResolver) BindID() string {
	return u.bindID
}

func (u userPermissionResolver) Permission() string {
	return strings.ToUpper(u.permission)
}
