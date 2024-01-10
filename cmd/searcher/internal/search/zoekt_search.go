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
	zoektquery "github.com/sourcegraph/zoekt/query"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/comby"
	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/backend"
	zoektutil "github.com/sourcegraph/sourcegraph/internal/search/zoekt"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func handleFilePathPatterns(query *search.TextPatternInfo) (zoektquery.Q, error) {
	var and []zoektquery.Q

	// Zoekt uses regular expressions for file paths.
	// Unhandled cases: PathPatternsAreCaseSensitive and whitespace in file path patterns.
	for _, p := range query.IncludePatterns {
		q, err := zoektutil.FileRe(p, query.IsCaseSensitive)
		if err != nil {
			return nil, err
		}
		and = append(and, q)
	}
	if query.ExcludePattern != "" {
		q, err := zoektutil.FileRe(query.ExcludePattern, query.IsCaseSensitive)
		if err != nil {
			return nil, err
		}
		and = append(and, &zoektquery.Not{Child: q})
	}

	return zoektquery.NewAnd(and...), nil
}

func buildQuery(args *search.TextPatternInfo, branchRepos []zoektquery.BranchRepos, filePathPatterns zoektquery.Q, shortcircuit bool) (zoektquery.Q, error) {
	regexString := comby.StructuralPatToRegexpQuery(args.Pattern, shortcircuit)
	if len(regexString) == 0 {
		return &zoektquery.Const{Value: true}, nil
	}
	re, err := syntax.Parse(regexString, syntax.ClassNL|syntax.PerlX|syntax.UnicodeGroups)
	if err != nil {
		return nil, err
	}
	return zoektquery.NewAnd(
		&zoektquery.BranchesRepos{List: branchRepos},
		filePathPatterns,
		&zoektquery.Regexp{
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
func zoektSearch(ctx context.Context, logger log.Logger, client zoekt.Streamer, args *search.TextPatternInfo, branchRepos []zoektquery.BranchRepos, contextLines int32, since func(t time.Time) time.Duration, repo api.RepoName, sender matchSender) (err error) {
	if len(branchRepos) == 0 {
		return nil
	}

	searchOpts := (&search.ZoektParameters{
		FileMatchLimit:  args.FileMatchLimit,
		NumContextLines: int(contextLines),
	}).ToSearchOptions(ctx)
	searchOpts.Whole = true

	filePathPatterns, err := handleFilePathPatterns(args)
	if err != nil {
		return err
	}

	t0 := time.Now()
	q, err := buildQuery(args, branchRepos, filePathPatterns, false)
	if err != nil {
		return err
	}

	var extensionHint string
	if len(args.IncludePatterns) > 0 {
		// Remove anchor that's added by autocomplete
		extensionHint = strings.TrimSuffix(filepath.Ext(args.IncludePatterns[0]), "$")
	}

	pool := pool.New().WithErrors()
	tarInputEventC := make(chan comby.TarInputEvent)
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	pool.Go(func() error {
		// Cancel the context on completion so that the writer doesn't
		// block indefinitely if this stops reading.
		defer cancel()
		return structuralSearch(ctx, logger, comby.Tar{TarInputEventC: tarInputEventC}, all, extensionHint, args.Pattern, args.CombyRule, args.Languages, repo, int32(searchOpts.NumContextLines), sender)
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
