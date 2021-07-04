package graphqlbackend

import (
	"context"

	"github.com/graph-gophers/graphql-go"
)

// This file just contains stub GraphQL resolvers and data types for Guide which return an error if
// not running in enterprise mode. The actual resolvers are in enterprise/internal/guide/resolvers.

type GuideRootResolver interface {
	GuideInfo(context.Context, *GuideInfoParams) (GuideInfoResolver, error)
}

type GuideInfoParams struct {
	Repository GuideRepositoryInput
	Selections []GuideSelectionInput
}

type GuideRepositoryInput struct {
	ID   *graphql.ID
	Name *string

	Revision                   *string
	CommitID                   *string
	Dirty                      *bool
	LastPublicAncestorCommitID *string
}

type GuideSelectionInput struct {
	Path *string

	Line      *int32
	Character *int32

	SymbolMonikers *[]MonikerInput
}

type GuideInfoResolver interface {
	Hello() string
	URL() string
	Monikers(context.Context) ([]MonikerResolver, error)
	Hover(context.Context) (HoverResolver, error)
	References(context.Context) (LocationConnectionResolver, error)
	EditCommits(context.Context) (*GitCommitConnectionResolver, error)
}
