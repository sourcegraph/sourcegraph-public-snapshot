package graphqlbackend

import (
	"context"
)

type ComputeArgs struct {
	Query string
}

type ComputeResolver interface {
	Compute(ctx context.Context, args *ComputeArgs) ([]ComputeResultResolver, error)
}

type ComputeResultResolver interface {
	ToComputeMatchContext() (ComputeMatchContextResolver, bool)
	ToComputeText() (ComputeTextResolver, bool)
}

type ComputeMatchContextResolver interface {
	Repository() *RepositoryResolver
	Commit() string
	Path() string
	Matches() []ComputeMatchResolver
}

type ComputeMatchResolver interface {
	Value() string
	Range() RangeResolver
	Environment() []ComputeEnvironmentEntryResolver
}

type ComputeEnvironmentEntryResolver interface {
	Variable() string
	Value() string
	Range() RangeResolver
}

type ComputeTextResolver interface {
	Repository() *RepositoryResolver
	Commit() *string
	Path() *string
	Kind() *string
	Value() string
}
