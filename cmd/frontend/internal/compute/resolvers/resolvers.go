package resolvers

import (
	"context"
	"fmt"

	"github.com/inconshreveable/log15" //nolint:logging // TODO move all logging to sourcegraph/log
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/go-langserver/pkg/lsp"

	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/compute"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func NewResolver(logger log.Logger, db database.DB) gql.ComputeResolver {
	return &Resolver{logger: logger, db: db}
}

type Resolver struct {
	logger log.Logger
	db     database.DB
}

type computeMatchContextResolver struct {
	repository *gql.RepositoryResolver
	commit     string
	path       string
	matches    []gql.ComputeMatchResolver
}

func (c *computeMatchContextResolver) Repository() *gql.RepositoryResolver { return c.repository }
func (c *computeMatchContextResolver) Commit() string                      { return c.commit }
func (c *computeMatchContextResolver) Path() string                        { return c.path }
func (c *computeMatchContextResolver) Matches() []gql.ComputeMatchResolver { return c.matches }

type computeMatchResolver struct {
	m *compute.Match
}

type computeEnvironmentEntryResolver struct {
	variable string
	value    string
	range_   compute.Range
}

type computeTextResolver struct {
	repository *gql.RepositoryResolver
	commit     string
	path       string
	t          *compute.Text
}

func (r *computeMatchResolver) Value() string {
	return r.m.Value
}

func (r *computeMatchResolver) Range() gql.RangeResolver {
	return gql.NewRangeResolver(toLspRange(r.m.Range))
}

func (r *computeMatchResolver) Environment() []gql.ComputeEnvironmentEntryResolver {
	var resolvers []gql.ComputeEnvironmentEntryResolver
	for variable, value := range r.m.Environment {
		resolvers = append(resolvers, newEnvironmentEntryResolver(variable, value))
	}
	return resolvers
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

func (r *computeEnvironmentEntryResolver) Range() gql.RangeResolver {
	return gql.NewRangeResolver(toLspRange(r.range_))
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

func (c *computeTextResolver) Repository() *gql.RepositoryResolver { return c.repository }

func (c *computeTextResolver) Commit() *string {
	return &c.commit
}

func (c *computeTextResolver) Path() *string {
	return &c.path
}

func (c *computeTextResolver) Kind() *string {
	return &c.t.Kind
}
func (c *computeTextResolver) Value() string { return c.t.Value }

// A dummy type to express the union of compute results. This how its done by the GQL library we use.
// https://github.com/graph-gophers/graphql-go/blob/af5bb93e114f0cd4cc095dd8eae0b67070ae8f20/example/starwars/starwars.go#L485-L487
//
// union ComputeResult = ComputeMatchContext | ComputeText

type computeResultResolver struct {
	result any
}

func (r *computeResultResolver) ToComputeMatchContext() (gql.ComputeMatchContextResolver, bool) {
	res, ok := r.result.(*computeMatchContextResolver)
	return res, ok
}

func (r *computeResultResolver) ToComputeText() (gql.ComputeTextResolver, bool) {
	res, ok := r.result.(*computeTextResolver)
	return res, ok
}

func toComputeMatchContextResolver(mc *compute.MatchContext, repository *gql.RepositoryResolver, path, commit string) *computeMatchContextResolver {
	computeMatches := make([]gql.ComputeMatchResolver, 0, len(mc.Matches))
	for _, m := range mc.Matches {
		mCopy := m
		computeMatches = append(computeMatches, &computeMatchResolver{m: &mCopy})
	}
	return &computeMatchContextResolver{
		repository: repository,
		commit:     commit,
		path:       path,
		matches:    computeMatches,
	}
}

func toComputeTextResolver(result *compute.Text, repository *gql.RepositoryResolver, path, commit string) *computeTextResolver {
	return &computeTextResolver{
		repository: repository,
		commit:     commit,
		path:       path,
		t:          result,
	}
}

func toComputeResultResolver(result compute.Result, repoResolver *gql.RepositoryResolver, path, commit string) gql.ComputeResultResolver {
	switch r := result.(type) {
	case *compute.MatchContext:
		return &computeResultResolver{result: toComputeMatchContextResolver(r, repoResolver, path, commit)}
	case *compute.Text:
		return &computeResultResolver{result: toComputeTextResolver(r, repoResolver, path, commit)}
	default:
		panic(fmt.Sprintf("unsupported compute result %T", r))
	}
}

func pathAndCommitFromResult(m result.Match) (string, string) {
	switch v := m.(type) {
	case *result.FileMatch:
		return v.Path, string(v.CommitID)
	case *result.CommitMatch:
		return "", string(v.Commit.ID)
	case *result.RepoMatch:
		return "", v.Rev
	}
	return "", ""
}

func toResultResolverList(ctx context.Context, cmd compute.Command, matches []result.Match, db database.DB) ([]gql.ComputeResultResolver, error) {
	gitserverClient := gitserver.NewClient("graphql.compute")

	repoResolvers := make(map[types.MinimalRepo]*gql.RepositoryResolver, 10)
	getRepoResolver := func(repoName types.MinimalRepo) *gql.RepositoryResolver {
		if existing, ok := repoResolvers[repoName]; ok {
			return existing
		}
		resolver := gql.NewMinimalRepositoryResolver(db, gitserverClient, repoName.ID, repoName.Name)
		repoResolvers[repoName] = resolver
		return resolver
	}

	results := make([]gql.ComputeResultResolver, 0, len(matches))
	for _, m := range matches {
		computeResult, err := cmd.Run(ctx, gitserverClient, m)
		if err != nil {
			return nil, err
		}

		if computeResult == nil {
			// We processed a match that compute doesn't generate a result for.
			continue
		}

		repoResolver := getRepoResolver(m.RepoName())
		path, commit := pathAndCommitFromResult(m)
		resolver := toComputeResultResolver(computeResult, repoResolver, path, commit)
		results = append(results, resolver)
	}
	return results, nil
}

// NewBatchComputeImplementer is a function that abstracts away the need to have a
// handle on (*schemaResolver) Compute.
func NewBatchComputeImplementer(ctx context.Context, logger log.Logger, db database.DB, args *gql.ComputeArgs) ([]gql.ComputeResultResolver, error) {
	computeQuery, err := compute.Parse(args.Query)
	if err != nil {
		return nil, err
	}

	searchQuery, err := computeQuery.ToSearchQuery()
	if err != nil {
		return nil, err
	}
	log15.Debug("compute", "search", searchQuery)

	patternType := "regexp"
	job, err := gql.NewBatchSearchImplementer(ctx, logger, db, &gql.SearchArgs{Query: searchQuery, PatternType: &patternType, Version: "V3"})
	if err != nil {
		return nil, err
	}

	results, err := job.Results(ctx)
	if err != nil {
		return nil, err
	}
	return toResultResolverList(ctx, computeQuery.Command, results.Matches, db)
}

func (r *Resolver) Compute(ctx context.Context, args *gql.ComputeArgs) ([]gql.ComputeResultResolver, error) {
	return NewBatchComputeImplementer(ctx, r.logger, r.db, args)
}
