package graphqlbackend

import (
	"context"
	"errors"

	graphql "github.com/graph-gophers/graphql-go"
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

// ChangesetsResolver is the interface for the GraphQL type ChangesetsMutation.
type ChangesetsResolver interface {
	CreateChangeset(context.Context, *struct {
		Input ChangesetsCreateChangesetInput
	}) (ChangesetsCreateChangesetPayload, error)
}

type ChangesetsCreateChangesetInput struct {
	Title   string
	Body    string
	Project graphql.ID
}

type ChangesetsCreateChangesetPayload interface {
	Thread() *discussionThreadResolver
}
