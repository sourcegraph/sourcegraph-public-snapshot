package search

import (
	"archive/tar"
	"context"
	"path/filepath"
	"regexp/syntax" //nolint:depguard // zoekt requires this pkg
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/zoekt"
	zoektquery "github.com/google/zoekt/query"

	"github.com/sourcegraph/log"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/comby"
	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/backend"
	zoektutil "github.com/sourcegraph/sourcegraph/internal/search/zoekt"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var (
	zoektOnce   sync.Once
	endpointMap atomicEndpoints
	zoektClient zoekt.Streamer
)

func getZoektClient(indexerEndpoints []string) zoekt.Streamer {
	zoektOnce.Do(func() {
		zoektClient = backend.NewMeteredSearcher(
			"", // no hostname means its the aggregator
			&backend.HorizontalSearcher{
				Map:  &endpointMap,
				Dial: backend.ZoektDial,
			},
		)
	})
	endpointMap.Set(indexerEndpoints)
	return zoektClient
}

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
func zoektSearch(ctx context.Context, args *search.TextPatternInfo, branchRepos []zoektquery.BranchRepos, since func(t time.Time) time.Duration, endpoints []string, repo api.RepoName, sender matchSender) (err error) {
	if len(branchRepos) == 0 {
		return nil
	}

	numRepos := 0
	for _, br := range branchRepos {
		numRepos += int(br.Repos.GetCardinality())
	}

	// Choose sensible values for k when we generalize this.
	k := zoektutil.ResultCountFactor(numRepos, args.FileMatchLimit, false)
	searchOpts := zoektutil.SearchOpts(ctx, k, args.FileMatchLimit, nil)
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

	tarInputEventC := make(chan comby.TarInputEvent)
	wg := sync.WaitGroup{}
	defer wg.Wait()

	var extensionHint string
	if len(args.IncludePatterns) > 0 {
		// Remove anchor that's added by autocomplete
		extensionHint = strings.TrimSuffix(filepath.Ext(args.IncludePatterns[0]), "$")
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		err := structuralSearch(ctx, comby.Tar{TarInputEventC: tarInputEventC}, all, extensionHint, args.Pattern, args.CombyRule, args.Languages, repo, sender)
		if err != nil {
			log.NamedError("structural search error", err)
		}
	}()

	client := getZoektClient(endpoints)
	err = client.StreamSearch(ctx, q, &searchOpts, backend.ZoektStreamFunc(func(event *zoekt.SearchResult) {
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
			tarInputEventC <- tarInput
		}
	}))
	close(tarInputEventC)

	if err != nil {
		return err
	}
	if since(t0) >= searchOpts.MaxWallTime {
		return errNoResultsInTimeout
	}

	return nil
}

var errNoResultsInTimeout = errors.New("no results found in specified timeout")

// atomicEndpoints allows us to update the endpoints used by our zoekt client.
type atomicEndpoints struct {
	endpoints atomic.Value
}

func (a *atomicEndpoints) Endpoints() ([]string, error) {
	eps := a.endpoints.Load()
	if eps == nil {
		return nil, errors.New("endpoints have not been set")
	}
	return eps.([]string), nil
}

func (a *atomicEndpoints) Set(endpoints []string) {
	if !a.needsUpdate(endpoints) {
		return
	}
	a.endpoints.Store(endpoints)
}

func (a *atomicEndpoints) needsUpdate(endpoints []string) bool {
	old, err := a.Endpoints()
	if err != nil {
		return true
	}
	if len(old) != len(endpoints) {
		return true
	}

	for i := range endpoints {
		if old[i] != endpoints[i] {
			return true
		}
	}

	return false
}
