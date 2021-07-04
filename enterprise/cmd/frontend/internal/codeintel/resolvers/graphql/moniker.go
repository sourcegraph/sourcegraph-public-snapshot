package graphql

import (
	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/semantic"
)

type MonikerResolver struct {
	moniker semantic.MonikerData
}

func NewMonikerResolver(moniker semantic.MonikerData) gql.MonikerResolver {
	return &MonikerResolver{
		moniker: moniker,
	}
}

func (r *MonikerResolver) Kind() string { return r.moniker.Kind }

func (r *MonikerResolver) Scheme() string { return r.moniker.Scheme }

func (r *MonikerResolver) Identifier() string { return r.moniker.Identifier }
