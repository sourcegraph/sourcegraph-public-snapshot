package resolvers

import (
	"context"

	"github.com/sourcegraph/go-langserver/pkg/lsp"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/lsif"
)

type locationWithSourceCommit struct {
	*lsif.LSIFLocation
	uploadCommit string
}

type locationConnectionResolver struct {
	repo      *types.Repo
	commit    graphqlbackend.GitObjectID
	locations []locationWithSourceCommit
	endCursor string
}

var _ graphqlbackend.LocationConnectionResolver = &locationConnectionResolver{}

func (r *locationConnectionResolver) Nodes(ctx context.Context) ([]graphqlbackend.LocationResolver, error) {
	collectionResolver := &repositoryCollectionResolver{
		commitCollectionResolvers: map[api.RepoID]*commitCollectionResolver{},
	}

	var l []graphqlbackend.LocationResolver
	for _, location := range r.locations {
		adjustedCommit, adjustedRange, err := r.adjustLocation(ctx, location)
		if err != nil {
			return nil, err
		}

		treeResolver, err := collectionResolver.resolve(ctx, location.RepositoryID, adjustedCommit, location.Path)
		if err != nil {
			return nil, err
		}

		if treeResolver == nil {
			continue
		}

		l = append(l, graphqlbackend.NewLocationResolver(treeResolver, &adjustedRange))
	}

	return l, nil
}

func (r *locationConnectionResolver) PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error) {
	if r.endCursor != "" {
		return graphqlutil.NextPageCursor(r.endCursor), nil
	}
	return graphqlutil.HasNextPage(false), nil
}

// locationConnectionResolver transforms the given position in the location's commit into
// a position in the commit the user original requested. This method will return the given
// position untouched if the location is not in the repo where the request originated.
func (r *locationConnectionResolver) adjustLocation(ctx context.Context, location locationWithSourceCommit) (string, lsp.Range, error) {
	if location.RepositoryID != r.repo.ID {
		return location.Commit, location.Range, nil
	}

	adjuster, err := newPositionAdjuster(ctx, r.repo, location.Commit, string(r.commit), location.Path)
	if err != nil {
		return "", lsp.Range{}, err
	}

	if adjustedRange, ok := adjuster.adjustRange(location.Range); ok {
		return string(r.commit), adjustedRange, nil
	}

	// Couldn't adjust range, return original result which is precise but
	// jump the user to another into another commit context on navigation.
	return location.Commit, location.Range, nil
}
