package executors

import (
	executor "github.com/sourcegraph/sourcegraph/internal/services/executors/transport/graphql"
)

type Resolver struct {
	executorResolver executor.Resolver
}

func newResolver(
	executorResolver executor.Resolver,
) *Resolver {
	return &Resolver{
		executorResolver: executorResolver,
	}
}

func (r *Resolver) ExecutorResolver() executor.Resolver {
	return r.executorResolver
}
