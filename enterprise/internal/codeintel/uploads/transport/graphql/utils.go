package graphql

import (
	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"

	"github.com/sourcegraph/sourcegraph/internal/api"
)

func unmarshalRepositoryID(id graphql.ID) (repo api.RepoID, err error) {
	err = relay.UnmarshalSpec(id, &repo)
	return
}
