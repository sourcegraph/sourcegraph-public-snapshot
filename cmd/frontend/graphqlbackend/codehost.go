package graphqlbackend

import (
	"fmt"
	graphql "github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"time"
)

type codehostResolver struct {
	codehost *types.Codehost
}

const codehostIDKind = "Codehost"

func marshalCodehostID(id int64) graphql.ID {
	return relay.MarshalID(codehostIDKind, id)
}

func unmarshalCodehostID(id graphql.ID) (codehostID int64, err error) {
	if kind := relay.UnmarshalKind(id); kind != codehostIDKind {
		err = fmt.Errorf("expected graphql ID to have kind %q; got %q", codehostIDKind, kind)
		return
	}
	err = relay.UnmarshalSpec(id, &codehostID)
	return
}

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
