package graphqlbackend

import (
	"context"
	"fmt"
	"net/url"
	"regexp/syntax"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/zoekt"
	zoektquery "github.com/google/zoekt/query"
	"github.com/inconshreveable/log15"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/gituri"
	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/query"
	"github.com/sourcegraph/sourcegraph/internal/symbols/protocol"
	"github.com/sourcegraph/sourcegraph/internal/trace"
)

type indexedRequestType string

const (
	textRequest   indexedRequestType = "text"
	symbolRequest indexedRequestType = "symbol"
	fileRequest   indexedRequestType = "file"
)

// indexedSearchRequest is responsible for translating a Sourcegraph search
// query into a Zoekt query and mapping the results from zoekt back to
// Sourcegraph result types.
type indexedSearchRequest struct {
	// Unindexed is a slice of repository revisions that can't be searched by
	// Zoekt. The repository revisions should be searched by the searcher
	// service.
	//
	// If IndexUnavailable is true or the query specifies index:no then all
	// repository revisions will be listed. Otherwise it will just be
	// repository revisions not indexed.
	Unindexed []*search.RepositoryRevisions

	// IndexUnavailable is true if zoekt is offline or disabled.
	IndexUnavailable bool

	// DisableUnindexedSearch is true if the query specified that only index
	// search should be used.
	DisableUnindexedSearch bool

	// inputs
	args *search.TextParameters
	typ  indexedRequestType

	// repos is the repository revisions that are indexed and will be
	// searched.
	repos *indexedRepoRevs

	// since if non-nil will be used instead of time.Since. For tests
	since func(time.Time) time.Duration
}

func newIndexedSearchRequest(ctx context.Context, args *search.TextParameters, typ indexedRequestType) (*indexedSearchRequest, error) {
	// Parse index:yes (default), index:only, and index:no in search query.
	indexParam := Yes
	if index, _ := args.Query.StringValues(query.FieldIndex); len(index) > 0 {
		index := index[len(index)-1]
		indexParam = parseYesNoOnly(index)
		if indexParam == Invalid {
			return nil, fmt.Errorf("invalid index:%q (valid values are: yes, only, no)", index)
		}
	}

	// If Zoekt is disabled just fallback to Unindexed.
	if !args.Zoekt.Enabled() {
		if indexParam == Only {
			return nil, fmt.Errorf("invalid index:%q (indexed search is not enabled)", indexParam)
		}

		return &indexedSearchRequest{
			Unindexed:        args.Repos,
			IndexUnavailable: true,
		}, nil
	}

	// Fallback to Unindexed if index:no
	if indexParam == No {
		return &indexedSearchRequest{
			Unindexed: args.Repos,
		}, nil
	}

	// Only include indexes with symbol information if a symbol request.
	var filter func(repo *zoekt.Repository) bool
	if typ == symbolRequest {
		filter = func(repo *zoekt.Repository) bool {
			return repo.HasSymbols
		}
	}

	// Consult Zoekt to find out which repository revisions can be searched.
	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()
	indexedSet, err := args.Zoekt.ListAll(ctx)
	if err != nil {
		if ctx.Err() == nil {
			// Only hard fail if the user specified index:only
			if indexParam == Only {
				return nil, errors.New("index:only failed since indexed search is not available yet")
			}

			log15.Warn("zoektIndexedRepos failed", "error", err)
		}

		return &indexedSearchRequest{
			Unindexed:        args.Repos,
			IndexUnavailable: true,
		}, ctx.Err()
	}

	// Split based on indexed vs unindexed
	indexed, searcherRepos := zoektIndexedRepos(indexedSet, args.Repos, filter)

	// We do not yet support searching non-HEAD for fileRequest (structural
	// search).
	if typ == fileRequest && indexed.NotHEADOnlySearch {
		return nil, errors.New("structural search only supports searching the default branch https://github.com/sourcegraph/sourcegraph/issues/11906")
	}

	return &indexedSearchRequest{
		args: args,
		typ:  typ,

		Unindexed: searcherRepos,
		repos:     indexed,

		DisableUnindexedSearch: indexParam == Only,
	}, nil
}

// Repos is a map of repository revisions that are indexed and will be
// searched by Zoekt. Do not mutate.
func (s *indexedSearchRequest) Repos() map[string]*search.RepositoryRevisions {
	if s.repos == nil {
		return nil
	}
	return s.repos.repoRevs
}

