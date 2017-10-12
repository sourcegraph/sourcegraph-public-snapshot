package graphqlbackend

import (
	"context"
	"time"

	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
)

type orgRepoResolver struct {
	org  *sourcegraph.Org
	repo *sourcegraph.OrgRepo
}

func (o *orgRepoResolver) ID() int32 {
	return o.repo.ID
}

func (o *orgRepoResolver) Org() *orgResolver {
	return &orgResolver{o.org}
}

func (o *orgRepoResolver) RemoteURI() string {
	return o.repo.RemoteURI
}

func (o *orgRepoResolver) CreatedAt() string {
	return o.repo.CreatedAt.Format(time.RFC3339)
}

func (o *orgRepoResolver) UpdatedAt() string {
	return o.repo.UpdatedAt.Format(time.RFC3339)
}

func (o *orgRepoResolver) Threads(ctx context.Context, args *struct {
	File   *string
	Branch *string
	Limit  *int32
}) ([]*threadResolver, error) {
	return o.Threads2(ctx, args).Nodes(ctx)
}

func (o *orgRepoResolver) Threads2(ctx context.Context, args *struct {
	File   *string
	Branch *string
	Limit  *int32
}) *threadConnectionResolver {
	var limit int32
	if args.Limit != nil {
		limit = *args.Limit
	}
	return &threadConnectionResolver{repoID: &o.repo.ID, file: args.File, branch: args.Branch, limit: limit}
}
