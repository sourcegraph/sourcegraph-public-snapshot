package graphqlbackend

import (
	"context"
	"errors"
	"regexp/syntax"
	"strings"
	"time"

	"github.com/google/zoekt"
	zoektquery "github.com/google/zoekt/query"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/gituri"
	"github.com/sourcegraph/sourcegraph/internal/search/search"
	"github.com/sourcegraph/sourcegraph/internal/symbols/protocol"
)

type symbolsArgs struct {
	graphqlutil.ConnectionArgs
	Query           *string
	IncludePatterns *[]string
}

func (r *GitTreeEntryResolver) Symbols(ctx context.Context, args *symbolsArgs) (*symbolConnectionResolver, error) {
	symbols, err := computeSymbols(ctx, r.commit, args.Query, args.First, args.IncludePatterns)
	if err != nil && len(symbols) == 0 {
		return nil, err
	}
	return &symbolConnectionResolver{symbols: symbols, first: args.First}, nil
}

func (r *GitCommitResolver) Symbols(ctx context.Context, args *symbolsArgs) (*symbolConnectionResolver, error) {
	symbols, err := computeSymbols(ctx, r, args.Query, args.First, args.IncludePatterns)
	if err != nil && len(symbols) == 0 {
		return nil, err
	}
	return &symbolConnectionResolver{symbols: symbols, first: args.First}, nil
}

type symbolConnectionResolver struct {
	first   *int32
	symbols []*symbolResolver
}

func limitOrDefault(first *int32) int {
	if first == nil {
		return 100
	}
	return int(*first)
}

