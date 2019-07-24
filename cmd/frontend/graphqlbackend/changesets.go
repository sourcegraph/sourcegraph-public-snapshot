package graphqlbackend

import (
	"context"
	"errors"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
)

// Changesets is the implementation of the GraphQL type ChangesetsMutation. If it is not set at
// runtime, a "not implemented" error is returned to API clients who invoke it.
var Changesets ChangesetsResolver

// ChangesetsForRepository returns an instance of the GraphQL ChangesetConnection type with the list
// of changesets defined in a repository.
func ChangesetsForRepository(ctx context.Context, repository *RepositoryResolver, arg *graphqlutil.ConnectionArgs) (ChangesetConnection, error) {
	if Changesets == nil {
		return nil, errors.New("changesets is not implemented")
	}
	return Changesets.ChangesetsForRepository(ctx, repository, arg)
}

func (schemaResolver) Changesets() (ChangesetsResolver, error) {
	if Changesets == nil {
		return nil, errors.New("changesets is not implemented")
	}
	return Changesets, nil
}

// ChangesetsResolver is the interface for the GraphQL type ChangesetsMutation.
type ChangesetsResolver interface {
	CreateChangeset(context.Context, *struct {
		Input ChangesetsCreateChangesetInput
	}) (ChangesetsCreateChangesetPayload, error)

	// ChangesetsForRepository is called by the ChangesetsForRepository func but is not in the
	// GraphQL API.
	ChangesetsForRepository(ctx context.Context, repository *RepositoryResolver, arg *graphqlutil.ConnectionArgs) (ChangesetConnection, error)
}

// Changeset is the interface for the GraphQL type Changeset.
type Changeset interface {
	Title() string
	ExternalURL() *string
}

type ChangesetsCreateChangesetInput struct {
	Repository  graphql.ID
	Title       string
	ExternalURL *string
}

type ChangesetsCreateChangesetPayload interface {
	Changeset() Changeset
}

// ChangesetConnection is the interface for the GraphQL type ChangesetConnection.
type ChangesetConnection interface {
	Nodes(context.Context) ([]Changeset, error)
	TotalCount(context.Context) (int32, error)
	PageInfo(context.Context) (*graphqlutil.PageInfo, error)
}
