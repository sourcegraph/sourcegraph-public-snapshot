package graphqlbackend

import (
	"context"
	"errors"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/search/searchcontexts"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/schema"
)

type searchContextResolver struct {
	sc *types.SearchContext

	db dbutil.DB
}

func marshalSearchContextID(searchContextSpec string) graphql.ID {
	return relay.MarshalID("SearchContext", searchContextSpec)
}

func (r searchContextResolver) ID() graphql.ID {
	return marshalSearchContextID(searchcontexts.GetSearchContextSpec(r.sc))
}

func (r searchContextResolver) Description(ctx context.Context) string {
	return r.sc.Description
}

func (r searchContextResolver) AutoDefined(ctx context.Context) bool {
	return searchcontexts.IsAutoDefinedSearchContext(r.sc)
}

func (r searchContextResolver) Spec(ctx context.Context) string {
	return searchcontexts.GetSearchContextSpec(r.sc)
}

func (r *schemaResolver) SearchContexts(ctx context.Context) ([]*searchContextResolver, error) {
	searchContexts, err := searchcontexts.GetUsersSearchContexts(ctx, r.db)
	if err != nil {
		return nil, err
	}

	searchContextResolvers := make([]*searchContextResolver, len(searchContexts))
	for idx, searchContext := range searchContexts {
		searchContextResolvers[idx] = &searchContextResolver{searchContext, r.db}
	}
	return searchContextResolvers, nil
}

func resolveVersionContext(versionContext string) (*schema.VersionContext, error) {
	for _, vc := range conf.Get().ExperimentalFeatures.VersionContexts {
		if vc.Name == versionContext {
			return vc, nil
		}
	}
	return nil, errors.New("version context not found")
}

func (r *schemaResolver) ConvertVersionContextToSearchContext(ctx context.Context, args *struct {
	Name string
}) (*searchContextResolver, error) {
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return nil, errors.New("converting a version context to a search context is limited to site admins")
	}
	versionContext, err := resolveVersionContext(args.Name)
	if err != nil {
		return nil, err
	}

	searchContext, err := searchcontexts.ConvertVersionContextToSearchContext(ctx, r.db, versionContext)
	if err != nil {
		return nil, err
	}
	return &searchContextResolver{searchContext, r.db}, nil
}
