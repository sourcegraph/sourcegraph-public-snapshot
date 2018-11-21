package graphqlbackend

import (
	"fmt"
	graphql "github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"time"
)

type externalServiceResolver struct {
	externalService *types.ExternalService
}

const externalServiceIDKind = "ExternalService"

func marshalExternalServiceID(id int64) graphql.ID {
	return relay.MarshalID(externalServiceIDKind, id)
}

func unmarshalExternalServiceID(id graphql.ID) (externalServiceID int64, err error) {
	if kind := relay.UnmarshalKind(id); kind != externalServiceIDKind {
		err = fmt.Errorf("expected graphql ID to have kind %q; got %q", externalServiceIDKind, kind)
		return
	}
	err = relay.UnmarshalSpec(id, &externalServiceID)
	return
}

func (r *externalServiceResolver) ID() graphql.ID {
	return marshalExternalServiceID(r.externalService.ID)
}

func (r *externalServiceResolver) Kind() string {
	return r.externalService.Kind
}

func (r *externalServiceResolver) DisplayName() string {
	return r.externalService.DisplayName
}

func (r *externalServiceResolver) Config() string {
	return r.externalService.Config
}

func (r *externalServiceResolver) CreatedAt() string {
	return r.externalService.CreatedAt.Format(time.RFC3339)
}

func (r *externalServiceResolver) UpdatedAt() string {
	return r.externalService.UpdatedAt.Format(time.RFC3339)
}
