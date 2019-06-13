package threads

import (
	"fmt"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
)

const GQLTypeThread = "Thread"

func MarshalID(id int64) graphql.ID {
	return relay.MarshalID(GQLTypeThread, id)
}

func UnmarshalID(id graphql.ID) (dbID int64, err error) {
	if typ := relay.UnmarshalKind(id); typ != GQLTypeThread {
		return 0, fmt.Errorf("thread ID has unexpected type type %q", typ)
	}
	err = relay.UnmarshalSpec(id, &dbID)
	return
}
