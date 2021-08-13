package graphqlbackend

import (
	"context"

	"github.com/sourcegraph/go-langserver/pkg/lsp"
	"github.com/sourcegraph/sourcegraph/internal/compute"
)

type ComputeArgs struct {
	Query string
}

type ComputeResolver interface {
	Compute(ctx context.Context, args *ComputeArgs) ([]*computeResultResolver, error)
}

// A dummy type to express the union of compute results. This how its done by the GQL library we use.
// https://github.com/graph-gophers/graphql-go/blob/af5bb93e114f0cd4cc095dd8eae0b67070ae8f20/example/starwars/starwars.go#L485-L487
//
// union ComputeResult = ComputeMatchContext | ComputeText
type computeResultResolver struct {
	result interface{}
}

// ComputeMatchContext GQL result resolver definitions.

type computeMatchContextResolver struct {
	repository *RepositoryResolver
	commit     string
	path       string
	matches    []*computeMatchResolver
}

func (c *computeMatchContextResolver) Repository() *RepositoryResolver  { return c.repository }
func (c *computeMatchContextResolver) Commit() string                   { return c.commit }
func (c *computeMatchContextResolver) Path() string                     { return c.path }
func (c *computeMatchContextResolver) Matches() []*computeMatchResolver { return c.matches }

// computeMatch resolvers.

type computeMatchResolver struct {
	m *compute.Match
}

func (r *computeMatchResolver) Value() string {
	return r.m.Value
}

func (r *computeMatchResolver) Range() RangeResolver {
	return NewRangeResolver(toLspRange(r.m.Range))
}

func (r *computeMatchResolver) Environment() []*computeEnvironmentEntryResolver {
	var result []*computeEnvironmentEntryResolver
	for variable, value := range r.m.Environment {
		result = append(result, newEnvironmentEntryResolver(variable, value))
	}
	return result
}

// computeEnvironmentEntry resolvers.

type computeEnvironmentEntryResolver struct {
	variable string
	value    string
	range_   compute.Range
}

func newEnvironmentEntryResolver(variable string, value compute.Data) *computeEnvironmentEntryResolver {
	return &computeEnvironmentEntryResolver{
		variable: variable,
		value:    value.Value,
		range_:   value.Range,
	}
}

func (r *computeEnvironmentEntryResolver) Variable() string {
	return r.variable
}

func (r *computeEnvironmentEntryResolver) Value() string {
	return r.value
}

func (r *computeEnvironmentEntryResolver) Range() RangeResolver {
	return NewRangeResolver(toLspRange(r.range_))
}

func toLspRange(r compute.Range) lsp.Range {
	return lsp.Range{
		Start: lsp.Position{
			Line:      r.Start.Line,
			Character: r.Start.Column,
		},
		End: lsp.Position{
			Line:      r.End.Line,
			Character: r.End.Column,
		},
	}
}

// ComputeText GQL result resolver definitions.

type computeTextResolver struct {
	repository *RepositoryResolver //nolint
	commit     string              //nolint
	path       string              //nolint
	t          *compute.Text
}

func (c *computeTextResolver) Repository() *RepositoryResolver { return nil }
func (r *computeTextResolver) Commit() *string                 { return nil }
func (r *computeTextResolver) Path() *string                   { return nil }
func (r *computeTextResolver) Kind() *string                   { return nil }
func (r *computeTextResolver) Value() string                   { return r.t.Value }

// Definitions required by https://github.com/graph-gophers/graphql-go to resolve
// a union type in GraphQL.

func (r *computeResultResolver) ToComputeMatchContext() (*computeMatchContextResolver, bool) {
	res, ok := r.result.(*computeMatchContextResolver)
	return res, ok
}

func (r *computeResultResolver) ToComputeText() (*computeTextResolver, bool) {
	res, ok := r.result.(*computeTextResolver)
	return res, ok
}

// NewComputeImplementer is a function that abstracts away the need to have a
// handle on (*schemaResolver) Compute.
func NewComputeImplementer(ctx context.Context, args *ComputeArgs) ([]*computeResultResolver, error) {
	return []*computeResultResolver{{&computeTextResolver{t: &compute.Text{Value: "value"}}}}, nil
}

func (r *schemaResolver) Compute(ctx context.Context, args *ComputeArgs) ([]*computeResultResolver, error) {
	return NewComputeImplementer(ctx, args)
}
