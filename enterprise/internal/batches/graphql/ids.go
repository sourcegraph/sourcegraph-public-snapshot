package graphql

import (
	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/graphql/kind"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func MarshalBatchChangeID(id int64) graphql.ID {
	return relay.MarshalID(kind.BatchChange, id)
}

func MarshalChangesetID(id int64) graphql.ID {
	return relay.MarshalID(kind.Changeset, id)
}

func MarshalNamespaceID(userID, orgID int32) (graphql.ID, error) {
	// This is essentially a reimplementation of code in
	// cmd/frontend/graphqlbackend to keep our import tree at least a little
	// clean.
	if userID != 0 {
		return MarshalUserID(userID), nil
	} else if orgID != 0 {
		return relay.MarshalID(kind.Org, orgID), nil
	}
	return "", errors.New("cannot marshal namespace ID: neither user nor org ID provided")
}

func MarshalRepoID(id api.RepoID) graphql.ID {
	return relay.MarshalID(kind.Repo, int32(id))
}

func MarshalUserID(id int32) graphql.ID {
	return relay.MarshalID(kind.User, id)
}
