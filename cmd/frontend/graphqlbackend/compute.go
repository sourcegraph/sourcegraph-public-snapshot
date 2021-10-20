package graphqlbackend

import (
	"context"
	"regexp"

	"github.com/sourcegraph/go-langserver/pkg/lsp"
	"github.com/sourcegraph/sourcegraph/internal/compute"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
	"github.com/sourcegraph/sourcegraph/internal/types"
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

func toComputeMatchContextResolver(fm *result.FileMatch, mc *compute.MatchContext, db dbutil.DB) *computeMatchContextResolver {
	type repoKey struct {
		Name types.RepoName
		Rev  string
	}
	repoResolvers := make(map[repoKey]*RepositoryResolver, 10)
	getRepoResolver := func(repoName types.RepoName, rev string) *RepositoryResolver {
		if existing, ok := repoResolvers[repoKey{repoName, rev}]; ok {
			return existing
		}
		resolver := NewRepositoryResolver(db, repoName.ToRepo())
		resolver.RepoMatch.Rev = rev
		repoResolvers[repoKey{repoName, rev}] = resolver
		return resolver
	}

	var computeMatches []*computeMatchResolver
	for _, m := range mc.Matches {
		mCopy := m
		computeMatches = append(computeMatches, &computeMatchResolver{m: &mCopy})
	}
	return &computeMatchContextResolver{
		repository: getRepoResolver(fm.Repo, ""),
		commit:     string(fm.CommitID),
		path:       fm.Path,
		matches:    computeMatches,
	}
}

func toComputeResultResolver(r *computeMatchContextResolver) *computeResultResolver {
	return &computeResultResolver{result: r}
}

func toResultResolverList(pattern *regexp.Regexp, matches []result.Match, db dbutil.DB) []*computeResultResolver {
	var computeResult []*computeResultResolver
	for _, m := range matches {
		if fm, ok := m.(*result.FileMatch); ok {
			matchContext := compute.FromFileMatch(fm, pattern)
			computeResult = append(computeResult, toComputeResultResolver(toComputeMatchContextResolver(fm, matchContext, db)))
		}
	}
	return computeResult
}

// NewComputeImplementer is a function that abstracts away the need to have a
// handle on (*schemaResolver) Compute.
func NewComputeImplementer(ctx context.Context, db dbutil.DB, args *ComputeArgs) ([]*computeResultResolver, error) {
	query, err := compute.Parse(args.Query)
	if err != nil {
		return nil, err
	}
	patternType := "regexp"
	job, err := NewSearchImplementer(ctx, db, &SearchArgs{Query: args.Query, PatternType: &patternType})
	if err != nil {
		return nil, err
	}

	results, err := job.Results(ctx)
	if err != nil {
		return nil, err
	}
	pattern := query.(*compute.MatchOnly).MatchPattern.(*compute.Regexp).Value
	return toResultResolverList(pattern, results.Matches, db), nil
}

func (r *schemaResolver) Compute(ctx context.Context, args *ComputeArgs) ([]*computeResultResolver, error) {
	return NewComputeImplementer(ctx, r.db, args)
}
