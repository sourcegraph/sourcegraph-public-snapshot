package graphqlbackend

import (
	"context"

	"github.com/cockroachdb/errors"
	"github.com/inconshreveable/log15"
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
	repository *RepositoryResolver
	commit     string
	path       string
	t          *compute.Text
}

func (c *computeTextResolver) Repository() *RepositoryResolver { return c.repository }

func (c *computeTextResolver) Commit() *string {
	value := c.commit
	return &value
}

func (c *computeTextResolver) Path() *string {
	value := c.path
	return &value
}

func (c *computeTextResolver) Kind() *string {
	value := c.t.Kind
	return &value
}
func (c *computeTextResolver) Value() string { return c.t.Value }

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

func toComputeMatchContextResolver(fm *result.FileMatch, mc *compute.MatchContext, repository *RepositoryResolver) *computeMatchContextResolver {
	var computeMatches []*computeMatchResolver
	for _, m := range mc.Matches {
		mCopy := m
		computeMatches = append(computeMatches, &computeMatchResolver{m: &mCopy})
	}
	return &computeMatchContextResolver{
		repository: repository,
		commit:     string(fm.CommitID),
		path:       fm.Path,
		matches:    computeMatches,
	}
}

func toComputeTextResolver(fm *result.FileMatch, result *compute.Text, repository *RepositoryResolver) *computeTextResolver {
	return &computeTextResolver{
		repository: repository,
		commit:     string(fm.CommitID),
		path:       fm.Path,
		t:          result,
	}
}

func toComputeResultResolver(r interface{}) *computeResultResolver {
	return &computeResultResolver{result: r}
}

func toResultResolverList(ctx context.Context, cmd compute.Command, matches []result.Match, db dbutil.DB) ([]*computeResultResolver, error) {
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

	computeResult := make([]*computeResultResolver, 0, len(matches))
	for _, m := range matches {
		if fm, ok := m.(*result.FileMatch); ok {
			repoResolver := getRepoResolver(fm.Repo, "")
			switch c := cmd.(type) {
			case *compute.MatchOnly:
				matchContext := compute.FromFileMatch(fm, c.MatchPattern.(*compute.Regexp).Value)
				computeResult = append(computeResult, toComputeResultResolver(toComputeMatchContextResolver(fm, matchContext, repoResolver)))
			case *compute.Replace:
				result, err := compute.ReplaceInPlaceFromFileMatch(ctx, fm, c)
				if err != nil {
					return nil, err
				}
				computeResult = append(computeResult, toComputeResultResolver(toComputeTextResolver(fm, result, repoResolver)))
			default:
				return nil, errors.Errorf("unsupported compute command %T", c)
			}

		}
	}
	return computeResult, nil
}

// NewComputeImplementer is a function that abstracts away the need to have a
// handle on (*schemaResolver) Compute.
func NewComputeImplementer(ctx context.Context, db dbutil.DB, args *ComputeArgs) ([]*computeResultResolver, error) {
	computeQuery, err := compute.Parse(args.Query)
	if err != nil {
		return nil, err
	}

	searchQuery, err := computeQuery.ToSearchQuery()
	if err != nil {
		return nil, err
	}
	log15.Info("compute", "search", searchQuery)

	patternType := "regexp"
	job, err := NewSearchImplementer(ctx, db, &SearchArgs{Query: searchQuery, PatternType: &patternType})
	if err != nil {
		return nil, err
	}

	results, err := job.Results(ctx)
	if err != nil {
		return nil, err
	}
	return toResultResolverList(ctx, computeQuery.Command, results.Matches, db)
}

func (r *schemaResolver) Compute(ctx context.Context, args *ComputeArgs) ([]*computeResultResolver, error) {
	return NewComputeImplementer(ctx, r.db, args)
}
