package graphql

import (
	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/graphql/kind"
)

func MarshalBatchChangeID(id int64) graphql.ID {
	return relay.MarshalID(kind.BatchChange, id)
}
