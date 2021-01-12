package search

import (
	"context"
	"net/url"
	"strings"
	"time"

	"github.com/google/zoekt"
	"github.com/inconshreveable/log15"
	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/internal/trace/ot"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

var (
	// The default timeout to use for queries.
	defaultTimeout = 20 * time.Second
)

const defaultMaxSearchResults = 30

func zoektResultCountFactor(numRepos int, fileMatchLimit int32, globalSearch bool) (k int) {
	if globalSearch {
		// for globalSearch, numRepos = 0, but effectively we are searching over all
		// indexed repos, hence k should be 1
		k = 1
	} else {
		// If we're only searching a small number of repositories, return more
		// comprehensive results. This is arbitrary.
		switch {
		case numRepos <= 5:
			k = 100
		case numRepos <= 10:
			k = 10
		case numRepos <= 25:
			k = 8
		case numRepos <= 50:
			k = 5
		case numRepos <= 100:
			k = 3
		case numRepos <= 500:
			k = 2
		default:
			k = 1
		}
	}
	if fileMatchLimit > defaultMaxSearchResults {
		k = int(float64(k) * 3 * float64(fileMatchLimit) / float64(defaultMaxSearchResults))
	}
	return k
}

func getSpanContext(ctx context.Context) (shouldTrace bool, spanContext map[string]string) {
	if !ot.ShouldTrace(ctx) {
		return false, nil
	}

	spanContext = make(map[string]string)
	if err := ot.GetTracer(ctx).Inject(opentracing.SpanFromContext(ctx).Context(), opentracing.TextMap, opentracing.TextMapCarrier(spanContext)); err != nil {
		log15.Warn("Error injecting span context into map: %s", err)
		return true, nil
	}
	return true, spanContext
}

func zoektSearchOpts(ctx context.Context, k int, query *search.TextPatternInfo) zoekt.SearchOptions {
	shouldTrace, spanContext := getSpanContext(ctx)
	searchOpts := zoekt.SearchOptions{
		Trace:                  shouldTrace,
		SpanContext:            spanContext,
		MaxWallTime:            defaultTimeout,
		ShardMaxMatchCount:     100 * k,
		TotalMaxMatchCount:     100 * k,
		ShardMaxImportantMatch: 15 * k,
		TotalMaxImportantMatch: 25 * k,
		MaxDocDisplayCount:     2 * defaultMaxSearchResults,
	}

	// We want zoekt to return more than FileMatchLimit results since we use
	// the extra results to populate reposLimitHit. Additionally the defaults
	// are very low, so we always want to return at least 2000.
	if query.FileMatchLimit > defaultMaxSearchResults {
		searchOpts.MaxDocDisplayCount = 2 * int(query.FileMatchLimit)
	}
	if searchOpts.MaxDocDisplayCount < 2000 {
		searchOpts.MaxDocDisplayCount = 2000
	}

	if userProbablyWantsToWaitLonger := query.FileMatchLimit > defaultMaxSearchResults; userProbablyWantsToWaitLonger {
		searchOpts.MaxWallTime *= time.Duration(3 * float64(query.FileMatchLimit) / float64(defaultMaxSearchResults))
	}

	return searchOpts
}

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

var errNoResultsInTimeout = errors.New("no results found in specified timeout")

// zoektLimitMatches is the logic which limits files based on
// limit. Additionally it calculates the set of repos with partial
// results. This information is not returned by zoekt, so if zoekt indicates a
// limit has been hit, we include all repos in partial.
func zoektLimitMatches(limitHit bool, limit int, files []zoekt.FileMatch, getRepoInputRev func(file *zoekt.FileMatch) (repo *types.RepoName, revs []string, ok bool)) (bool, []zoekt.FileMatch, map[api.RepoID]struct{}) {
	var resultFiles []zoekt.FileMatch
	var partialFiles []zoekt.FileMatch

	resultFiles = files
	if limitHit {
		partialFiles = files
	}

	if len(files) > limit {
		resultFiles = files[:limit]
		if !limitHit {
			limitHit = true
			partialFiles = files[limit:]
		}
	}

	if len(partialFiles) == 0 {
		return limitHit, resultFiles, nil
	}

	partial := make(map[api.RepoID]struct{})
	last := ""
	for _, file := range partialFiles {
		// PERF: skip lookup if it is the same repo as the last result
		if file.Repository == last {
			continue
		}
		last = file.Repository

		if repo, _, ok := getRepoInputRev(&file); ok {
			partial[repo.ID] = struct{}{}
		}
	}

	return limitHit, resultFiles, partial
}

// zoektSearch searches repositories using zoekt, returning only the file paths containing
// content matching the given pattern.
//
// Timeouts are reported through the context, and as a special case errNoResultsInTimeout
// is returned if no results are found in the given timeout (instead of the more common
// case of finding partial or full results in the given timeout).
func zoektSearch(ctx context.Context, args *search.TextParameters, repos *graphqlbackend.IndexedRepoRevs, since func(t time.Time) time.Duration) (fm []*graphqlbackend.FileMatchResolver, limitHit bool, partial map[api.RepoID]struct{}, err error) {
	if len(repos.RepoRevs) == 0 {
		return nil, false, nil, nil
	}

	k := zoektResultCountFactor(len(repos.RepoBranches), args.PatternInfo.FileMatchLimit, args.Mode == search.ZoektGlobalSearch)
	searchOpts := zoektSearchOpts(ctx, k, args.PatternInfo)

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

	filePathPatterns, err := graphqlbackend.HandleFilePathPatterns(args.PatternInfo)
	if err != nil {
		return nil, false, nil, err
	}

	t0 := time.Now()
	q, err := graphqlbackend.BuildQuery(args, repos, filePathPatterns, true)
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
		q, err = graphqlbackend.BuildQuery(args, repos, filePathPatterns, false)
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

	limitHit, files, partial := zoektLimitMatches(limitHit, int(args.PatternInfo.FileMatchLimit), resp.Files, func(file *zoekt.FileMatch) (repo *types.RepoName, revs []string, ok bool) {
		repo, inputRevs := repos.GetRepoInputRev(file)
		return repo, inputRevs, true
	})
	resp.Files = files

	maxLineMatches := 25 + k
	matches := make([]*graphqlbackend.FileMatchResolver, len(resp.Files))
	repoResolvers := make(graphqlbackend.RepositoryResolverCache)
	for i, file := range resp.Files {
		fileLimitHit := false
		if len(file.LineMatches) > maxLineMatches {
			file.LineMatches = file.LineMatches[:maxLineMatches]
			fileLimitHit = true
			limitHit = true
		}
		repoRev := repos.RepoRevs[file.Repository]
		if repoResolvers[repoRev.Repo.Name] == nil {
			repoResolvers[repoRev.Repo.Name] = graphqlbackend.NewRepositoryResolver(repoRev.Repo.ToRepo())
		}
		matches[i] = &graphqlbackend.FileMatchResolver{
			JPath:     file.FileName,
			JLimitHit: fileLimitHit,
			URI:       fileMatchURI(repoRev.Repo.Name, "", file.FileName),
			Repo:      repoResolvers[repoRev.Repo.Name],
			CommitID:  api.CommitID(file.Version),
		}
	}

	return matches, limitHit, partial, nil
}

func fileMatchURI(name api.RepoName, ref, path string) string {
	var b strings.Builder
	ref = url.QueryEscape(ref)
	b.Grow(len(name) + len(ref) + len(path) + len("git://?#"))
	b.WriteString("git://")
	b.WriteString(string(name))
	if ref != "" {
		b.WriteByte('?')
		b.WriteString(ref)
	}
	b.WriteByte('#')
	b.WriteString(path)
	return b.String()
}
