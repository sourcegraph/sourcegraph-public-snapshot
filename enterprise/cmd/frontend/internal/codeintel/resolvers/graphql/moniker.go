package graphql

import (
	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/stores/lsifstore"
)

type MonikerResolver struct {
	moniker lsifstore.MonikerData
}

func NewMonikerResolver(moniker lsifstore.MonikerData) gql.MonikerResolver {
	return &MonikerResolver{
		moniker: moniker,
	}
}

func (r *MonikerResolver) Kind() string { return r.moniker.Kind }

func (r *MonikerResolver) Scheme() string { return r.moniker.Scheme }

func (r *MonikerResolver) Identifier() string { return r.moniker.Identifier }
