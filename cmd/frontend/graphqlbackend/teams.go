package graphqlbackend

import (
	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
)

type TeamResolver struct{}

func (r *TeamResolver) ID() graphql.ID {
	return relay.MarshalID("Team", 1)
}

func (r *TeamResolver) UniqueField() *string {
	return nil
}

func (r *TeamResolver) OwnerField() string {
	return EnterpriseResolvers.ownResolver.TeamOwnerField(r)
}
