package graphqlbackend

import "context"

// This file just contains stub GraphQL resolvers and data types for Guide which return an error if
// not running in enterprise mode. The actual resolvers are in enterprise/internal/guide/resolvers.

type GuideRootResolver interface {
	GuideInfo(context.Context, *GuideInfoParams) (GuideInfoResolver, error)
}

type GuideInfoParams struct {
	Repository                   GuideRepositoryInput
	Head                         string
	CommonPublicAncestorRevision *string
	Selections                   []GuideSelectionInput
}

type GuideRepositoryInput struct {
	Name string
}

type GuideSelectionInput struct {
	Path *string

	Line      *int32
	Character *int32

	SymbolMonikers *[]MonikerInput
}

type GuideInfoResolver interface{ Hello() string }
