package resolvers

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
)

type componentStatusResolver struct {
	contexts []gql.ComponentStatusContextResolver
	entityID graphql.ID
}

func (r *componentStatusResolver) ID() graphql.ID {
	return relay.MarshalID("ComponentStatus", string(r.entityID)) // TODO
}

func (r *componentStatusResolver) Contexts() []gql.ComponentStatusContextResolver {
	return r.contexts
}

func (r *componentStatusResolver) State() gql.ComponentStatusState {
	for _, sc := range r.contexts {
		switch sc.State() {
		case "FAILURE", "ERROR", "PENDING", "EXPECTED":
			return sc.State()
		}
	}
	return "SUCCESS"
}

type componentStatusContextResolver struct {
	name, title, description, targetURL string
	state                               gql.ComponentStatusState
}

func (r *componentStatusContextResolver) ID() graphql.ID {
	b := sha256.Sum256([]byte(fmt.Sprintf("%s:%s:%s", r.name, r.title, r.description)))
	return relay.MarshalID("ComponentStatus", base64.RawURLEncoding.EncodeToString(b[:16]))
}

func (r *componentStatusContextResolver) Name() string                        { return r.name }
func (r *componentStatusContextResolver) State() gql.ComponentStatusState { return r.state }
func (r *componentStatusContextResolver) Title() string                       { return r.title }
func (r *componentStatusContextResolver) Description() *string {
	if r.description == "" {
		return nil
	}
	return &r.description
}
func (r *componentStatusContextResolver) TargetURL() *string {
	if r.targetURL == "" {
		return nil
	}
	return &r.targetURL
}
