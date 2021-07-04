package graphqlbackend

import "context"

// This file just contains stub GraphQL resolvers and data types for Guide which return an error if
// not running in enterprise mode. The actual resolvers are in enterprise/internal/guide/resolvers.

type GuideRootResolver interface {
	GuideInfo(context.Context, *GuideInfoParams) (GuideInfoResolver, error)
}

type GuideInfoParams struct {
	RepositoryName               string
	Head                         string
	CommonPublicAncestorRevision *string
	Path                         string
	Line, Character              int32
}

type GuideInfoResolver interface{ Hello() string }
