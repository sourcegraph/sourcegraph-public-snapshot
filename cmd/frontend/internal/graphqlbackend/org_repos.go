package graphqlbackend

import (
	"context"
	"time"

	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	store "sourcegraph.com/sourcegraph/sourcegraph/pkg/localstore"
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
	File  *string
	Limit *int32
}) ([]*threadResolver, error) {
	limit := int32(1000)
	if args.Limit != nil && *args.Limit < limit {
		limit = *args.Limit
	}
	var threads []*sourcegraph.Thread
	var err error
	if args.File != nil {
		threads, err = store.Threads.GetAllForFile(ctx, o.repo.ID, *args.File, limit)
	} else {
		threads, err = store.Threads.GetAllForRepo(ctx, o.repo.ID, limit)
	}
	if err != nil {
		return nil, err
	}
	threadResolvers := []*threadResolver{}
	for _, thread := range threads {
		threadResolvers = append(threadResolvers, &threadResolver{o.org, o.repo, thread})
	}
	return threadResolvers, nil
}
