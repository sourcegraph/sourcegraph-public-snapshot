package graphqlbackend

import (
	"context"
	"fmt"
	"math"
	"net/url"
	"regexp/syntax"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/zoekt"
	zoektquery "github.com/google/zoekt/query"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/gituri"
	"github.com/sourcegraph/sourcegraph/internal/search"
	searchbackend "github.com/sourcegraph/sourcegraph/internal/search/backend"
	"github.com/sourcegraph/sourcegraph/internal/symbols/protocol"
	"github.com/sourcegraph/sourcegraph/internal/trace"
)

func zoektResultCountFactor(numRepos int, query *search.TextPatternInfo) int {
	// If we're only searching a small number of repositories, return more comprehensive results. This is
	// arbitrary.
	k := 1
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
	}
	if query.FileMatchLimit > defaultMaxSearchResults {
		k = int(float64(k) * 3 * float64(query.FileMatchLimit) / float64(defaultMaxSearchResults))
	}
	return k
}

func zoektSearchOpts(k int, query *search.TextPatternInfo) zoekt.SearchOptions {
	searchOpts := zoekt.SearchOptions{
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

var errNoResultsInTimeout = errors.New("no results found in specified timeout")

// zoektSearchHEAD searches repositories using zoekt.
//
// Timeouts are reported through the context, and as a special case errNoResultsInTimeout
// is returned if no results are found in the given timeout (instead of the more common
// case of finding partial or full results in the given timeout).
func zoektSearchHEAD(ctx context.Context, args *search.TextParameters, repos []*search.RepositoryRevisions, isSymbol bool, since func(t time.Time) time.Duration) (fm []*FileMatchResolver, limitHit bool, reposLimitHit map[string]struct{}, err error) {
	if len(repos) == 0 {
		return nil, false, nil, nil
	}

	// Tell zoekt which repos to search
	repoSet := &zoektquery.RepoSet{Set: make(map[string]bool, len(repos))}
	repoMap := make(map[api.RepoName]*search.RepositoryRevisions, len(repos))
	for _, repoRev := range repos {
		repoSet.Set[string(repoRev.Repo.Name)] = true
		repoMap[api.RepoName(strings.ToLower(string(repoRev.Repo.Name)))] = repoRev
	}

	queryExceptRepos, err := queryToZoektQuery(args.PatternInfo, isSymbol)
	if err != nil {
		return nil, false, nil, err
	}
	finalQuery := zoektquery.NewAnd(repoSet, queryExceptRepos)

	tr, ctx := trace.New(ctx, "zoekt.Search", fmt.Sprintf("%d %+v", len(repoSet.Set), finalQuery.String()))
	defer func() {
		tr.SetError(err)
		if len(fm) > 0 {
			tr.LazyPrintf("%d file matches", len(fm))
		}
		tr.Finish()
	}()

	k := zoektResultCountFactor(len(repos), args.PatternInfo)
	searchOpts := zoektSearchOpts(k, args.PatternInfo)

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
		cNew, cancel := context.WithCancel(context.Background())
		go func(cOld context.Context) {
			<-cOld.Done()
			// cancel the new context if the old one is done for some reason other than the deadline passing.
			if cOld.Err() != context.DeadlineExceeded {
				cancel()
			}
		}(ctx)
		ctx = cNew
		defer cancel()
	}

	// If the query has a `repohasfile` or `-repohasfile` flag, we want to construct a new reposet based
	// on the values passed in to the flag.
	newRepoSet, err := createNewRepoSetWithRepoHasFileInputs(ctx, args.PatternInfo, args.Zoekt.Client, repoSet)
	if err != nil {
		return nil, false, nil, err
	}
	finalQuery = zoektquery.NewAnd(newRepoSet, queryExceptRepos)
	tr.LazyPrintf("after repohasfile filters: nRepos=%d query=%v", len(newRepoSet.Set), finalQuery)

	t0 := time.Now()
	resp, err := args.Zoekt.Client.Search(ctx, finalQuery, &searchOpts)
	if err != nil {
		return nil, false, nil, err
	}
	if resp.FileCount == 0 && resp.MatchCount == 0 && since(t0) >= searchOpts.MaxWallTime {
		return nil, false, nil, errNoResultsInTimeout
	}
	limitHit = resp.FilesSkipped+resp.ShardsSkipped > 0
	// Repositories that weren't fully evaluated because they hit the Zoekt or Sourcegraph file match limits.
	reposLimitHit = make(map[string]struct{})
	if limitHit {
		// Zoekt either did not evaluate some files in repositories, or ignored some repositories altogether.
		// In this case, we can't be sure that we have exhaustive results for _any_ repository. So, all file
		// matches are from repos with potentially skipped matches.
		for _, file := range resp.Files {
			if _, ok := reposLimitHit[file.Repository]; !ok {
				reposLimitHit[file.Repository] = struct{}{}
			}
		}
	}

	if len(resp.Files) == 0 {
		return nil, false, nil, nil
	}

	maxLineMatches := 25 + k
	maxLineFragmentMatches := 3 + k
	if limit := int(args.PatternInfo.FileMatchLimit); len(resp.Files) > limit {
		// List of files we cut out from the Zoekt response because they exceed the file match limit on the Sourcegraph end.
		// We use this to get a list of repositories that do not have complete results.
		fileMatchesInSkippedRepos := resp.Files[limit:]
		resp.Files = resp.Files[:limit]

		if !limitHit {
			// Zoekt evaluated all files and repositories, but Zoekt returned more file matches
			// than the limit we set on Sourcegraph, so we cut out more results.

			// Generate a list of repositories that had results cut because they exceeded the file match limit set on Sourcegraph.
			for _, file := range fileMatchesInSkippedRepos {
				if _, ok := reposLimitHit[file.Repository]; !ok {
					reposLimitHit[file.Repository] = struct{}{}
				}
			}
		}

		limitHit = true
	}

	matches := make([]*FileMatchResolver, len(resp.Files))
	for i, file := range resp.Files {
		fileLimitHit := false
		if len(file.LineMatches) > maxLineMatches {
			file.LineMatches = file.LineMatches[:maxLineMatches]
			fileLimitHit = true
			limitHit = true
		}
		repoRev := repoMap[api.RepoName(strings.ToLower(string(file.Repository)))]
		inputRev := repoRev.RevSpecs()[0]
		baseURI := &gituri.URI{URL: url.URL{Scheme: "git://", Host: string(repoRev.Repo.Name), RawQuery: "?" + url.QueryEscape(inputRev)}}
		lines := make([]*lineMatch, 0, len(file.LineMatches))
		symbols := []*searchSymbolResult{}
		for _, l := range file.LineMatches {
			if !l.FileName {
				if len(l.LineFragments) > maxLineFragmentMatches {
					l.LineFragments = l.LineFragments[:maxLineFragmentMatches]
				}
				offsets := make([][2]int32, len(l.LineFragments))
				for k, m := range l.LineFragments {
					offset := utf8.RuneCount(l.Line[:m.LineOffset])
					length := utf8.RuneCount(l.Line[m.LineOffset : m.LineOffset+m.MatchLength])
					offsets[k] = [2]int32{int32(offset), int32(length)}
					if isSymbol && m.SymbolInfo != nil {
						commit := &GitCommitResolver{
							repo:     &RepositoryResolver{repo: repoRev.Repo},
							oid:      GitObjectID(repoRev.IndexedHEADCommit()),
							inputRev: &inputRev,
						}

						symbols = append(symbols, &searchSymbolResult{
							symbol: protocol.Symbol{
								Name:       m.SymbolInfo.Sym,
								Kind:       m.SymbolInfo.Kind,
								Parent:     m.SymbolInfo.Parent,
								ParentKind: m.SymbolInfo.ParentKind,
								Path:       file.FileName,
								Line:       l.LineNumber,
							},
							lang:    strings.ToLower(file.Language),
							baseURI: baseURI,
							commit:  commit,
						})
					}
				}
				if !isSymbol {
					lines = append(lines, &lineMatch{
						JPreview:          string(l.Line),
						JLineNumber:       int32(l.LineNumber - 1),
						JOffsetAndLengths: offsets,
					})
				}
			}
		}
		matches[i] = &FileMatchResolver{
			JPath:        file.FileName,
			JLineMatches: lines,
			JLimitHit:    fileLimitHit,
			uri:          fileMatchURI(repoRev.Repo.Name, "", file.FileName),
			symbols:      symbols,
			Repo:         repoRev.Repo,
			CommitID:     repoRev.IndexedHEADCommit(),
		}
	}

	return matches, limitHit, reposLimitHit, nil
}

// createNewRepoSetWithRepoHasFileInputs mutates repoSet such that it accounts
// for the `repohasfile` and `-repohasfile` flags that may have been passed in
// the query. As a convenience it returns the mutated RepoSet.
func createNewRepoSetWithRepoHasFileInputs(ctx context.Context, query *search.TextPatternInfo, searcher zoekt.Searcher, repoSet *zoektquery.RepoSet) (*zoektquery.RepoSet, error) {
	// Shortcut if we have no repos to search
	if len(repoSet.Set) == 0 {
		return repoSet, nil
	}

	flagIsInQuery := len(query.FilePatternsReposMustInclude) > 0
	negatedFlagIsInQuery := len(query.FilePatternsReposMustExclude) > 0

	// Construct queries which search for repos containing the files passed into `repohasfile`
	filesToIncludeQueries, err := queryToZoektFileOnlyQueries(query, query.FilePatternsReposMustInclude)
	if err != nil {
		return nil, err
	}

	newSearchOpts := zoekt.SearchOptions{
		ShardMaxMatchCount:     1,
		TotalMaxMatchCount:     math.MaxInt32,
		ShardMaxImportantMatch: 1,
		TotalMaxImportantMatch: math.MaxInt32,
		MaxDocDisplayCount:     0,
	}
	newSearchOpts.SetDefaults()

	if flagIsInQuery {
		for _, q := range filesToIncludeQueries {
			// Shortcut if we have no repos to search
			if len(repoSet.Set) == 0 {
				return repoSet, nil
			}

			// Execute a new Zoekt search for each file passed in to a `repohasfile` flag.
			includeResp, err := searcher.Search(ctx, zoektquery.NewAnd(repoSet, q), &newSearchOpts)
			if err != nil {
				return nil, errors.Wrapf(err, "searching for %v", q.String())
			}

			newRepoSet := make(map[string]bool, len(includeResp.RepoURLs))
			for repoURL := range includeResp.RepoURLs {
				newRepoSet[repoURL] = true
			}

			// We want repoSet = repoSet intersect newRepoSet. but newRepoSet
			// is a subset, so we can just set repoSet = newRepoSet.
			repoSet.Set = newRepoSet
		}
	}

	// Construct queries which search for repos containing the files passed into `-repohasfile`
	filesToExcludeQueries, err := queryToZoektFileOnlyQueries(query, query.FilePatternsReposMustExclude)
	if err != nil {
		return nil, err
	}

	if negatedFlagIsInQuery {
		for _, q := range filesToExcludeQueries {
			// Shortcut if we have no repos to search
			if len(repoSet.Set) == 0 {
				return repoSet, nil
			}

			excludeResp, err := searcher.Search(ctx, zoektquery.NewAnd(repoSet, q), &newSearchOpts)
			if err != nil {
				return nil, err
			}
			for repoURL := range excludeResp.RepoURLs {
				// For each repo that had a result in the exclude set, if it exists in the repoSet, set the value to false so we don't search over it.
				if repoSet.Set[repoURL] {
					delete(repoSet.Set, repoURL)
				}
			}
		}
	}

	return repoSet, nil
}

func noOpAnyChar(re *syntax.Regexp) {
	if re.Op == syntax.OpAnyChar {
		re.Op = syntax.OpAnyCharNotNL
	}
	for _, s := range re.Sub {
		noOpAnyChar(s)
	}
}

func parseRe(pattern string, filenameOnly bool, queryIsCaseSensitive bool) (zoektquery.Q, error) {
	// these are the flags used by zoekt, which differ to searcher.
	re, err := syntax.Parse(pattern, syntax.ClassNL|syntax.PerlX|syntax.UnicodeGroups)
	if err != nil {
		return nil, err
	}
	noOpAnyChar(re)
	// zoekt decides to use its literal optimization at the query parser
	// level, so we check if our regex can just be a literal.
	if re.Op == syntax.OpLiteral {
		return &zoektquery.Substring{
			Pattern:       string(re.Rune),
			CaseSensitive: queryIsCaseSensitive,

			FileName: filenameOnly,
		}, nil
	}
	return &zoektquery.Regexp{
		Regexp:        re,
		CaseSensitive: queryIsCaseSensitive,

		FileName: filenameOnly,
	}, nil
}

func fileRe(pattern string, queryIsCaseSensitive bool) (zoektquery.Q, error) {
	return parseRe(pattern, true, queryIsCaseSensitive)
}

func queryToZoektQuery(query *search.TextPatternInfo, isSymbol bool) (zoektquery.Q, error) {
	var and []zoektquery.Q

	var q zoektquery.Q
	var err error
	if query.IsRegExp {
		fileNameOnly := query.PatternMatchesPath && !query.PatternMatchesContent
		q, err = parseRe(query.Pattern, fileNameOnly, query.IsCaseSensitive)
		if err != nil {
			return nil, err
		}
	} else {
		q = &zoektquery.Substring{
			Pattern:       query.Pattern,
			CaseSensitive: query.IsCaseSensitive,

			FileName: true,
			Content:  true,
		}
	}

	if isSymbol {
		q = &zoektquery.Symbol{
			Expr: q,
		}
	}

	and = append(and, q)

	// zoekt also uses regular expressions for file paths
	// TODO PathPatternsAreCaseSensitive
	// TODO whitespace in file path patterns?
	if !query.PathPatternsAreRegExps {
		return nil, errors.New("zoekt only supports regex path patterns")
	}
	for _, p := range query.IncludePatterns {
		q, err := fileRe(p, query.IsCaseSensitive)
		if err != nil {
			return nil, err
		}
		and = append(and, q)
	}
	if query.ExcludePattern != "" {
		q, err := fileRe(query.ExcludePattern, query.IsCaseSensitive)
		if err != nil {
			return nil, err
		}
		and = append(and, &zoektquery.Not{Child: q})
	}

	return zoektquery.Simplify(zoektquery.NewAnd(and...)), nil
}

// queryToZoektFileOnlyQueries constructs a list of Zoekt queries that search for a file pattern(s).
// `listOfFilePaths` specifies which field on `query` should be the list of file patterns to look for.
//  A separate zoekt query is created for each file path that should be searched.
func queryToZoektFileOnlyQueries(query *search.TextPatternInfo, listOfFilePaths []string) ([]zoektquery.Q, error) {
	var zoektQueries []zoektquery.Q
	if !query.PathPatternsAreRegExps {
		return nil, errors.New("zoekt only supports regex path patterns")
	}
	for _, p := range listOfFilePaths {
		q, err := fileRe(p, query.IsCaseSensitive)
		if err != nil {
			return nil, err
		}
		zoektQueries = append(zoektQueries, zoektquery.Simplify(q))
	}

	return zoektQueries, nil
}

func zoektSingleIndexedRepo(ctx context.Context, z *searchbackend.Zoekt, rev *search.RepositoryRevisions, filter func(*zoekt.Repository) bool) (indexed, unindexed []*search.RepositoryRevisions, err error) {
	indexed = []*search.RepositoryRevisions{}
	unindexed = []*search.RepositoryRevisions{}

	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()
	if len(rev.RevSpecs()) >= 2 || len(rev.RevSpecs()) != len(rev.Revs) {
		// Zoekt only indexes 1 rev per repository, so it will not have the full results for the
		// query on repositories for which multiple revs are searched.
		return indexed, append(unindexed, rev), nil
	}

	set, err := z.ListAll(ctx)
	if err != nil {
		return nil, nil, err
	}

	repo, ok := set[strings.ToLower(string(rev.Repo.Name))]
	if !ok || (filter != nil && !filter(repo)) {
		return indexed, append(unindexed, rev), nil
	}

	for _, branch := range repo.Branches {
		if branch.Name == "HEAD" {
			rev.SetIndexedHEADCommit(api.CommitID(branch.Version))
			break
		}
	}

	if len(rev.Revs) == 1 {
		revSpecToSearch := rev.Revs[0].RevSpec
		if len(revSpecToSearch) > 0 && len(revSpecToSearch) < 4 {
			// revSpecToSearch is nonempty but shorter than the
			// minimum 4 chars expected for a short SHA. It can't
			// match a commit, maybe it refers to a one-character
			// branch name.
			return indexed, append(unindexed, rev), nil
		}
		if revSpecToSearch == "" || revSpecToSearch == "HEAD" || strings.HasPrefix(string(rev.IndexedHEADCommit()), revSpecToSearch) {
			return append(indexed, rev), unindexed, nil
		}
	}

	return indexed, append(unindexed, rev), nil
}

// zoektIndexedRepos splits the input repo list into two parts: (1) the
// repositories `indexed` by Zoekt and (2) the repositories that are
// `unindexed`.
func zoektIndexedRepos(ctx context.Context, z *searchbackend.Zoekt, revs []*search.RepositoryRevisions, filter func(*zoekt.Repository) bool) (indexed, unindexed []*search.RepositoryRevisions, err error) {
	if len(revs) == 1 {
		// Classify indexed versus unindexed for the common case of a single revision
		return zoektSingleIndexedRepo(ctx, z, revs[0], filter)
	}

	count := 0
	for _, r := range revs {
		if len(r.Revs) > 0 && r.Revs[0].RevSpec == "" {
			count++
		}
	}

	// Return early if we don't need to querying zoekt
	if count == 0 {
		return nil, revs, nil
	}

	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()
	set, err := z.ListAll(ctx)
	if err != nil {
		return nil, nil, err
	}

	indexed = make([]*search.RepositoryRevisions, 0, count)
	unindexed = make([]*search.RepositoryRevisions, 0, len(revs)-count)

	for _, rev := range revs {
		if len(rev.RevSpecs()) >= 2 || len(rev.RevSpecs()) != len(rev.Revs) {
			// Zoekt only indexes 1 rev per repository, so it will not have the full results for the
			// query on repositories for which multiple revs are searched.
			unindexed = append(unindexed, rev)
			continue
		}

		repo, ok := set[strings.ToLower(string(rev.Repo.Name))]
		if !ok || (filter != nil && !filter(repo)) {
			unindexed = append(unindexed, rev)
			continue
		}

		for _, branch := range repo.Branches {
			if branch.Name == "HEAD" {
				rev.SetIndexedHEADCommit(api.CommitID(branch.Version))
				break
			}
		}

		indexed = append(indexed, rev)
	}

	return indexed, unindexed, nil
}
