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

// DEPRECATED: to be replaced by CanonicalRemoteID.
func (o *orgRepoResolver) RemoteURI() string {
	return o.repo.CanonicalRemoteID
}

func (o *orgRepoResolver) CanonicalRemoteID() string {
	return o.repo.CanonicalRemoteID
}

func (o *orgRepoResolver) CloneURL() string {
	return o.repo.CloneURL
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
	return &threadConnectionResolver{o.org, o.repo, args.File, args.Branch, args.Limit}
}
