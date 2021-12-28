package zoekt

import (
	"context"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/RoaringBitmap/roaring"
	"github.com/cockroachdb/errors"
	"github.com/google/zoekt"
	zoektquery "github.com/google/zoekt/query"
	"github.com/inconshreveable/log15"
	"github.com/opentracing/opentracing-go/log"
	"go.uber.org/atomic"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/backend"
	"github.com/sourcegraph/sourcegraph/internal/search/filter"
	"github.com/sourcegraph/sourcegraph/internal/search/query"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

// indexedRepoRevs creates both the Sourcegraph and Zoekt representation of a
// list of repository and refs to search.
type IndexedRepoRevs struct {
	// repoRevs is the Sourcegraph representation of a the list of repoRevs
	// repository and revisions to search.
	repoRevs map[api.RepoID]*search.RepositoryRevisions

	// branchRepos is used to construct a zoektquery.BranchesRepos to efficiently
	// marshal and send to zoekt
	branchRepos map[string]*zoektquery.BranchRepos
}

// add will add reporev and repo to the list of repository and branches to
// search if reporev's refs are a subset of repo's branches. It will return
// the revision specifiers it can't add.
func (rb *IndexedRepoRevs) add(reporev *search.RepositoryRevisions, repo *zoekt.MinimalRepoListEntry) []search.RevisionSpecifier {
	// A repo should only appear once in revs. However, in case this
	// invariant is broken we will treat later revs as if it isn't
	// indexed.
	if _, ok := rb.repoRevs[reporev.Repo.ID]; ok {
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

	if len(reporev.Revs) == 1 && repo.Branches[0].Name == "HEAD" && (reporev.Revs[0].RevSpec == "" || reporev.Revs[0].RevSpec == "HEAD") {
		rb.repoRevs[reporev.Repo.ID] = reporev
		br, ok := rb.branchRepos["HEAD"]
		if !ok {
			br = &zoektquery.BranchRepos{Branch: "HEAD", Repos: roaring.New()}
			rb.branchRepos["HEAD"] = br
		}
		br.Repos.Add(uint32(reporev.Repo.ID))
		return nil
	}

	// Assume for large searches they will mostly involve indexed
	// revisions, so just allocate that.
	var unindexed []search.RevisionSpecifier

	branches := make([]string, 0, len(reporev.Revs))
	reporev = reporev.Copy()
	indexed := reporev.Revs[:0]

	for _, rev := range reporev.Revs {
		if rev.RevSpec == "" || rev.RevSpec == "HEAD" {
			// Zoekt convention that first branch is HEAD
			branches = append(branches, repo.Branches[0].Name)
			indexed = append(indexed, rev)
			continue
		}

		found := false
		for _, branch := range repo.Branches {
			if branch.Name == rev.RevSpec {
				branches = append(branches, branch.Name)
				found = true
				break
			}
			// Check if rev is an abbrev commit SHA
			if len(rev.RevSpec) >= 4 && strings.HasPrefix(branch.Version, rev.RevSpec) {
				branches = append(branches, branch.Name)
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
		reporev.Revs = indexed
		rb.repoRevs[reporev.Repo.ID] = reporev
		for _, branch := range branches {
			br, ok := rb.branchRepos[branch]
			if !ok {
				br = &zoektquery.BranchRepos{Branch: branch, Repos: roaring.New()}
				rb.branchRepos[branch] = br
			}
			br.Repos.Add(uint32(reporev.Repo.ID))
		}
	}

	return unindexed
}

// getRepoInputRev returns the repo and inputRev associated with file.
func (rb *IndexedRepoRevs) getRepoInputRev(file *zoekt.FileMatch) (repo types.MinimalRepo, inputRevs []string) {
	repoRev := rb.repoRevs[api.RepoID(file.RepositoryID)]

	// We search zoekt by repo ID. It is possible that the name has come out
	// of sync, so the above lookup will fail. We fallback to linking the rev
	// hash in that case. We intend to restucture this code to avoid this, but
	// this is the fix to avoid potential nil panics.
	if repoRev == nil {
		repo := types.MinimalRepo{
			ID:   api.RepoID(file.RepositoryID),
			Name: api.RepoName(file.Repository),
		}
		return repo, []string{file.Version}
	}

	// We inverse the logic in add to work out the revspec from the zoekt
	// branches.
	//
	// Note: RevSpec is guaranteed to be explicit via zoektIndexedRepos
	inputRevs = make([]string, 0, len(file.Branches))
	for _, rev := range repoRev.Revs {
		// We rely on the Sourcegraph implementation that the HEAD branch is
		// indexed as "HEAD" rather than resolving the symref.
		revBranchName := rev.RevSpec
		if revBranchName == "" {
			revBranchName = "HEAD" // empty string in Sourcegraph means HEAD
		}

		found := false
		for _, branch := range file.Branches {
			if branch == revBranchName {
				found = true
				break
			}
		}
		if found {
			inputRevs = append(inputRevs, rev.RevSpec)
			continue
		}

		// Check if rev is an abbrev commit SHA
		if len(rev.RevSpec) >= 4 && strings.HasPrefix(file.Version, rev.RevSpec) {
			inputRevs = append(inputRevs, rev.RevSpec)
			continue
		}
	}

	if len(inputRevs) == 0 {
		// Did not find a match. This is unexpected, but we can fallback to
		// file.Version to generate correct links.
		inputRevs = append(inputRevs, file.Version)
	}

	return repoRev.Repo, inputRevs
}

// IndexedSearchRequest exposes a method Search(...) to search over indexed
// repositories. Two kinds of indexed searches implement it:
// (1) IndexedUniverseSearchRequest that searches over the universe of indexed repositories.
// (2) IndexedSubsetSearchRequest that searches over an indexed subset of repos in the universe of indexed repositories.
type IndexedSearchRequest interface {
	Search(context.Context, streaming.Sender) error
	IndexedRepos() map[api.RepoID]*search.RepositoryRevisions
	UnindexedRepos() []*search.RepositoryRevisions
}

func fallbackIndexUnavailable(repos []*search.RepositoryRevisions, limit int, onMissing OnMissingRepoRevs) *IndexedSubsetSearchRequest {
	return &IndexedSubsetSearchRequest{
		Unindexed:        limitUnindexedRepos(repos, limit, onMissing),
		IndexUnavailable: true,
	}
}

func fallbackUnindexed(repos []*search.RepositoryRevisions, limit int, onMissing OnMissingRepoRevs) *IndexedSubsetSearchRequest {
	return &IndexedSubsetSearchRequest{
		Unindexed: limitUnindexedRepos(repos, limit, onMissing),
	}
}

func OnlyUnindexed(repos []*search.RepositoryRevisions, zoekt zoekt.Streamer, useIndex query.YesNoOnly, containsRefGlobs bool, onMissing OnMissingRepoRevs) (IndexedSearchRequest, bool, error) {
	// If Zoekt is disabled just fallback to Unindexed.
	if zoekt == nil {
		if useIndex == query.Only {
			return nil, false, errors.Errorf("invalid index:%q (indexed search is not enabled)", useIndex)
		}
		return fallbackIndexUnavailable(repos, maxUnindexedRepoRevSearchesPerQuery, onMissing), true, nil
	}
	// Fallback to Unindexed if the query contains valid ref-globs.
	if containsRefGlobs {
		return fallbackUnindexed(repos, maxUnindexedRepoRevSearchesPerQuery, onMissing), true, nil
	}
	// Fallback to Unindexed if index:no
	if useIndex == query.No {
		return fallbackUnindexed(repos, maxUnindexedRepoRevSearchesPerQuery, onMissing), true, nil
	}
	return nil, false, nil
}

func NewIndexedSearchRequest(ctx context.Context, args *search.TextParameters, globalSearch bool, typ search.IndexedRequestType, onMissing OnMissingRepoRevs) (IndexedSearchRequest, error) {
	request, ok, err := OnlyUnindexed(args.Repos, args.Zoekt, args.PatternInfo.Index, query.ContainsRefGlobs(args.Query), onMissing)
	if err != nil {
		return nil, err
	}
	if ok {
		return request, nil
	}
	q, err := search.QueryToZoektQuery(args.PatternInfo, &args.Features, typ)
	if err != nil {
		return nil, err
	}
	zoektArgs := &search.ZoektParameters{
		Query:          q,
		Typ:            typ,
		FileMatchLimit: args.PatternInfo.FileMatchLimit,
		Select:         args.PatternInfo.Select,
		Zoekt:          args.Zoekt,
	}

	if globalSearch {
		// performance: optimize global searches where Zoekt searches
		// all shards anyway.
		return newIndexedUniverseSearchRequest(ctx, zoektArgs, args.RepoOptions, args.UserPrivateRepos)
	}
	return NewIndexedSubsetSearchRequest(ctx, args.Repos, args.PatternInfo.Index, zoektArgs, onMissing)
}

// IndexedUniverseSearchRequest represents a request to run a search over the universe of indexed repositories.
type IndexedUniverseSearchRequest struct {
	Args *search.ZoektParameters
}

func (s *IndexedUniverseSearchRequest) Search(ctx context.Context, c streaming.Sender) error {
	if s.Args == nil {
		return nil
	}
	return DoZoektSearchGlobal(ctx, s.Args, c)
}

// IndexedRepos for a request over the indexed universe cannot answer which
// repositories are searched. This return value is always empty.
func (s *IndexedUniverseSearchRequest) IndexedRepos() map[api.RepoID]*search.RepositoryRevisions {
	return map[api.RepoID]*search.RepositoryRevisions{}
}

// UnindexedRepos over the indexed universe implies that we do not search unindexed repositories.
func (s *IndexedUniverseSearchRequest) UnindexedRepos() []*search.RepositoryRevisions {
	return nil
}

// newIndexedUniverseSearchRequest creates a search request for indexed search
// on all indexed repositories. Strongly avoid calling this constructor
// directly, and use NewIndexedSearchRequest instead, which will validate your
// inputs and figure out the kind of indexed search to run.
func newIndexedUniverseSearchRequest(ctx context.Context, zoektArgs *search.ZoektParameters, repoOptions search.RepoOptions, userPrivateRepos []types.MinimalRepo) (_ *IndexedUniverseSearchRequest, err error) {
	tr, _ := trace.New(ctx, "newIndexedUniverseSearchRequest", "text")
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	defaultScope, err := DefaultGlobalQueryScope(repoOptions)
	if err != nil {
		return nil, err
	}
	includePrivate := repoOptions.Visibility == query.Private || repoOptions.Visibility == query.Any
	zoektGlobalQuery := NewGlobalZoektQuery(zoektArgs.Query, defaultScope, includePrivate)
	zoektGlobalQuery.ApplyPrivateFilter(userPrivateRepos)
	zoektArgs.Query = zoektGlobalQuery.Generate()
	return &IndexedUniverseSearchRequest{Args: zoektArgs}, nil
}

// IndexedSubsetSearchRequest is responsible for:
// (1) partitioning repos into indexed and unindexed sets of repos to search.
//     These sets are a subset of the universe of repos.
// (2) providing a method Search(...) that runs Zoekt over the indexed set of
//     repositories.
type IndexedSubsetSearchRequest struct {
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

	// Inputs
	Args *search.ZoektParameters

	// RepoRevs is the repository revisions that are indexed and will be
	// searched.
	RepoRevs *IndexedRepoRevs

	// since if non-nil will be used instead of time.Since. For tests
	since func(time.Time) time.Duration
}

// IndxedRepos is a map of indexed repository revisions will be searched by
// Zoekt. Do not mutate.
func (s *IndexedSubsetSearchRequest) IndexedRepos() map[api.RepoID]*search.RepositoryRevisions {
	if s.RepoRevs == nil {
		return nil
	}
	return s.RepoRevs.repoRevs
}

// UnindexedRepos is a slice of unindexed repositories to search.
func (s *IndexedSubsetSearchRequest) UnindexedRepos() []*search.RepositoryRevisions {
	return s.Unindexed
}

// Search streams 0 or more events to c.
func (s *IndexedSubsetSearchRequest) Search(ctx context.Context, c streaming.Sender) error {
	if s.Args == nil {
		return nil
	}

	if len(s.IndexedRepos()) == 0 {
		return nil
	}

	since := time.Since
	if s.since != nil {
		since = s.since
	}

	return zoektSearch(ctx, s.RepoRevs, s.Args.Query, s.Args.Typ, s.Args.Zoekt, s.Args.FileMatchLimit, s.Args.Select, since, c)
}

const maxUnindexedRepoRevSearchesPerQuery = 200

type OnMissingRepoRevs func([]*search.RepositoryRevisions)

func MissingRepoRevStatus(stream streaming.Sender) OnMissingRepoRevs {
	if stream == nil {
		return func([]*search.RepositoryRevisions) {}
	}
	return func(repoRevs []*search.RepositoryRevisions) {
		var status search.RepoStatusMap
		for _, r := range repoRevs {
			status.Update(r.Repo.ID, search.RepoStatusMissing)
		}
		stream.Send(streaming.SearchEvent{
			Stats: streaming.Stats{
				Status: status,
			},
		})
	}
}

// NewIndexedSubsetSearchRequest creates a search request for indexed search on
// a subset of repos. Strongly avoid calling this constructor directly, and use
// NewIndexedSearchRequest instead, which will validate your inputs and figure
// out the kind of indexed search to run.
func NewIndexedSubsetSearchRequest(ctx context.Context, repos []*search.RepositoryRevisions, index query.YesNoOnly, zoektArgs *search.ZoektParameters, onMissing OnMissingRepoRevs) (_ *IndexedSubsetSearchRequest, err error) {
	tr, ctx := trace.New(ctx, "newIndexedSubsetSearchRequest", string(zoektArgs.Typ))
	// Only include indexes with symbol information if a symbol request.
	var filter func(repo *zoekt.MinimalRepoListEntry) bool
	if zoektArgs.Typ == search.SymbolRequest {
		filter = func(repo *zoekt.MinimalRepoListEntry) bool {
			return repo.HasSymbols
		}
	}

	// Consult Zoekt to find out which repository revisions can be searched.
	ctx, cancel := context.WithTimeout(ctx, time.Minute)
	defer cancel()
	list, err := zoektArgs.Zoekt.List(ctx, &zoektquery.Const{Value: true}, &zoekt.ListOptions{Minimal: true})
	if err != nil {
		if ctx.Err() == nil {
			// Only hard fail if the user specified index:only
			if index == query.Only {
				return nil, errors.New("index:only failed since indexed search is not available yet")
			}

			log15.Warn("zoektIndexedRepos failed", "error", err)
		}

		return fallbackIndexUnavailable(repos, maxUnindexedRepoRevSearchesPerQuery, onMissing), ctx.Err()
	}

	tr.LogFields(log.Int("all_indexed_set.size", len(list.Minimal)))

	// Split based on indexed vs unindexed
	indexed, searcherRepos := zoektIndexedRepos(list.Minimal, repos, filter)

	tr.LogFields(
		log.Int("indexed.size", len(indexed.repoRevs)),
		log.Int("searcher_repos.size", len(searcherRepos)),
	)

	// Disable unindexed search
	if index == query.Only {
		searcherRepos = limitUnindexedRepos(searcherRepos, 0, onMissing)
	}

	return &IndexedSubsetSearchRequest{
		Args:      zoektArgs,
		Unindexed: limitUnindexedRepos(searcherRepos, maxUnindexedRepoRevSearchesPerQuery, onMissing),
		RepoRevs:  indexed,

		DisableUnindexedSearch: index == query.Only,
	}, nil
}

func DoZoektSearchGlobal(ctx context.Context, args *search.ZoektParameters, c streaming.Sender) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	k := ResultCountFactor(0, args.FileMatchLimit, true)
	searchOpts := SearchOpts(ctx, k, args.FileMatchLimit, args.Select)

	if deadline, ok := ctx.Deadline(); ok {
		// If the user manually specified a timeout, allow zoekt to use all of the remaining timeout.
		searchOpts.MaxWallTime = time.Until(deadline)
		if searchOpts.MaxWallTime < 0 {
			return ctx.Err()
		}
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

	return args.Zoekt.StreamSearch(ctx, args.Query, &searchOpts, backend.ZoektStreamFunc(func(event *zoekt.SearchResult) {
		sendMatches(event, func(file *zoekt.FileMatch) (types.MinimalRepo, []string) {
			repo := types.MinimalRepo{
				ID:   api.RepoID(file.RepositoryID),
				Name: api.RepoName(file.Repository),
			}
			return repo, []string{""}
		}, args.Typ, args.Select, c)
	}))
}

// zoektSearch searches repositories using zoekt.
func zoektSearch(ctx context.Context, repos *IndexedRepoRevs, q zoektquery.Q, typ search.IndexedRequestType, client zoekt.Streamer, fileMatchLimit int32, selector filter.SelectPath, since func(t time.Time) time.Duration, c streaming.Sender) error {
	if len(repos.repoRevs) == 0 {
		return nil
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	brs := make([]zoektquery.BranchRepos, 0, len(repos.branchRepos))
	for _, br := range repos.branchRepos {
		brs = append(brs, *br)
	}

	finalQuery := zoektquery.NewAnd(&zoektquery.BranchesRepos{List: brs}, q)

	k := ResultCountFactor(len(repos.repoRevs), fileMatchLimit, false)
	searchOpts := SearchOpts(ctx, k, fileMatchLimit, selector)

	// Start event stream.
	t0 := time.Now()

	if deadline, ok := ctx.Deadline(); ok {
		// If the user manually specified a timeout, allow zoekt to use all of the remaining timeout.
		searchOpts.MaxWallTime = time.Until(deadline)
		if searchOpts.MaxWallTime < 0 {
			return ctx.Err()
		}
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

	foundResults := atomic.Bool{}
	err := client.StreamSearch(ctx, finalQuery, &searchOpts, backend.ZoektStreamFunc(func(event *zoekt.SearchResult) {
		foundResults.CAS(false, event.FileCount != 0 || event.MatchCount != 0)
		sendMatches(event, repos.getRepoInputRev, typ, selector, c)
	}))
	if err != nil {
		return err
	}

	mkStatusMap := func(mask search.RepoStatus) search.RepoStatusMap {
		var statusMap search.RepoStatusMap
		for _, r := range repos.repoRevs {
			statusMap.Update(r.Repo.ID, mask)
		}
		return statusMap
	}

	if !foundResults.Load() && since(t0) >= searchOpts.MaxWallTime {
		c.Send(streaming.SearchEvent{Stats: streaming.Stats{Status: mkStatusMap(search.RepoStatusTimedout)}})
	}
	return nil
}

func sendMatches(event *zoekt.SearchResult, getRepoInputRev repoRevFunc, typ search.IndexedRequestType, selector filter.SelectPath, c streaming.Sender) {
	files := event.Files
	limitHit := event.FilesSkipped+event.ShardsSkipped > 0

	if selector.Root() == filter.Repository {
		// By default we stream up to "all" repository results per
		// select:repo request, and we never communicate whether a limit
		// is reached here based on Zoekt progress (because Zoekt can't
		// tell us the value of something like `ReposSkipped`). Instead,
		// limitHit is determined by other factors, like whether the
		// request is cancelled, or when we find the maximum number of
		// `count` results. I.e., from the webapp, this is
		// `max(defaultMaxSearchResultsStreaming,count)` which comes to
		// `max(500,count)`.
		limitHit = false
	}

	if len(files) == 0 {
		c.Send(streaming.SearchEvent{
			Stats: streaming.Stats{IsLimitHit: limitHit},
		})
		return
	}

	matches := make([]result.Match, 0, len(files))
	for _, file := range files {
		repo, inputRevs := getRepoInputRev(&file)

		if selector.Root() == filter.Repository {
			matches = append(matches, &result.RepoMatch{
				Name: repo.Name,
				ID:   repo.ID,
			})
			continue
		}

		var lines []*result.LineMatch
		if typ != search.SymbolRequest {
			lines = zoektFileMatchToLineMatches(&file)
		}

		for _, inputRev := range inputRevs {
			inputRev := inputRev // copy so we can take the pointer

			var symbols []*result.SymbolMatch
			if typ == search.SymbolRequest {
				symbols = zoektFileMatchToSymbolResults(repo, inputRev, &file)
			}
			fm := result.FileMatch{
				LineMatches: lines,
				Symbols:     symbols,
				File: result.File{
					InputRev: &inputRev,
					CommitID: api.CommitID(file.Version),
					Repo:     repo,
					Path:     file.FileName,
				},
			}
			matches = append(matches, &fm)
		}
	}

	c.Send(streaming.SearchEvent{
		Results: matches,
		Stats: streaming.Stats{
			IsLimitHit: limitHit,
		},
	})
}

func zoektFileMatchToLineMatches(file *zoekt.FileMatch) []*result.LineMatch {
	lines := make([]*result.LineMatch, 0, len(file.LineMatches))

	for _, l := range file.LineMatches {
		if l.FileName {
			continue
		}

		offsets := make([][2]int32, len(l.LineFragments))
		for k, m := range l.LineFragments {
			offset := utf8.RuneCount(l.Line[:m.LineOffset])
			length := utf8.RuneCount(l.Line[m.LineOffset : m.LineOffset+m.MatchLength])
			offsets[k] = [2]int32{int32(offset), int32(length)}
		}
		lines = append(lines, &result.LineMatch{
			Preview:          string(l.Line),
			LineNumber:       int32(l.LineNumber - 1),
			OffsetAndLengths: offsets,
		})
	}

	return lines
}

func zoektFileMatchToSymbolResults(repoName types.MinimalRepo, inputRev string, file *zoekt.FileMatch) []*result.SymbolMatch {
	newFile := &result.File{
		Path:     file.FileName,
		Repo:     repoName,
		CommitID: api.CommitID(file.Version),
		InputRev: &inputRev,
	}

	symbols := make([]*result.SymbolMatch, 0, len(file.LineMatches))
	for _, l := range file.LineMatches {
		if l.FileName {
			continue
		}

		for _, m := range l.LineFragments {
			if m.SymbolInfo == nil {
				continue
			}

			symbols = append(symbols, result.NewSymbolMatch(
				newFile,
				l.LineNumber,
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

	return symbols
}

// contextWithoutDeadline returns a context which will cancel if the cOld is
// canceled.
func contextWithoutDeadline(cOld context.Context) (context.Context, context.CancelFunc) {
	cNew, cancel := context.WithCancel(context.Background())

	// Set trace context so we still get spans propagated
	cNew = trace.CopyContext(cNew, cOld)

	// Copy actor from cOld to cNew.
	cNew = actor.WithActor(cNew, actor.FromContext(cOld))

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

// zoektIndexedRepos splits the revs into two parts: (1) the repository
// revisions in indexedSet (indexed) and (2) the repositories that are
// unindexed.
func zoektIndexedRepos(indexedSet map[uint32]*zoekt.MinimalRepoListEntry, revs []*search.RepositoryRevisions, filter func(repo *zoekt.MinimalRepoListEntry) bool) (indexed *IndexedRepoRevs, unindexed []*search.RepositoryRevisions) {
	// PERF: If len(revs) is large, we expect to be doing an indexed
	// search. So set indexed to the max size it can be to avoid growing.
	indexed = &IndexedRepoRevs{
		repoRevs:    make(map[api.RepoID]*search.RepositoryRevisions, len(revs)),
		branchRepos: make(map[string]*zoektquery.BranchRepos, 1),
	}
	unindexed = make([]*search.RepositoryRevisions, 0)

	for _, reporev := range revs {
		repo, ok := indexedSet[uint32(reporev.Repo.ID)]
		if !ok || (filter != nil && !filter(repo)) {
			unindexed = append(unindexed, reporev)
			continue
		}

		unindexedRevs := indexed.add(reporev, repo)
		if len(unindexedRevs) > 0 {
			copy := reporev.Copy()
			copy.Revs = unindexedRevs
			unindexed = append(unindexed, copy)
		}
	}

	return indexed, unindexed
}

// limitUnindexedRepos limits the number of repo@revs searched by the
// unindexed searcher codepath.  Sending many requests to searcher would
// otherwise cause a flood of system and network requests that result in
// timeouts or long delays.
//
// It returns the new repositories destined for the unindexed searcher code
// path, and sends an event to stream for any repositories that are limited /
// excluded.
//
// A slice to the input list is returned, it is not copied.
func limitUnindexedRepos(unindexed []*search.RepositoryRevisions, limit int, onMissing OnMissingRepoRevs) []*search.RepositoryRevisions {
	var missing []*search.RepositoryRevisions

	for i, repoRevs := range unindexed {
		limit -= len(repoRevs.Revs)
		if limit < 0 {
			missing = unindexed[i:]
			unindexed = unindexed[:i]
			break
		}
	}

	if len(missing) > 0 {
		onMissing(missing)
	}

	return unindexed
}
