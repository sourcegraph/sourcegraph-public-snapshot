package search

import (
	"archive/tar"
	"context"
	"path/filepath"
	"regexp/syntax" //nolint:depguard // zoekt requires this pkg
	"strings"
	"time"

	"github.com/sourcegraph/conc/pool"
	"github.com/sourcegraph/log"
	"github.com/sourcegraph/zoekt"
	"github.com/sourcegraph/zoekt/query"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/comby"
	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/backend"
	"github.com/sourcegraph/sourcegraph/internal/search/zoektquery"
	"github.com/sourcegraph/sourcegraph/internal/searcher/protocol"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func handleSearchFilters(patternInfo *protocol.PatternInfo) (query.Q, error) {
	var and []query.Q

	// Zoekt uses regular expressions for file paths.
	// Unhandled cases: PathPatternsAreCaseSensitive and whitespace in file path patterns.
	for _, p := range patternInfo.IncludePaths {
		q, err := zoektquery.FileRe(p, patternInfo.IsCaseSensitive)
		if err != nil {
			return nil, err
		}
		and = append(and, q)
	}
	if patternInfo.ExcludePaths != "" {
		q, err := zoektquery.FileRe(patternInfo.ExcludePaths, patternInfo.IsCaseSensitive)
		if err != nil {
			return nil, err
		}
		and = append(and, &query.Not{Child: q})
	}

	for _, lang := range patternInfo.IncludeLangs {
		and = append(and, &query.Language{Language: lang})
	}

	for _, lang := range patternInfo.ExcludeLangs {
		and = append(and, &query.Not{Child: &query.Language{Language: lang}})
	}

	return query.NewAnd(and...), nil
}

func buildQuery(pattern string, branchRepos []query.BranchRepos, filterQuery query.Q, shortcircuit bool) (query.Q, error) {
	regexString := comby.StructuralPatToRegexpQuery(pattern, shortcircuit)
	if len(regexString) == 0 {
		return &query.Const{Value: true}, nil
	}
	re, err := syntax.Parse(regexString, syntax.ClassNL|syntax.PerlX|syntax.UnicodeGroups)
	if err != nil {
		return nil, err
	}
	return query.NewAnd(
		&query.BranchesRepos{List: branchRepos},
		filterQuery,
		&query.Regexp{
			Regexp:        re,
			CaseSensitive: true,
			Content:       true,
		},
	), nil
}

// zoektSearch searches repositories using zoekt, returning file contents for
// files that match the given pattern.
//
// Timeouts are reported through the context, and as a special case errNoResultsInTimeout
// is returned if no results are found in the given timeout (instead of the more common
// case of finding partial or full results in the given timeout).
func zoektSearch(ctx context.Context, logger log.Logger, client zoekt.Streamer, args *protocol.PatternInfo, branchRepos []query.BranchRepos, contextLines int32, since func(t time.Time) time.Duration, repo api.RepoName, sender matchSender) (err error) {
	if len(branchRepos) == 0 {
		return nil
	}

	atom, err := extractQueryAtom(args)
	if err != nil {
		return err
	}

	searchOpts := (&search.ZoektParameters{
		FileMatchLimit:  int32(args.Limit),
		NumContextLines: int(contextLines),
	}).ToSearchOptions(ctx)
	searchOpts.Whole = true

	filterQuery, err := handleSearchFilters(args)
	if err != nil {
		return err
	}

	t0 := time.Now()
	q, err := buildQuery(atom.Value, branchRepos, filterQuery, false)
	if err != nil {
		return err
	}

	var extensionHint string
	if len(args.IncludePaths) > 0 {
		// Remove anchor that's added by autocomplete
		extensionHint = strings.TrimSuffix(filepath.Ext(args.IncludePaths[0]), "$")
	}

	pool := pool.New().WithErrors()
	tarInputEventC := make(chan comby.TarInputEvent)
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	pool.Go(func() error {
		// Cancel the context on completion so that the writer doesn't
		// block indefinitely if this stops reading.
		defer cancel()
		return structuralSearch(ctx, logger, comby.Tar{TarInputEventC: tarInputEventC}, all, extensionHint, atom.Value, args.CombyRule, args.Languages, repo, int32(searchOpts.NumContextLines), sender)
	})

	pool.Go(func() error {
		defer close(tarInputEventC)

		return client.StreamSearch(ctx, q, searchOpts, backend.ZoektStreamFunc(func(event *zoekt.SearchResult) {
			for _, file := range event.Files {
				hdr := tar.Header{
					Name: file.FileName,
					Mode: 0600,
					Size: int64(len(file.Content)),
				}
				tarInput := comby.TarInputEvent{
					Header:  hdr,
					Content: file.Content,
				}
				select {
				case tarInputEventC <- tarInput:
				case <-ctx.Done():
					return
				}
			}
		}))
	})

	err = pool.Wait()
	if err != nil {
		return err
	}
	if since(t0) >= searchOpts.MaxWallTime {
		return errNoResultsInTimeout
	}

	return nil
}

var errNoResultsInTimeout = errors.New("no results found in specified timeout")
