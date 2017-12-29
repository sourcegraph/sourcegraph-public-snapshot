package graphqlbackend

import (
	"context"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/backend"
)

func (r *schemaResolver) Orgs(ctx context.Context) (*orgConnectionResolver, error) {
	orgs, err := backend.Orgs.List(ctx)
	if err != nil {
		return nil, err
	}
	var resolvers []*orgResolver
	for _, org := range orgs {
		resolvers = append(resolvers, &orgResolver{org: org})
	}
	return &orgConnectionResolver{orgs: resolvers}, nil
}

type orgConnectionResolver struct {
	orgs []*orgResolver
}

func (r *orgConnectionResolver) Nodes() []*orgResolver { return r.orgs }

func (r *orgConnectionResolver) TotalCount() int32 { return int32(len(r.orgs)) }
