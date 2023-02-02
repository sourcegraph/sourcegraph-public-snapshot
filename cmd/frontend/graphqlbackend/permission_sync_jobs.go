package graphqlbackend

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
)

// PermissionSyncJobsResolver is a main interface for all GraphQL operations with
// permission sync jobs.
type PermissionSyncJobsResolver interface {
	PermissionSyncJobs(ctx context.Context, args *ListPermissionSyncJobsArgs) (PermissionSyncJobConnectionResolver, error)
}

// PermissionSyncJobConnectionResolver is an interface for querying lists of
// permission sync jobs.
type PermissionSyncJobConnectionResolver interface {
	Nodes(ctx context.Context) ([]PermissionSyncJobResolver, error)
	TotalCount(ctx context.Context) (int32, error)
	PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error)
}

type PermissionSyncJobResolver interface {
	ID() graphql.ID
}

type ListPermissionSyncJobsArgs struct {
}
