package resolvers

import (
	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/authz/syncjobs"
	"github.com/sourcegraph/sourcegraph/internal/gqlutil"
)

type permissionsSyncJobsConnection struct {
	jobs []graphqlbackend.PermissionsSyncJobResolver
}

var _ graphqlbackend.PermissionsSyncJobsConnection = &permissionsSyncJobsConnection{}

func (r *permissionsSyncJobsConnection) Nodes() []graphqlbackend.PermissionsSyncJobResolver {
	return r.jobs
}

// We don't yet support pagination, but we have the fields for future-compat
func (r *permissionsSyncJobsConnection) TotalCount() int32 { return int32(len(r.jobs)) }
func (r *permissionsSyncJobsConnection) PageInfo() *graphqlutil.PageInfo {
	return graphqlutil.HasNextPage(false)
}

type permissionsSyncJobResolver struct{ s syncjobs.Status }

var _ graphqlbackend.PermissionsSyncJobResolver = permissionsSyncJobResolver{}

func (j permissionsSyncJobResolver) ID() graphql.ID {
	// Use eventTime because job ID can repeat
	return relay.MarshalID("PermissionsSyncJob", j.s.Completed)
}

func (j permissionsSyncJobResolver) JobID() int32 { return j.s.JobID }

func (j permissionsSyncJobResolver) Type() string { return j.s.JobType }

func (j permissionsSyncJobResolver) CompletedAt() *gqlutil.DateTime {
	return &gqlutil.DateTime{Time: j.s.Completed}
}

func (j permissionsSyncJobResolver) Status() string { return j.s.Status }

func (j permissionsSyncJobResolver) Message() string { return j.s.Message }

func (j permissionsSyncJobResolver) Providers() (providers []graphqlbackend.PermissionsProviderStateResolver, err error) {
	for _, p := range j.s.Providers {
		providers = append(providers, permissionsProviderStatusResolver{ProviderStatus: p})
	}
	return
}

type permissionsProviderStatusResolver struct{ syncjobs.ProviderStatus }

var _ graphqlbackend.PermissionsProviderStateResolver = permissionsProviderStatusResolver{}

func (p permissionsProviderStatusResolver) ID() string { return p.ProviderID }

func (p permissionsProviderStatusResolver) Type() string { return p.ProviderType }

func (p permissionsProviderStatusResolver) Status() string { return p.ProviderStatus.Status }

func (p permissionsProviderStatusResolver) Message() string { return p.ProviderStatus.Message }
