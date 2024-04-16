package graphql

import (
	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
)

const batchChangeIDKind = "BatchChange"

func MarshalBatchChangeID(id int64) graphql.ID {
	return relay.MarshalID(batchChangeIDKind, id)
}

const changesetIDKind = "Changeset"

func MarshalChangesetID(id int64) graphql.ID {
	return relay.MarshalID(changesetIDKind, id)
}
