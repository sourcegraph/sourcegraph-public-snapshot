package resolvers

import (
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/authz/syncjobs"
	"github.com/sourcegraph/sourcegraph/internal/gqlutil"
)

type permissionsSyncJobsConnection struct {
	jobs []graphqlbackend.PermissionsSyncJob
}

var _ graphqlbackend.PermissionsSyncJobsConnection = &permissionsSyncJobsConnection{}

func (r *permissionsSyncJobsConnection) Nodes() []graphqlbackend.PermissionsSyncJob {
	return r.jobs
}

// We don't yet support pagination, but we have the fields for future-compat
func (r *permissionsSyncJobsConnection) TotalCount() int32 { return int32(len(r.jobs)) }
func (r *permissionsSyncJobsConnection) PageInfo() *graphqlutil.PageInfo {
	return graphqlutil.HasNextPage(false)
}

type permissionsSyncJobResolver struct{ s syncjobs.Status }

var _ graphqlbackend.PermissionsSyncJob = permissionsSyncJobResolver{}

func (j permissionsSyncJobResolver) ID() int32 { return j.s.RequestID }

func (j permissionsSyncJobResolver) Type() string { return j.s.RequestType }

func (j permissionsSyncJobResolver) CompletedAt() *gqlutil.DateTime {
	return &gqlutil.DateTime{Time: j.s.Completed}
}

func (j permissionsSyncJobResolver) Status() string { return j.s.Status }

func (j permissionsSyncJobResolver) Message() string { return j.s.Message }

func (j permissionsSyncJobResolver) Providers() (providers []graphqlbackend.PermissionsProviderStatus, err error) {
	for _, p := range j.s.Providers {
		providers = append(providers, permissionsProviderStatusResolver{ProviderStatus: p})
	}
	return
}

type permissionsProviderStatusResolver struct{ syncjobs.ProviderStatus }

var _ graphqlbackend.PermissionsProviderStatus = permissionsProviderStatusResolver{}

func (p permissionsProviderStatusResolver) ID() string { return p.ProviderID }

func (p permissionsProviderStatusResolver) Type() string { return p.ProviderType }

func (p permissionsProviderStatusResolver) Status() string { return p.ProviderStatus.Status }

func (p permissionsProviderStatusResolver) Message() string { return p.ProviderStatus.Message }
