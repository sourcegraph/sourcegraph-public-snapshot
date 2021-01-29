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

func marshalSearchContextID(searchContextID int32) graphql.ID {
	return relay.MarshalID("SearchContext", searchContextID)
}

func unmarshalSearchContextID(id graphql.ID) (searchContextID int32, err error) {
	err = relay.UnmarshalSpec(id, &searchContextID)
	return
}

func (r searchContextResolver) ID() graphql.ID {
	return marshalSearchContextID(r.sc.ID)
}

func (r searchContextResolver) Name() string {
	return r.sc.Name
}

func (r searchContextResolver) Namespace(ctx context.Context) (*NamespaceResolver, error) {
	if r.sc.OrgID != nil {
		n, err := NamespaceByID(ctx, MarshalOrgID(*r.sc.OrgID))
		if err != nil {
			return nil, err
		}
		return &NamespaceResolver{n}, nil
	}
	if r.sc.UserID != nil {
		n, err := NamespaceByID(ctx, MarshalUserID(*r.sc.UserID))
		if err != nil {
			return nil, err
		}
		return &NamespaceResolver{n}, nil
	}
	return nil, nil
}

func (r searchContextResolver) AutoDefined(ctx context.Context) bool {
	return r.sc.ID == 0
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
		searchContext := types.SearchContext{Name: user.Username, UserID: &a.UID}
		searchContextResolvers = append(searchContextResolvers, &searchContextResolver{sc: searchContext})
	}
	return searchContextResolvers, nil
}
