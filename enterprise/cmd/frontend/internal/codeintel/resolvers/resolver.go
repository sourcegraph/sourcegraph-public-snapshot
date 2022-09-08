package resolvers

import (
	executor "github.com/sourcegraph/sourcegraph/internal/services/executors/transport/graphql"
)

// Resolver is the main interface to code intel-related operations exposed to the GraphQL API.
// This resolver consolidates the logic for code intel operations and is not itself concerned
// with GraphQL/API specifics (auth, validation, marshaling, etc.). This resolver is wrapped
// by a symmetrics resolver in this package's graphql subpackage, which is exposed directly
// by the API.
type Resolver interface {
	ExecutorResolver() executor.Resolver
	CodeNavResolver() CodeNavResolver
	PoliciesResolver() PoliciesResolver
	AutoIndexingResolver() AutoIndexingResolver
	UploadsResolver() UploadsResolver
}

type resolver struct {
	executorResolver     executor.Resolver
	codenavResolver      CodeNavResolver
	policiesResolver     PoliciesResolver
	autoIndexingResolver AutoIndexingResolver
	uploadsResolver      UploadsResolver
}

// NewResolver creates a new resolver with the given services.
func NewResolver(
	codenavResolver CodeNavResolver,
	executorResolver executor.Resolver,
	policiesResolver PoliciesResolver,
	autoIndexingResolver AutoIndexingResolver,
	uploadsResolver UploadsResolver,
) Resolver {
	return &resolver{
		executorResolver:     executorResolver,
		codenavResolver:      codenavResolver,
		policiesResolver:     policiesResolver,
		autoIndexingResolver: autoIndexingResolver,
		uploadsResolver:      uploadsResolver,
	}
}

func (r *resolver) CodeNavResolver() CodeNavResolver {
	return r.codenavResolver
}

func (r *resolver) PoliciesResolver() PoliciesResolver {
	return r.policiesResolver
}

func (r *resolver) AutoIndexingResolver() AutoIndexingResolver {
	return r.autoIndexingResolver
}

func (r *resolver) UploadsResolver() UploadsResolver {
	return r.uploadsResolver
}

func (r *resolver) ExecutorResolver() executor.Resolver {
	return r.executorResolver
}
