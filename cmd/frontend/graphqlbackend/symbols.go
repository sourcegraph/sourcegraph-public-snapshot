package graphqlbackend

import (
	"context"
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/gqlutil"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
	"github.com/sourcegraph/sourcegraph/internal/search/symbol"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

type symbolsArgs struct {
	gqlutil.ConnectionArgs
	Query           *string
	IncludePatterns *[]string
}

func (r *GitTreeEntryResolver) Symbols(ctx context.Context, args *symbolsArgs) (*symbolConnectionResolver, error) {
	symbols, err := symbol.DefaultZoektSymbolsClient().Compute(
		ctx,
		types.MinimalRepo{ID: r.commit.repoResolver.id, Name: r.commit.repoResolver.name},
		api.CommitID(r.commit.oid),
		r.commit.inputRev,
		args.Query,
		args.First,
		args.IncludePatterns,
	)
	if err != nil && len(symbols) == 0 {
		return nil, err
	}
	return &symbolConnectionResolver{
		symbols: symbolResultsToResolvers(r.db, r.commit, symbols),
		first:   args.First,
	}, nil
}

func (r *GitTreeEntryResolver) Symbol(ctx context.Context, args *struct {
	Line      int32
	Character int32
}) (*symbolResolver, error) {
	symbolMatch, err := symbol.DefaultZoektSymbolsClient().GetMatchAtLineCharacter(
		ctx,
		types.MinimalRepo{ID: r.commit.repoResolver.id, Name: r.commit.repoResolver.name},
		api.CommitID(r.commit.oid),
		r.Path(),
		int(args.Line),
		int(args.Character),
	)
	if err != nil || symbolMatch == nil {
		return nil, err
	}
	return &symbolResolver{r.db, r.commit, symbolMatch}, nil
}

func (r *GitCommitResolver) Symbols(ctx context.Context, args *symbolsArgs) (*symbolConnectionResolver, error) {
	symbols, err := symbol.DefaultZoektSymbolsClient().Compute(
		ctx,
		types.MinimalRepo{ID: r.repoResolver.id, Name: r.repoResolver.name},
		api.CommitID(r.oid),
		r.inputRev,
		args.Query,
		args.First,
		args.IncludePatterns,
	)
	if err != nil && len(symbols) == 0 {
		return nil, err
	}
	return &symbolConnectionResolver{
		symbols: symbolResultsToResolvers(r.db, r, symbols),
		first:   args.First,
	}, nil
}

func symbolResultsToResolvers(db database.DB, commit *GitCommitResolver, symbolMatches []*result.SymbolMatch) []symbolResolver {
	symbolResolvers := make([]symbolResolver, 0, len(symbolMatches))
	for _, symbolMatch := range symbolMatches {
		symbolResolvers = append(symbolResolvers, toSymbolResolver(db, commit, symbolMatch))
	}
	return symbolResolvers
}

func toSymbolResolver(db database.DB, commit *GitCommitResolver, sr *result.SymbolMatch) symbolResolver {
	return symbolResolver{
		db:          db,
		commit:      commit,
		SymbolMatch: sr,
	}
}

type symbolConnectionResolver struct {
	first   *int32
	symbols []symbolResolver
}

func limitOrDefault(first *int32) int {
	if first == nil {
		return symbol.DefaultSymbolLimit
	}
	return int(*first)
}

func (r *symbolConnectionResolver) Nodes(ctx context.Context) ([]symbolResolver, error) {
	symbols := r.symbols
	if len(r.symbols) > limitOrDefault(r.first) {
		symbols = symbols[:limitOrDefault(r.first)]
	}
	return symbols, nil
}

func (r *symbolConnectionResolver) PageInfo(ctx context.Context) (*gqlutil.PageInfo, error) {
	return gqlutil.HasNextPage(len(r.symbols) > limitOrDefault(r.first)), nil
}

type symbolResolver struct {
	db     database.DB
	commit *GitCommitResolver
	*result.SymbolMatch
}

func (r symbolResolver) Name() string { return r.Symbol.Name }

func (r symbolResolver) ContainerName() *string {
	if r.Symbol.Parent == "" {
		return nil
	}
	return &r.Symbol.Parent
}

func (r symbolResolver) Kind() string /* enum SymbolKind */ {
	kind := r.Symbol.LSPKind()
	if kind == 0 {
		return "UNKNOWN"
	}
	return strings.ToUpper(kind.String())
}

func (r symbolResolver) Language() string { return r.Symbol.Language }

func (r symbolResolver) Location() *locationResolver {
	stat := CreateFileInfo(r.Symbol.Path, false)
	sr := r.Symbol.Range()
	opts := GitTreeEntryResolverOpts{
		Commit: r.commit,
		Stat:   stat,
	}
	return &locationResolver{
		resource: NewGitTreeEntryResolver(r.db, gitserver.NewClient("graphql.symbols"), opts),
		lspRange: &sr,
	}
}

func (r symbolResolver) URL(ctx context.Context) (string, error) { return r.Location().URL(ctx) }

func (r symbolResolver) CanonicalURL() string { return r.Location().CanonicalURL() }

func (r symbolResolver) FileLocal() bool { return r.Symbol.FileLimited }