func (s *indexedSearchRequest) Search(ctx context.Context) (fm []*FileMatchResolver, limitHit bool, reposLimitHit map[string]struct{}, err error) {
	if len(s.Repos()) == 0 {
		return nil, false, nil, nil
	}

	since := time.Since
	if s.since != nil {
		since = s.since
	}

	switch s.typ {
	case textRequest:
		return zoektSearch(ctx, s.args, s.repos, s.typ, since)
	case symbolRequest:
		return zoektSearch(ctx, s.args, s.repos, s.typ, since)
	case fileRequest:
		return zoektSearchHEADOnlyFiles(ctx, s.args, s.repos.repoRevs, false, since)
	default:
		return nil, false, nil, fmt.Errorf("unexpected indexedSearchRequest type: %q", s.typ)
	}
}

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

// zoektSearch searches repositories using zoekt.
//
// Timeouts are reported through the context, and as a special case errNoResultsInTimeout
// is returned if no results are found in the given timeout (instead of the more common
// case of finding partial or full results in the given timeout).
func zoektSearch(ctx context.Context, args *search.TextParameters, repos *indexedRepoRevs, typ indexedRequestType, since func(t time.Time) time.Duration) (fm []*FileMatchResolver, limitHit bool, reposLimitHit map[string]struct{}, err error) {
	if len(repos.repoRevs) == 0 {
		return nil, false, nil, nil
	}

	queryExceptRepos, err := queryToZoektQuery(args.PatternInfo, typ)
	if err != nil {
		return nil, false, nil, err
	}
	finalQuery := zoektquery.NewAnd(&zoektquery.RepoBranches{Set: repos.repoBranches}, queryExceptRepos)

	tr, ctx := trace.New(ctx, "zoekt.Search", finalQuery.String())
	defer func() {
		tr.SetError(err)
		if len(fm) > 0 {
			tr.LazyPrintf("%d file matches", len(fm))
		}
		tr.Finish()
	}()

	k := zoektResultCountFactor(len(repos.repoBranches), args.PatternInfo)
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

	matches := make([]*FileMatchResolver, 0, len(resp.Files))
	repoResolvers := make(RepositoryResolverCache)
	for _, file := range resp.Files {
		fileLimitHit := false
		if len(file.LineMatches) > maxLineMatches {
			file.LineMatches = file.LineMatches[:maxLineMatches]
			fileLimitHit = true
			limitHit = true
		}

		repo, inputRevs := repos.GetRepoInputRev(&file)
		repoResolver := repoResolvers[repo.Name]
		if repoResolver == nil {
			repoResolver = &RepositoryResolver{repo: repo}
			repoResolvers[repo.Name] = repoResolver
		}

		var lines []*lineMatch
		var matchCount int
		if typ != symbolRequest {
			lines, matchCount = zoektFileMatchToLineMatches(maxLineFragmentMatches, &file)
		}

		for _, inputRev := range inputRevs {
			inputRev := inputRev // copy so we can take the pointer

			var symbols []*searchSymbolResult
			if typ == symbolRequest {
				symbols = zoektFileMatchToSymbolResults(repoResolver, inputRev, &file)
			}

			matches = append(matches, &FileMatchResolver{
				JPath:        file.FileName,
				JLineMatches: lines,
				JLimitHit:    fileLimitHit,
				MatchCount:   matchCount, // We do not use resp.MatchCount because it counts the number of lines matched, not the number of fragments.
				uri:          fileMatchURI(repo.Name, inputRev, file.FileName),
				symbols:      symbols,
				Repo:         repoResolver,
				CommitID:     api.CommitID(file.Version),
				InputRev:     &inputRev,
			})
		}
	}

	return matches, limitHit, reposLimitHit, nil
}