// indexedSymbols checks to see if Zoekt has indexed
// symbols information for a repository at a specific
// commit.
func indexedSymbols(commit *GitCommitResolver) bool {
	z := search.Indexed()
	if !z.Enabled() {
		return false
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	set, err := z.ListAll(ctx)
	if err != nil {
		return false
	}

	repo, ok := set[string(commit.repo.repo.Name)]
	if !ok || !repo.HasSymbols {
		return false
	}

	for _, branch := range repo.Branches {
		if branch.Version == string(commit.oid) {
			return true
		}
	}

	return false
}

func searchZoektSymbols(ctx context.Context, commit *GitCommitResolver, queryString *string, first *int32, includePatterns *[]string) (res []*symbolResolver, err error) {
	raw := *queryString
	if raw == "" {
		raw = ".*"
	}

	expr, err := syntax.Parse(raw, syntax.ClassNL|syntax.PerlX|syntax.UnicodeGroups)
	if err != nil {
		return
	}

	var query zoektquery.Q
	if expr.Op == syntax.OpLiteral {
		query = &zoektquery.Substring{
			Pattern: string(expr.Rune),
			Content: true,
		}
	} else {
		query = &zoektquery.Regexp{
			Regexp:  expr,
			Content: true,
		}
	}

	sym := &zoektquery.Symbol{Expr: query}
	repo := &zoektquery.RepoSet{Set: map[string]bool{
		string(commit.repo.repo.Name): true,
	}}
	ands := []zoektquery.Q{repo, sym}
	for _, p := range *includePatterns {
		q, err := fileRe(p, true)
		if err != nil {
			return nil, err
		}
		ands = append(ands, q)
	}

	final := zoektquery.Simplify(zoektquery.NewAnd(ands...))
	match := limitOrDefault(first) + 1
	resp, err := search.Indexed().Client.Search(ctx, final, &zoekt.SearchOptions{
		MaxWallTime:            3 * time.Second,
		ShardMaxMatchCount:     match * 25,
		TotalMaxMatchCount:     match * 25,
		ShardMaxImportantMatch: match * 25,
		TotalMaxImportantMatch: match * 25,
		MaxDocDisplayCount:     match,
	})
	if err != nil {
		return nil, err
	}

	baseURI, err := gituri.Parse("git://" + string(commit.repo.repo.Name) + "?" + string(commit.oid))
	for _, file := range resp.Files {
		for _, l := range file.LineMatches {
			if !l.FileName {
				for _, m := range l.LineFragments {
					if m.SymbolInfo != nil {
						res = append(res, toSymbolResolver(
							protocol.Symbol{
								Name:       m.SymbolInfo.Sym,
								Kind:       m.SymbolInfo.Kind,
								Parent:     m.SymbolInfo.Parent,
								ParentKind: m.SymbolInfo.ParentKind,
								Path:       file.FileName,
								Line:       l.LineNumber,
							},
							baseURI,
							strings.ToLower(file.Language),
							commit,
						))
					}
				}
			}
		}
	}
	return
}

func computeSymbols(ctx context.Context, commit *GitCommitResolver, query *string, first *int32, includePatterns *[]string) (res []*symbolResolver, err error) {
	if indexedSymbols(commit) {
		return searchZoektSymbols(ctx, commit, query, first, includePatterns)
	}

	ctx, done := context.WithTimeout(ctx, 5*time.Second)
	defer done()
	defer func() {
		if ctx.Err() != nil && len(res) == 0 {
			err = errors.New("processing symbols is taking longer than expected. Try again in a while")
		}
	}()
	var includePatternsSlice []string
	if includePatterns != nil {
		includePatternsSlice = *includePatterns
	}

	searchArgs := search.SymbolsParameters{
		CommitID:        api.CommitID(commit.oid),
		First:           limitOrDefault(first) + 1, // add 1 so we can determine PageInfo.hasNextPage
		Repo:            commit.repo.repo.Name,
		IncludePatterns: includePatternsSlice,
	}
	if query != nil {
		searchArgs.Query = *query
	}
	baseURI, err := gituri.Parse("git://" + string(commit.repo.repo.Name) + "?" + string(commit.oid))
	if err != nil {
		return nil, err
	}
	symbols, err := backend.Symbols.ListTags(ctx, searchArgs)
	if baseURI == nil {
		return
	}
	resolvers := make([]*symbolResolver, 0, len(symbols))
	for _, symbol := range symbols {
		resolver := toSymbolResolver(symbol, baseURI, strings.ToLower(symbol.Language), commit)
		if resolver == nil {
			continue
		}
		resolvers = append(resolvers, resolver)
	}
	return resolvers, err
}

func toSymbolResolver(symbol protocol.Symbol, baseURI *gituri.URI, lang string, commitResolver *GitCommitResolver) *symbolResolver {
	resolver := &symbolResolver{
		symbol:   symbol,
		language: lang,
		uri:      baseURI.WithFilePath(symbol.Path),
	}
	symbolRange := symbolRange(symbol)
	resolver.location = &locationResolver{
		resource: &GitTreeEntryResolver{
			commit: commitResolver,
			stat:   CreateFileInfo(resolver.uri.Fragment, false), // assume the path refers to a file (not dir)
		},
		lspRange: &symbolRange,
	}
	return resolver
}

func (r *symbolConnectionResolver) Nodes(ctx context.Context) ([]*symbolResolver, error) {
	symbols := r.symbols
	if len(r.symbols) > limitOrDefault(r.first) {
		symbols = symbols[:limitOrDefault(r.first)]
	}
	return symbols, nil
}

func (r *symbolConnectionResolver) PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error) {
	return graphqlutil.HasNextPage(len(r.symbols) > limitOrDefault(r.first)), nil
}

type symbolResolver struct {
	symbol   protocol.Symbol
	language string
	location *locationResolver
	uri      *gituri.URI
}

func (r *symbolResolver) Name() string { return r.symbol.Name }

func (r *symbolResolver) ContainerName() *string {
	if r.symbol.Parent == "" {
		return nil
	}
	return &r.symbol.Parent
}

func (r *symbolResolver) Kind() string /* enum SymbolKind */ {
	kind := ctagsKindToLSPSymbolKind(r.symbol.Kind)
	if kind == 0 {
		return "UNKNOWN"
	}
	return strings.ToUpper(kind.String())
}

func (r *symbolResolver) Language() string { return r.language }

func (r *symbolResolver) Location() *locationResolver { return r.location }

func (r *symbolResolver) URL(ctx context.Context) (string, error) { return r.location.URL(ctx) }

func (r *symbolResolver) CanonicalURL() (string, error) { return r.location.CanonicalURL() }

func (r *symbolResolver) FileLocal() bool { return r.symbol.FileLimited }
