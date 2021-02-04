package graphqlbackend

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/search/searchcontexts"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

type searchContextResolver struct {
	sc types.SearchContext
}

func marshalSearchContextID(searchContextSpec string) graphql.ID {
	return relay.MarshalID("SearchContext", searchContextSpec)
}

func (r searchContextResolver) ID() graphql.ID {
	return marshalSearchContextID(searchcontexts.GetSearchContextSpec(&r.sc))
}

func (r searchContextResolver) Description(ctx context.Context) string {
	return r.sc.Description
}

func (r searchContextResolver) AutoDefined(ctx context.Context) bool {
	return r.sc.ID == 0
}

func (r searchContextResolver) Spec(ctx context.Context) string {
	return searchcontexts.GetSearchContextSpec(&r.sc)
}

func (r *schemaResolver) SearchContexts(ctx context.Context) ([]*searchContextResolver, error) {
	if !envvar.SourcegraphDotComMode() {
		return []*searchContextResolver{}, nil
	}

	searchContextResolvers := []*searchContextResolver{{sc: *searchcontexts.GetGlobalSearchContext()}}
	a := actor.FromContext(ctx)
	if a.IsAuthenticated() {
		user, err := database.GlobalUsers.GetByID(ctx, a.UID)
		if err != nil {
			return nil, err
		}
		searchContext := searchcontexts.GetUserSearchContext(user.Username, a.UID)
		searchContextResolvers = append(searchContextResolvers, &searchContextResolver{sc: *searchContext})
	}
	return searchContextResolvers, nil
}
