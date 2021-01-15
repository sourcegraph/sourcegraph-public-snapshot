package search

import (
	"archive/zip"
	"context"
	"errors"
	"fmt"
	"os"
	"regexp/syntax"
	"time"

	"github.com/google/zoekt"
	zoektquery "github.com/google/zoekt/query"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/comby"
	"github.com/sourcegraph/sourcegraph/internal/search"
	zoektutil "github.com/sourcegraph/sourcegraph/internal/search/zoekt"
	"github.com/sourcegraph/sourcegraph/internal/trace"
)

func HandleFilePathPatterns(query *search.TextPatternInfo) (zoektquery.Q, error) {
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

func buildQuery(args *search.TextParameters, repoBranches map[string][]string, filePathPatterns zoektquery.Q, shortcircuit bool) (zoektquery.Q, error) {
	regexString := comby.StructuralPatToRegexpQuery(args.PatternInfo.Pattern, shortcircuit)
	if len(regexString) == 0 {
		return &zoektquery.Const{Value: true}, nil
	}
	re, err := syntax.Parse(regexString, syntax.ClassNL|syntax.PerlX|syntax.UnicodeGroups)
	if err != nil {
		return nil, err
	}
	return zoektquery.NewAnd(
		&zoektquery.RepoBranches{Set: repoBranches},
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

func zoektSearchHEADOnlyFilesStream(ctx context.Context, args *search.TextParameters, repoBranches map[string][]string, since func(t time.Time) time.Duration) <-chan zoektSearchStreamEvent {
	c := make(chan zoektSearchStreamEvent)
	go func() {
		defer close(c)
		_, _, _, _ = zoektSearch(ctx, args, repoBranches, since, c)
	}()

	return c
}

const defaultMaxSearchResults = 30

// zoektSearch searches repositories using zoekt, returning only the file paths containing
// content matching the given pattern.
//
// Timeouts are reported through the context, and as a special case errNoResultsInTimeout
// is returned if no results are found in the given timeout (instead of the more common
// case of finding partial or full results in the given timeout).
func zoektSearch(ctx context.Context, args *search.TextParameters, repoBranches map[string][]string, since func(t time.Time) time.Duration, c chan<- zoektSearchStreamEvent) (fm []zoekt.FileMatch, limitHit bool, partial map[api.RepoID]struct{}, err error) {
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
	if len(repoBranches) == 0 {
		return nil, false, nil, nil
	}

	k := zoektutil.ResultCountFactor(len(repoBranches), args.PatternInfo.FileMatchLimit, args.Mode == search.ZoektGlobalSearch)
	searchOpts := zoektutil.SearchOpts(ctx, k, args.PatternInfo)
	searchOpts.Whole = true

	if args.UseFullDeadline {
		// If the user manually specified a timeout, allow zoekt to use all of the remaining timeout.
		deadline, _ := ctx.Deadline()
		searchOpts.MaxWallTime = time.Until(deadline)

		// We don't want our context's deadline to cut off zoekt so that we can get the results
		// found before the deadline.
		//
		// We'll create a new context that gets cancelled if the other context is cancelled for any
		// reason other than the deadline being exceeded. This essentially means the deadline for the new context
		// will be `deadline + time for zoekt to cancel + network latency`.
		var cancel context.CancelFunc
		ctx, cancel = contextWithoutDeadline(ctx)
		defer cancel()
	}

	filePathPatterns, err := HandleFilePathPatterns(args.PatternInfo)
	if err != nil {
		return nil, false, nil, err
	}

	t0 := time.Now()
	q, err := buildQuery(args, repoBranches, filePathPatterns, true)
	if err != nil {
		return nil, false, nil, err
	}
	resp, err := args.Zoekt.Client.Search(ctx, q, &searchOpts)
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
	if resp.FileCount < 10 || args.PatternInfo.FileMatchLimit != defaultMaxSearchResults {
		q, err = buildQuery(args, repoBranches, filePathPatterns, false)
		if err != nil {
			return nil, false, nil, err
		}
		resp, err = args.Zoekt.Client.Search(ctx, q, &searchOpts)
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

	// TODO make this work
	// limitHit, files, partial := zoektLimitMatches(limitHit, int(args.PatternInfo.FileMatchLimit), resp.Files, func(file *zoekt.FileMatch) (repo *types.RepoName, revs []string, ok bool) {
	// 	repo, inputRevs := repos.GetRepoInputRev(file)
	// 	return repo, inputRevs, true
	// })
	// resp.Files = files

	maxLineMatches := 25 + k
	for _, file := range resp.Files {
		if len(file.LineMatches) > maxLineMatches {
			file.LineMatches = file.LineMatches[:maxLineMatches]
			limitHit = true
		}
	}

	return resp.Files, limitHit, partial, nil
}

// TODO is this necessary? It looks like a request will always only be for one repo
// func groupFilesByRepo(fileMatches []zoekt.FileMatch) map[string][]zoekt.FileMatch {
// 	m := make(map[string][]zoekt.FileMatch)
// 	for _, match := range fileMatches {
// 		m[match.Repository] = append(m[match.Repository], match)
// 	}
// 	return m
// }

func writeZip(path string, fileMatches []zoekt.FileMatch) (err error) {
	f, err := os.Create(path)
	if err != nil {
		return err
	}

	// TODO is this handled outside of here?
	defer func() {
		if err != nil {
			rmErr := os.Remove(path)
			if rmErr != nil {
				// TODO is there a conventional way to handle errors during cleanup?
				err = fmt.Errorf("error1: %s, error2: %s", err, rmErr)
			}
		}
	}()
	defer f.Close()

	zw := zip.NewWriter(f)
	defer zw.Close()

	for _, match := range fileMatches {
		mw, err := zw.Create(match.FileName)
		if err != nil {
			return err
		}

		_, err = mw.Write(match.Content)
		if err != nil {
			return err
		}
	}

	return nil
}

var errNoResultsInTimeout = errors.New("no results found in specified timeout")

// contextWithoutDeadline returns a context which will cancel if the cOld is
// canceled.
func contextWithoutDeadline(cOld context.Context) (context.Context, context.CancelFunc) {
	cNew, cancel := context.WithCancel(context.Background())

	// Set trace context so we still get spans propagated
	cNew = trace.CopyContext(cNew, cOld)

	go func() {
		select {
		case <-cOld.Done():
			// cancel the new context if the old one is done for some reason other than the deadline passing.
			if cOld.Err() != context.DeadlineExceeded {
				cancel()
			}
		case <-cNew.Done():
		}
	}()

	return cNew, cancel
}
