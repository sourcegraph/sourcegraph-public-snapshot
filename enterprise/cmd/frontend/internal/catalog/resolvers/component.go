package resolvers

import (
	"sync"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/catalog"
	"github.com/sourcegraph/sourcegraph/internal/database"
)

type componentResolver struct {
	component catalog.Component
	db        database.DB

	// Cached source locations
	sourceLocationsOnce   sync.Once
	sourceLocationsCached []*componentSourceLocationResolver
	sourceLocationsErr    error
}

func (r *componentResolver) ID() graphql.ID {
	return relay.MarshalID("Component", r.component.Name) // TODO(sqs)
}

func (r *componentResolver) Name() string {
	return r.component.Name
}

func (r *componentResolver) Description() *string {
	if r.component.Description == "" {
		return nil
	}
	return &r.component.Description
}

func (r *componentResolver) Lifecycle() gql.ComponentLifecycle {
	return gql.ComponentLifecycle(r.component.Lifecycle)
}

func (r *componentResolver) URL() string {
	return "/catalog/components/" + string(r.Name())
}

func (r *componentResolver) Kind() gql.ComponentKind {
	return gql.ComponentKind(r.component.Kind)
}
