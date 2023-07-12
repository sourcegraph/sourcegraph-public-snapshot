package graphql

import (
	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

const batchChangeIDKind = "BatchChange"

func MarshalBatchChangeID(id int64) graphql.ID {
	return relay.MarshalID(batchChangeIDKind, id)
}

const changesetIDKind = "Changeset"

func MarshalChangesetID(id int64) graphql.ID {
	return relay.MarshalID(changesetIDKind, id)
}

const orgIDKind = "Org"

func MarshalNamespaceID(userID, orgID int32) (graphql.ID, error) {
	// This is essentially a reimplementation of code in
	// cmd/frontend/graphqlbackend to keep our import tree at least a little
	// clean.
	if userID != 0 {
		return MarshalUserID(userID), nil
	} else if orgID != 0 {
		return relay.MarshalID(orgIDKind, orgID), nil
	}
	return "", errors.New("cannot marshal namespace ID: neither user nor org ID provided")
}

const repoIDKind = "Repo"

func MarshalRepoID(id api.RepoID) graphql.ID {
	return relay.MarshalID(repoIDKind, int32(id))
}

const userIDKind = "User"

func MarshalUserID(id int32) graphql.ID {
	return relay.MarshalID(userIDKind, id)
}
