package graphqlbackend

import (
	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
)

func marshalExecutorID(id int64) graphql.ID {
	return relay.MarshalID("Executor", id)
}

func unmarshalExecutorID(id graphql.ID) (executorID int64, err error) {
	err = relay.UnmarshalSpec(id, &executorID)
	return
}
