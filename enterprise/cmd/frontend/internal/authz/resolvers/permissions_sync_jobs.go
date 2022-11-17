package resolvers

import (
	"context"
	"time"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/authz/syncjobs"
	"github.com/sourcegraph/sourcegraph/internal/gqlutil"
	"github.com/sourcegraph/sourcegraph/lib/errors"
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

const permissionsSyncJobKind = "PermissionsSyncJob"

func getPermissionsSyncJobByIDFunc(r *Resolver) graphqlbackend.NodeByIDFunc {
	return func(ctx context.Context, id graphql.ID) (graphqlbackend.Node, error) {
		if kind := relay.UnmarshalKind(id); kind != permissionsSyncJobKind {
			return nil, errors.Errorf("expected graphql ID to have kind %q; got %q", permissionsSyncJobKind, kind)
		}
		var unixNano int64
		err := relay.UnmarshalSpec(id, &unixNano)
		if err != nil {
			return nil, errors.Wrap(err, "unmarshal ID")
		}
		status, err := r.syncJobsRecords.Get(time.Unix(0, unixNano))
		if err != nil {
			return nil, errors.Wrap(err, "node with ID not found - it may have expired")
		}
		return &permissionsSyncJobResolver{*status}, nil
	}
}

func (j permissionsSyncJobResolver) ID() graphql.ID {
	// Use event time because job ID can repeat
	return relay.MarshalID(permissionsSyncJobKind, j.s.Completed.UnixNano())
}

func (j permissionsSyncJobResolver) JobID() int32 { return j.s.JobID }
func (j permissionsSyncJobResolver) Type() string { return j.s.JobType }
func (j permissionsSyncJobResolver) CompletedAt() *gqlutil.DateTime {
	return &gqlutil.DateTime{Time: j.s.Completed}
}
func (j permissionsSyncJobResolver) Status() string  { return j.s.Status }
func (j permissionsSyncJobResolver) Message() string { return j.s.Message }
func (j permissionsSyncJobResolver) Providers() (providers []graphqlbackend.PermissionsProviderStateResolver, err error) {
	for _, p := range j.s.Providers {
		providers = append(providers, permissionsProviderStatusResolver{ProviderStatus: p})
	}
	return
}

type permissionsProviderStatusResolver struct{ syncjobs.ProviderStatus }

var _ graphqlbackend.PermissionsProviderStateResolver = permissionsProviderStatusResolver{}

func (p permissionsProviderStatusResolver) ID() string      { return p.ProviderID }
func (p permissionsProviderStatusResolver) Type() string    { return p.ProviderType }
func (p permissionsProviderStatusResolver) Status() string  { return p.ProviderStatus.Status }
func (p permissionsProviderStatusResolver) Message() string { return p.ProviderStatus.Message }
