package graphqlbackend

import (
	"context"
	"errors"

	graphql "github.com/graph-gophers/graphql-go"
)

// Git is the implementation of the GraphQL type GitMutation. If it is not set at runtime, a "not
// implemented" error is returned to API clients who invoke it.
var Git GitResolver

func (schemaResolver) Git() (GitResolver, error) {
	if Git == nil {
		return nil, errors.New("git is not implemented")
	}
	return Git, nil
}

// GitResolver is the interface for the GraphQL type GitMutation.
type GitResolver interface {
	CreateRefFromPatch(context.Context, *struct{ Input GitCreateRefFromPatchInput }) (GitCreateRefFromPatchPayload, error)
}

type GitCreateRefFromPatchInput struct {
	Repository graphql.ID
	Name       string
	BaseCommit gitObjectID
	Patch      string
}

type GitCreateRefFromPatchPayload interface {
	Ref(ctx context.Context) (*GitRefResolver, error)
}
