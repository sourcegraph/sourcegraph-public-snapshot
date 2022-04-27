package search

import (
	"archive/zip"
	"context"
	"io"
	"regexp/syntax" //nolint:depguard // zoekt requires this pkg
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/zoekt"
	zoektquery "github.com/google/zoekt/query"
	"github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/comby"
	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/backend"
	zoektutil "github.com/sourcegraph/sourcegraph/internal/search/zoekt"
	"github.com/sourcegraph/sourcegraph/internal/trace/ot"
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

	// For conditionals that happen on a repo we can use type:repo queries. eg
	// (type:repo file:foo) (type:repo file:bar) will match all repos which
	// contain a filename matching "foo" and a filename matchinb "bar".
	//
	// Note: (type:repo file:foo file:bar) will only find repos with a
	// filename containing both "foo" and "bar".
	for _, p := range query.FilePatternsReposMustInclude {
		q, err := zoektutil.FileRe(p, query.IsCaseSensitive)
		if err != nil {
			return nil, err
		}
		and = append(and, &zoektquery.Type{Type: zoektquery.TypeRepo, Child: q})
	}
	for _, p := range query.FilePatternsReposMustExclude {
		q, err := zoektutil.FileRe(p, query.IsCaseSensitive)
		if err != nil {
			return nil, err
		}
		and = append(and, &zoektquery.Not{Child: &zoektquery.Type{Type: zoektquery.TypeRepo, Child: q}})
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

type zoektSearchStreamEvent struct {
	fm       []zoekt.FileMatch
	limitHit bool
	partial  map[api.RepoID]struct{}
	err      error
}

const defaultMaxSearchResults = 30

// zoektSearch searches repositories using zoekt, returning file contents for
// files that match the given pattern.
//
// Timeouts are reported through the context, and as a special case errNoResultsInTimeout
// is returned if no results are found in the given timeout (instead of the more common
// case of finding partial or full results in the given timeout).
func zoektSearch(ctx context.Context, args *search.TextPatternInfo, branchRepos []zoektquery.BranchRepos, since func(t time.Time) time.Duration, endpoints []string, c chan<- zoektSearchStreamEvent) (fm []zoekt.FileMatch, limitHit bool, partial map[api.RepoID]struct{}, err error) {
	defer func() {
		if c != nil {
			c <- zoektSearchStreamEvent{
				fm:       fm,
				limitHit: limitHit,
				partial:  partial,
				err:      err,
			}
		}
	}()
	if len(branchRepos) == 0 {
		return nil, false, nil, nil
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
		return nil, false, nil, err
	}

	t0 := time.Now()
	q, err := buildQuery(args, branchRepos, filePathPatterns, true)
	if err != nil {
		return nil, false, nil, err
	}

	client := getZoektClient(endpoints)
	resp, err := client.Search(ctx, q, &searchOpts)
	if err != nil {
		return nil, false, nil, err
	}
	if since(t0) >= searchOpts.MaxWallTime {
		return nil, false, nil, errNoResultsInTimeout
	}

	// We always return approximate results (limitHit true) unless we run the branch to perform a more complete search.
	limitHit = true
	// If the previous indexed search did not return a substantial number of matching file candidates or count was
	// manually specified, run a more complete and expensive search.
	if resp.FileCount < 10 || args.FileMatchLimit != defaultMaxSearchResults {
		q, err = buildQuery(args, branchRepos, filePathPatterns, false)
		if err != nil {
			return nil, false, nil, err
		}
		resp, err = client.Search(ctx, q, &searchOpts)
		if err != nil {
			return nil, false, nil, err
		}
		if since(t0) >= searchOpts.MaxWallTime {
			return nil, false, nil, errNoResultsInTimeout
		}
		// This is the only place limitHit can be set false, meaning we covered everything.
		limitHit = resp.FilesSkipped+resp.ShardsSkipped > 0
	}

	if len(resp.Files) == 0 {
		return nil, false, nil, nil
	}

	maxLineMatches := 25 + k
	for _, file := range resp.Files {
		if len(file.LineMatches) > maxLineMatches {
			file.LineMatches = file.LineMatches[:maxLineMatches]
			limitHit = true
		}
	}

	return resp.Files, limitHit, partial, nil
}

func writeZip(ctx context.Context, w io.Writer, fileMatches []zoekt.FileMatch) (err error) {
	bytesWritten := 0
	span, _ := ot.StartSpanFromContext(ctx, "WriteZip")
	defer func() {
		span.LogFields(log.Int("bytes_written", bytesWritten))
		span.Finish()
	}()

	zw := zip.NewWriter(w)
	defer zw.Close()

	for _, match := range fileMatches {
		mw, err := zw.Create(match.FileName)
		if err != nil {
			return err
		}

		n, err := mw.Write(match.Content)
		if err != nil {
			return err
		}
		bytesWritten += n
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
