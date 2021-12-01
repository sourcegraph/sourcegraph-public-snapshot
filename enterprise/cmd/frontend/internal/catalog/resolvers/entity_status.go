package resolvers

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"fmt"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
)

func (r *catalogComponentResolver) Status(ctx context.Context) (gql.CatalogEntityStatusResolver, error) {
	return &catalogEntityStatusResolver{
		component: r,
	}, nil
}

type catalogEntityStatusResolver struct {
	component *catalogComponentResolver
}

func (r *catalogEntityStatusResolver) ID() graphql.ID {
	return relay.MarshalID("CatalogEntityStatus", r.component.Name())
}

func (r *catalogEntityStatusResolver) Contexts() []gql.CatalogEntityStatusContextResolver {
	return []gql.CatalogEntityStatusContextResolver{
		&catalogEntityStatusContextResolver{
			name:        "deploy",
			state:       "SUCCESS",
			title:       "Deploy",
			description: "Deployed `f38ca7d` to Sourcegraph.com 4 min ago ([monitor](#TODO))",
		},
		&catalogEntityStatusContextResolver{
			name:        "ci",
			state:       "SUCCESS",
			title:       "CI",
			description: "Build `f38ca7d` passed 7 min ago",
			targetURL:   "https://example.com",
		},
	}
}

type catalogEntityStatusContextResolver struct {
	name, title, description, targetURL string
	state                               gql.CatalogEntityStatusState
}

func (r *catalogEntityStatusContextResolver) ID() graphql.ID {
	b := sha256.Sum256([]byte(fmt.Sprintf("%s:%s:%s", r.name, r.title, r.description)))
	return relay.MarshalID("CatalogEntityStatus", base64.RawURLEncoding.EncodeToString(b[:16]))
}

func (r *catalogEntityStatusContextResolver) Name() string                        { return r.name }
func (r *catalogEntityStatusContextResolver) State() gql.CatalogEntityStatusState { return r.state }
func (r *catalogEntityStatusContextResolver) Title() string                       { return r.title }
func (r *catalogEntityStatusContextResolver) Description() string                 { return r.description }
func (r *catalogEntityStatusContextResolver) TargetURL() *string {
	if r.targetURL == "" {
		return nil
	}
	return &r.targetURL
}
