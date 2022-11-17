package graphqlbackend

import "github.com/sourcegraph/sourcegraph/internal/codeintel/resolvers"

type CodeIntelResolver interface {
	resolvers.RootResolver

	NodeResolvers() map[string]NodeByIDFunc
}
