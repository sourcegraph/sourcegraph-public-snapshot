package graphqlbackend

import (
	graphql "github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"time"
)

type codehostResolver struct {
	codehost *types.Codehost
}

func marshalCodehostID(id int64) graphql.ID { return relay.MarshalID("Codehost", id) }

// func unmarshalCodehostID(id graphql.ID) (codehostID int64, err error) {
// 	err = relay.UnmarshalSpec(id, &codehostID)
// 	return
// }

func (r *codehostResolver) ID() graphql.ID {
	return marshalCodehostID(r.codehost.ID)
}

func (r *codehostResolver) Kind() string {
	return r.codehost.Kind
}

func (r *codehostResolver) DisplayName() string {
	return r.codehost.DisplayName
}

func (r *codehostResolver) Config() string {
	return r.codehost.Config
}

func (r *codehostResolver) CreatedAt() string {
	return r.codehost.CreatedAt.Format(time.RFC3339)
}

func (r *codehostResolver) UpdatedAt() string {
	return r.codehost.UpdatedAt.Format(time.RFC3339)
}
