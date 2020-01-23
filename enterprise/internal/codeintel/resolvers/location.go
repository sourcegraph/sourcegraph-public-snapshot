package resolvers

import (
	"context"
	"encoding/base64"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/lsif"
)

type locationConnectionResolver struct {
	locations []*lsif.LSIFLocation
	nextURL   string
}

var _ graphqlbackend.LocationConnectionResolver = &locationConnectionResolver{}

func (r *locationConnectionResolver) Nodes(ctx context.Context) ([]graphqlbackend.LocationResolver, error) {
	collectionResolver := &repositoryCollectionResolver{
		commitCollectionResolvers: map[api.RepoID]*commitCollectionResolver{},
	}

	var l []graphqlbackend.LocationResolver
	for _, location := range r.locations {
		treeResolver, err := collectionResolver.resolve(ctx, location.RepositoryID, location.Commit, location.Path)
		if err != nil {
			return nil, err
		}

		if treeResolver == nil {
			continue
		}

		l = append(l, graphqlbackend.NewLocationResolver(
			treeResolver,
			&location.Range,
		))
	}

	return l, nil
}

func (r *locationConnectionResolver) PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error) {
	if r.nextURL != "" {
		return graphqlutil.NextPageCursor(base64.StdEncoding.EncodeToString([]byte(r.nextURL))), nil
	}
	return graphqlutil.HasNextPage(false), nil
}