func zoektFileMatchToLineMatches(maxLineFragmentMatches int, file *zoekt.FileMatch) ([]*lineMatch, int) {
	var matchCount int
	lines := make([]*lineMatch, 0, len(file.LineMatches))

	for _, l := range file.LineMatches {
		if l.FileName {
			continue
		}

		if len(l.LineFragments) > maxLineFragmentMatches {
			l.LineFragments = l.LineFragments[:maxLineFragmentMatches]
		}
		offsets := make([][2]int32, len(l.LineFragments))
		for k, m := range l.LineFragments {
			offset := utf8.RuneCount(l.Line[:m.LineOffset])
			length := utf8.RuneCount(l.Line[m.LineOffset : m.LineOffset+m.MatchLength])
			offsets[k] = [2]int32{int32(offset), int32(length)}
		}
		matchCount += len(offsets)
		lines = append(lines, &lineMatch{
			JPreview:          string(l.Line),
			JLineNumber:       int32(l.LineNumber - 1),
			JOffsetAndLengths: offsets,
		})
	}

	return lines, matchCount
}

func zoektFileMatchToSymbolResults(repo *RepositoryResolver, inputRev string, file *zoekt.FileMatch) []*searchSymbolResult {
	// Symbol search returns a resolver so we need to pass in some
	// extra stuff. This is a sign that we can probably restructure
	// resolvers to avoid this.
	baseURI := &gituri.URI{URL: url.URL{Scheme: "git", Host: repo.Name(), RawQuery: url.QueryEscape(inputRev)}}
	commit := &GitCommitResolver{
		repoResolver: repo,
		oid:          GitObjectID(file.Version),
		inputRev:     &inputRev,
	}
	lang := strings.ToLower(file.Language)

	symbols := make([]*searchSymbolResult, 0, len(file.LineMatches))
	for _, l := range file.LineMatches {
		if l.FileName {
			continue
		}

		for _, m := range l.LineFragments {
			if m.SymbolInfo == nil {
				continue
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
				lang:    lang,
				baseURI: baseURI,
				commit:  commit,
			})
		}
	}

	return symbols
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

func queryToZoektQuery(query *search.TextPatternInfo, typ indexedRequestType) (zoektquery.Q, error) {
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

	if query.IsNegated {
		q = &zoektquery.Not{Child: q}
	}

	if typ == symbolRequest {
		// Tell zoekt q must match on symbols
		q = &zoektquery.Symbol{
			Expr: q,
		}
	}

	and = append(and, q)

	// zoekt also uses regular expressions for file paths
	// TODO PathPatternsAreCaseSensitive
	// TODO whitespace in file path patterns?
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

	// For conditionals that happen on a repo we can use type:repo queries. eg
	// (type:repo file:foo) (type:repo file:bar) will match all repos which
	// contain a filename matching "foo" and a filename matchinb "bar".
	//
	// Note: (type:repo file:foo file:bar) will only find repos with a
	// filename containing both "foo" and "bar".
	for _, p := range query.FilePatternsReposMustInclude {
		q, err := fileRe(p, query.IsCaseSensitive)
		if err != nil {
			return nil, err
		}
		and = append(and, &zoektquery.Type{Type: zoektquery.TypeRepo, Child: q})
	}
	for _, p := range query.FilePatternsReposMustExclude {
		q, err := fileRe(p, query.IsCaseSensitive)
		if err != nil {
			return nil, err
		}
		and = append(and, &zoektquery.Not{Child: &zoektquery.Type{Type: zoektquery.TypeRepo, Child: q}})
	}

	return zoektquery.Simplify(zoektquery.NewAnd(and...)), nil
}

// zoektIndexedRepos splits the revs into two parts: (1) the repository
// revisions in indexedSet (indexed) and (2) the repositories that are
// unindexed.
func zoektIndexedRepos(indexedSet map[string]*zoekt.Repository, revs []*search.RepositoryRevisions, filter func(*zoekt.Repository) bool) (indexed *indexedRepoRevs, unindexed []*search.RepositoryRevisions) {
	// PERF: If len(revs) is large, we expect to be doing an indexed
	// search. So set indexed to the max size it can be to avoid growing.
	indexed = &indexedRepoRevs{
		repoRevs:     make(map[string]*search.RepositoryRevisions, len(revs)),
		repoBranches: make(map[string][]string, len(revs)),
	}
	unindexed = make([]*search.RepositoryRevisions, 0)

	for _, reporev := range revs {
		repo, ok := indexedSet[string(reporev.Repo.Name)]
		if !ok || (filter != nil && !filter(repo)) {
			unindexed = append(unindexed, reporev)
			continue
		}

		unindexedRevs := indexed.Add(reporev, repo)
		if len(unindexedRevs) > 0 {
			copy := *reporev
			copy.Revs = unindexedRevs
			unindexed = append(unindexed, &copy)
		}
	}

	return indexed, unindexed
}

