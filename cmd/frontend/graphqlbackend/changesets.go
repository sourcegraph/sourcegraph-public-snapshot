package graphqlbackend

import (
	"context"
	"errors"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
)

// Changesets is the implementation of the GraphQL type ChangesetsMutation. If it is not set at
// runtime, a "not implemented" error is returned to API clients who invoke it.
var Changesets ChangesetsResolver

func (schemaResolver) Changesets() (ChangesetsResolver, error) {
	if Changesets == nil {
		return nil, errors.New("changesets is not implemented")
	}
	return Changesets, nil
}

// ChangesetFor returns an instance of the GraphQL Changeset type for a DiscussionThread.
func ChangesetFor(t *types.DiscussionThread) (Changeset, error) {
	if Changesets == nil {
		return nil, errors.New("changeset is not implemented")
	}
	return Changesets.ChangesetFor(t)
}

// ChangesetsResolver is the interface for the GraphQL type ChangesetsMutation.
type ChangesetsResolver interface {
	CreateChangeset(context.Context, *struct {
		Input ChangesetsCreateChangesetInput
	}) (ChangesetsCreateChangesetPayload, error)

	// ChangesetFor is called by the ChangesetFor func but is not in the GraphQL API.
	ChangesetFor(*types.DiscussionThread) (Changeset, error)
}

// Changeset is the interface for the GraphQL type Changeset.
type Changeset interface {
	Repositories(context.Context) ([]*RepositoryResolver, error)
	RepositoryComparisons(context.Context) ([]*RepositoryComparisonResolver, error)
}

type ChangesetsCreateChangesetInput struct {
	Title   string
	Body    string
	Project graphql.ID
}

type ChangesetsCreateChangesetPayload interface {
	Thread() *discussionThreadResolver
}
