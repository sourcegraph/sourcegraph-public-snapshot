package symbol

import (
	"context"
	"regexp/syntax" //nolint:depguard // zoekt requires this pkg
	"time"

	"github.com/RoaringBitmap/roaring"
	"github.com/google/zoekt"
	zoektquery "github.com/google/zoekt/query"
	"github.com/grafana/regexp"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
	zoektutil "github.com/sourcegraph/sourcegraph/internal/search/zoekt"
	"github.com/sourcegraph/sourcegraph/internal/trace/ot"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

const DefaultSymbolLimit = 100

// indexedSymbols checks to see if Zoekt has indexed symbols information for a
// repository at a specific commit. If it has it returns the branch name (for
// use when querying zoekt). Otherwise an empty string is returned.
func indexedSymbolsBranch(ctx context.Context, repo *types.MinimalRepo, commit string) string {
	if !conf.SearchIndexEnabled() {
		return ""
	}
	z := search.Indexed()

	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()
	list, err := z.List(ctx, &zoektquery.Const{Value: true}, &zoekt.ListOptions{Minimal: true})
	if err != nil {
		return ""
	}

	r, ok := list.Minimal[uint32(repo.ID)]
	if !ok || !r.HasSymbols {
		return ""
	}

	for _, branch := range r.Branches {
		if branch.Version == commit {
			return branch.Name
		}
	}

	return ""
}

func searchZoekt(ctx context.Context, repoName types.MinimalRepo, commitID api.CommitID, inputRev *string, branch string, queryString *string, first *int32, includePatterns *[]string) (res []*result.SymbolMatch, err error) {
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

	ands := []zoektquery.Q{
		&zoektquery.BranchesRepos{List: []zoektquery.BranchRepos{
			{Branch: branch, Repos: roaring.BitmapOf(uint32(repoName.ID))},
		}},
		&zoektquery.Symbol{Expr: query},
	}
	if includePatterns != nil {
		for _, p := range *includePatterns {
			q, err := zoektutil.FileRe(p, true)
			if err != nil {
				return nil, err
			}
			ands = append(ands, q)
		}
	}

	final := zoektquery.Simplify(zoektquery.NewAnd(ands...))
	match := limitOrDefault(first) + 1
	resp, err := search.Indexed().Search(ctx, final, &zoekt.SearchOptions{
		Trace:                  ot.ShouldTrace(ctx),
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

	for _, file := range resp.Files {
		newFile := &result.File{
			Repo:     repoName,
			CommitID: commitID,
			InputRev: inputRev,
			Path:     file.FileName,
		}

		for _, l := range file.LineMatches {
			if l.FileName {
				continue
			}

			for _, m := range l.LineFragments {
				if m.SymbolInfo == nil {
					continue
				}

				res = append(res, result.NewSymbolMatch(
					newFile,
					l.LineNumber,
					-1, // -1 means infer the column
					m.SymbolInfo.Sym,
					m.SymbolInfo.Kind,
					m.SymbolInfo.Parent,
					m.SymbolInfo.ParentKind,
					file.Language,
					string(l.Line),
					false,
				))
			}
		}
	}
	return
}

func Compute(ctx context.Context, repoName types.MinimalRepo, commitID api.CommitID, inputRev *string, query *string, first *int32, includePatterns *[]string) (res []*result.SymbolMatch, err error) {
	// TODO(keegancsmith) we should be able to use indexedSearchRequest here
	// and remove indexedSymbolsBranch.
	if branch := indexedSymbolsBranch(ctx, &repoName, string(commitID)); branch != "" {
		return searchZoekt(ctx, repoName, commitID, inputRev, branch, query, first, includePatterns)
	}

	serverTimeout := 5 * time.Second
	clientTimeout := 2 * serverTimeout

	ctx, done := context.WithTimeout(ctx, clientTimeout)
	defer done()
	defer func() {
		if ctx.Err() != nil && len(res) == 0 {
			err = errors.Newf("The symbols service appears unresponsive, check the logs for errors.")
		}
	}()
	var includePatternsSlice []string
	if includePatterns != nil {
		includePatternsSlice = *includePatterns
	}

	searchArgs := search.SymbolsParameters{
		CommitID:        commitID,
		First:           limitOrDefault(first) + 1, // add 1 so we can determine PageInfo.hasNextPage
		Repo:            repoName.Name,
		IncludePatterns: includePatternsSlice,
		Timeout:         int(serverTimeout.Seconds()),
	}
	if query != nil {
		searchArgs.Query = *query
	}

	symbols, err := backend.Symbols.ListTags(ctx, searchArgs)
	if err != nil {
		return nil, err
	}

	fileWithPath := func(path string) *result.File {
		return &result.File{
			Path:     path,
			Repo:     repoName,
			InputRev: inputRev,
			CommitID: commitID,
		}
	}

	matches := make([]*result.SymbolMatch, 0, len(symbols))
	for _, symbol := range symbols {
		matches = append(matches, &result.SymbolMatch{
			Symbol: symbol,
			File:   fileWithPath(symbol.Path),
		})
	}
	return matches, err
}

// GetMatchAtLineCharacter retrieves the shortest matching symbol (if exists) defined
// at a specific line number and character offset in the provided file.
func GetMatchAtLineCharacter(ctx context.Context, repo types.MinimalRepo, commitID api.CommitID, filePath string, line int, character int) (*result.SymbolMatch, error) {
	// Should be large enough to include all symbols from a single file
	first := int32(999999)
	emptyString := ""
	includePatterns := []string{regexp.QuoteMeta(filePath)}
	symbolMatches, err := Compute(ctx, repo, commitID, &emptyString, &emptyString, &first, &includePatterns)
	if err != nil {
		return nil, err
	}

	var match *result.SymbolMatch
	for _, symbolMatch := range symbolMatches {
		symbolRange := symbolMatch.Symbol.Range()
		isWithinRange := line >= symbolRange.Start.Line && character >= symbolRange.Start.Character && line <= symbolRange.End.Line && character <= symbolRange.End.Character
		if isWithinRange && (match == nil || len(symbolMatch.Symbol.Name) < len(match.Symbol.Name)) {
			match = symbolMatch
		}
	}
	return match, nil
}

func limitOrDefault(first *int32) int {
	if first == nil {
		return DefaultSymbolLimit
	}
	return int(*first)
}