// indexedRepoRevs creates both the Sourcegraph and Zoekt representation of a
// list of repository and refs to search.
type indexedRepoRevs struct {
	// repoRevs is the Sourcegraph representation of a the list of repoRevs
	// repository and revisions to search.
	repoRevs map[string]*search.RepositoryRevisions

	// repoBranches will be used when we query zoekt. The order of branches
	// must match that in a reporev such that we can map back results. IE this
	// invariant is maintained:
	//
	//  repoBranches[reporev.Repo.Name][i] <-> reporev.Revs[i]
	repoBranches map[string][]string

	// NotHEADOnlySearch is true if we are searching a branch other than HEAD.
	//
	// This option can be removed once structural search supports searching
	// more than HEAD.
	NotHEADOnlySearch bool
}

// Add will add reporev and repo to the list of repository and branches to
// search if reporev's refs are a subset of repo's branches. It will return
// the revision specifiers it can't add.
func (rb *indexedRepoRevs) Add(reporev *search.RepositoryRevisions, repo *zoekt.Repository) []search.RevisionSpecifier {
	// A repo should only appear once in revs. However, in case this
	// invariant is broken we will treat later revs as if it isn't
	// indexed.
	if _, ok := rb.repoBranches[string(reporev.Repo.Name)]; ok {
		return reporev.Revs
	}

	if !reporev.OnlyExplicit() {
		// Contains a RefGlob or ExcludeRefGlob so we can't do indexed
		// search on it.
		//
		// TODO we could only process the explicit revs and return the non
		// explicit ones as unindexed.
		return reporev.Revs
	}

	// notHEADOnlySearch is set to true if we search any branch other than
	// repo.Branches[0]
	notHEADOnlySearch := false

	// Assume for large searches they will mostly involve indexed
	// revisions, so just allocate that.
	var unindexed []search.RevisionSpecifier
	indexed := make([]search.RevisionSpecifier, 0, len(reporev.Revs))

	branches := make([]string, 0, len(reporev.Revs))
	for _, rev := range reporev.Revs {
		if rev.RevSpec == "" || rev.RevSpec == "HEAD" {
			// Zoekt convention that first branch is HEAD
			branches = append(branches, repo.Branches[0].Name)
			indexed = append(indexed, rev)
			continue
		}

		found := false
		for i, branch := range repo.Branches {
			if branch.Name == rev.RevSpec {
				branches = append(branches, branch.Name)
				notHEADOnlySearch = notHEADOnlySearch || i > 0
				found = true
				break
			}
			// Check if rev is an abbrev commit SHA
			if len(rev.RevSpec) >= 4 && strings.HasPrefix(branch.Version, rev.RevSpec) {
				branches = append(branches, branch.Name)
				notHEADOnlySearch = notHEADOnlySearch || i > 0
				found = true
				break
			}
		}

		if found {
			indexed = append(indexed, rev)
		} else {
			unindexed = append(unindexed, rev)
		}
	}

	// We found indexed branches! Track them.
	if len(indexed) > 0 {
		rb.repoRevs[string(reporev.Repo.Name)] = reporev
		rb.repoBranches[string(reporev.Repo.Name)] = branches
		rb.NotHEADOnlySearch = rb.NotHEADOnlySearch || notHEADOnlySearch
	}

	return unindexed
}

// GetRepoInputRev returns the repo and inputRev associated with file.
func (rb *indexedRepoRevs) GetRepoInputRev(file *zoekt.FileMatch) (repo *types.Repo, inputRevs []string) {
	repoRev := rb.repoRevs[file.Repository]

	inputRevs = make([]string, 0, len(file.Branches))
	for _, branch := range file.Branches {
		for i, b := range rb.repoBranches[file.Repository] {
			if branch == b {
				// RevSpec is guaranteed to be explicit via zoektIndexedRepos
				inputRevs = append(inputRevs, repoRev.Revs[i].RevSpec)
			}
		}
	}

	if len(inputRevs) == 0 {
		// Did not find a match. This is unexpected, but we can fallback to
		// file.Version to generate correct links.
		inputRevs = append(inputRevs, file.Version)
	}

	return repoRev.Repo, inputRevs
}
