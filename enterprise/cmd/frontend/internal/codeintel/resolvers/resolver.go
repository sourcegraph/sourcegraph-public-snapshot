package resolvers

import (
	autoindexinggraphql "github.com/sourcegraph/sourcegraph/internal/codeintel/autoindexing/transport/graphql"
	codenavgraphql "github.com/sourcegraph/sourcegraph/internal/codeintel/codenav/transport/graphql"
	policiesgraphql "github.com/sourcegraph/sourcegraph/internal/codeintel/policies/transport/graphql"
	uploadsgraphql "github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/transport/graphql"
	executor "github.com/sourcegraph/sourcegraph/internal/services/executors/transport/graphql"
)

// Resolver is the main interface to code intel-related operations exposed to the GraphQL API.
// This resolver consolidates the logic for code intel operations and is not itself concerned
// with GraphQL/API specifics (auth, validation, marshaling, etc.). This resolver is wrapped
// by a symmetrics resolver in this package's graphql subpackage, which is exposed directly
// by the API.
type Resolver interface {
	ExecutorResolver() executor.Resolver
	CodeNavResolver() codenavgraphql.RootResolver
	PoliciesRootResolver() policiesgraphql.RootResolver
	AutoIndexingRootResolver() autoindexinggraphql.RootResolver
	UploadRootResolver() uploadsgraphql.RootResolver
}

type resolver struct {
	executorResolver         executor.Resolver
	codenavResolver          codenavgraphql.RootResolver
	policiesRootResolver     policiesgraphql.RootResolver
	autoIndexingRootResolver autoindexinggraphql.RootResolver
	uploadsRootResolver      uploadsgraphql.RootResolver
}

// NewResolver creates a new resolver with the given services.
func NewResolver(
	codenavResolver codenavgraphql.RootResolver,
	executorResolver executor.Resolver,
	policiesRootResolver policiesgraphql.RootResolver,
	autoIndexingRootResolver autoindexinggraphql.RootResolver,
	uploadsRootResolver uploadsgraphql.RootResolver,
) Resolver {
	return &resolver{
		executorResolver:         executorResolver,
		codenavResolver:          codenavResolver,
		policiesRootResolver:     policiesRootResolver,
		autoIndexingRootResolver: autoIndexingRootResolver,
		uploadsRootResolver:      uploadsRootResolver,
	}
}

func (r *resolver) CodeNavResolver() codenavgraphql.RootResolver {
	return r.codenavResolver
}

func (r *resolver) PoliciesRootResolver() policiesgraphql.RootResolver {
	return r.policiesRootResolver
}

func (r *resolver) AutoIndexingRootResolver() autoindexinggraphql.RootResolver {
	return r.autoIndexingRootResolver
}

func (r *resolver) UploadRootResolver() uploadsgraphql.RootResolver {
	return r.uploadsRootResolver
}

func (r *resolver) ExecutorResolver() executor.Resolver {
	return r.executorResolver
}
