package resolvers

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
)

type catalogEntityStatusResolver struct {
	contexts []gql.CatalogEntityStatusContextResolver
	entityID graphql.ID
}

func (r *catalogEntityStatusResolver) ID() graphql.ID {
	return relay.MarshalID("CatalogEntityStatus", string(r.entityID)) // TODO
}

func (r *catalogEntityStatusResolver) Contexts() []gql.CatalogEntityStatusContextResolver {
	return r.contexts
}

func (r *catalogEntityStatusResolver) State() gql.CatalogEntityStatusState {
	for _, sc := range r.contexts {
		switch sc.State() {
		case "FAILURE", "ERROR", "PENDING", "EXPECTED":
			return sc.State()
		}
	}
	return "SUCCESS"
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
func (r *catalogEntityStatusContextResolver) Description() *string {
	if r.description == "" {
		return nil
	}
	return &r.description
}
func (r *catalogEntityStatusContextResolver) TargetURL() *string {
	if r.targetURL == "" {
		return nil
	}
	return &r.targetURL
}
