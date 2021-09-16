package zoekt

import (
	"context"
	"strings"
	"time"
	"unicode/utf8"

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
	repoRevs map[string]*search.RepositoryRevisions

	// repoBranches will be used when we query zoekt. The order of branches
	// must match that in a reporev such that we can map back results. IE this
	// invariant is maintained:
	//
	//  repoBranches[reporev.Repo.Name][i] <-> reporev.Revs[i]
	repoBranches map[string][]string
}

// headBranch is used as a singleton of the indexedRepoRevs.repoBranches to save
// common-case allocations within indexedRepoRevs.Add.
var headBranch = []string{"HEAD"}

// add will add reporev and repo to the list of repository and branches to
// search if reporev's refs are a subset of repo's branches. It will return
// the revision specifiers it can't add.
func (rb *IndexedRepoRevs) add(reporev *search.RepositoryRevisions, repo *zoekt.MinimalRepoListEntry) []search.RevisionSpecifier {
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

	if len(reporev.Revs) == 1 && repo.Branches[0].Name == "HEAD" && (reporev.Revs[0].RevSpec == "" || reporev.Revs[0].RevSpec == "HEAD") {
		rb.repoRevs[string(reporev.Repo.Name)] = reporev
		rb.repoBranches[string(reporev.Repo.Name)] = headBranch
		return nil
	}

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
		rb.repoRevs[string(reporev.Repo.Name)] = reporev
		rb.repoBranches[string(reporev.Repo.Name)] = branches
	}

	return unindexed
}

// getRepoInputRev returns the repo and inputRev associated with file.
func (rb *IndexedRepoRevs) getRepoInputRev(file *zoekt.FileMatch) (repo types.RepoName, inputRevs []string) {
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

// IndexedSearchRequest exposes a method Search(...) to search over indexed
// repositories. Two kinds of indexed searches implement it:
// (1) IndexedUniverseSearchRequest that searches over the universe of indexed repositories.
// (2) IndexedSubsetSearchRequest that searches over an indexed subset of repos in the universe of indexed repositories.
type IndexedSearchRequest interface {
	Search(context.Context, streaming.Sender) error
	IndexedRepos() map[string]*search.RepositoryRevisions
	UnindexedRepos() []*search.RepositoryRevisions
}

func NewIndexedSearchRequest(ctx context.Context, args *search.TextParameters, typ search.IndexedRequestType, onMissing OnMissingRepoRevs) (IndexedSearchRequest, error) {
	if args.Mode == search.ZoektGlobalSearch {
		// performance: optimize global searches where Zoekt searches
		// all shards anyway.
		return NewIndexedUniverseSearchRequest(ctx, args, typ, args.RepoOptions, args.UserPrivateRepos)
	}
	return NewIndexedSubsetSearchRequest(ctx, args, typ, onMissing)
}

// IndexedUniverseSearchRequest represents a request to run a search over the universe of indexed repositories.
type IndexedUniverseSearchRequest struct {
	RepoOptions      search.RepoOptions
	UserPrivateRepos []types.RepoName
	Args             *search.ZoektParameters
}

func (s *IndexedUniverseSearchRequest) Search(ctx context.Context, c streaming.Sender) error {
	if s.Args == nil {
		return nil
	}

	q := zoektGlobalQuery(s.Args.Query, s.RepoOptions, s.UserPrivateRepos)
	return doZoektSearchGlobal(ctx, q, s.Args.Typ, s.Args.Zoekt, s.Args.FileMatchLimit, s.Args.Select, c)
}

// IndexedRepos for a request over the indexed universe cannot answer which
// repositories are searched. This return value is always empty.
func (s *IndexedUniverseSearchRequest) IndexedRepos() map[string]*search.RepositoryRevisions {
	return map[string]*search.RepositoryRevisions{}
}

// UnindexedRepos over the indexed universe implies that we do not search unindexed repositories.
func (s *IndexedUniverseSearchRequest) UnindexedRepos() []*search.RepositoryRevisions {
	return nil
}

func NewIndexedUniverseSearchRequest(ctx context.Context, args *search.TextParameters, typ search.IndexedRequestType, repoOptions search.RepoOptions, userPrivateRepos []types.RepoName) (_ *IndexedUniverseSearchRequest, err error) {
	tr, _ := trace.New(ctx, "NewIndexedUniverseSearchRequest", "text")
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	q, err := search.QueryToZoektQuery(args.PatternInfo, typ == search.SymbolRequest)
	if err != nil {
		return nil, err
	}

	return &IndexedUniverseSearchRequest{
		RepoOptions:      repoOptions,
		UserPrivateRepos: userPrivateRepos,
		Args: &search.ZoektParameters{
			Query:          q,
			Typ:            typ,
			FileMatchLimit: args.PatternInfo.FileMatchLimit,
			Select:         args.PatternInfo.Select,
			Zoekt:          args.Zoekt,
		},
	}, nil
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
func (s *IndexedSubsetSearchRequest) IndexedRepos() map[string]*search.RepositoryRevisions {
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

func NewIndexedSubsetSearchRequest(ctx context.Context, args *search.TextParameters, typ search.IndexedRequestType, onMissing OnMissingRepoRevs) (_ *IndexedSubsetSearchRequest, err error) {
	tr, ctx := trace.New(ctx, "NewIndexedSubsetSearchRequest", string(typ))
	tr.LogFields(trace.Stringer("global_search_mode", args.Mode))
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	// If Zoekt is disabled just fallback to Unindexed.
	if args.Zoekt == nil {
		if args.PatternInfo.Index == query.Only {
			return nil, errors.Errorf("invalid index:%q (indexed search is not enabled)", args.PatternInfo.Index)
		}

		return &IndexedSubsetSearchRequest{
			Unindexed:        limitUnindexedRepos(args.Repos, maxUnindexedRepoRevSearchesPerQuery, onMissing),
			IndexUnavailable: true,
		}, nil
	}

	// Fallback to Unindexed if the query contains ref-globs
	if query.ContainsRefGlobs(args.Query) {
		if args.PatternInfo.Index == query.Only {
			return nil, errors.Errorf("invalid index:%q (revsions with glob pattern cannot be resolved for indexed searches)", args.PatternInfo.Index)
		}
		return &IndexedSubsetSearchRequest{
			Unindexed: limitUnindexedRepos(args.Repos, maxUnindexedRepoRevSearchesPerQuery, onMissing),
		}, nil
	}

	// Fallback to Unindexed if index:no
	if args.PatternInfo.Index == query.No {
		return &IndexedSubsetSearchRequest{
			Unindexed: limitUnindexedRepos(args.Repos, maxUnindexedRepoRevSearchesPerQuery, onMissing),
		}, nil
	}

	// Only include indexes with symbol information if a symbol request.
	var filter func(repo *zoekt.MinimalRepoListEntry) bool
	if typ == search.SymbolRequest {
		filter = func(repo *zoekt.MinimalRepoListEntry) bool {
			return repo.HasSymbols
		}
	}

	// Consult Zoekt to find out which repository revisions can be searched.
	ctx, cancel := context.WithTimeout(ctx, time.Minute)
	defer cancel()
	list, err := args.Zoekt.List(ctx, &zoektquery.Const{Value: true}, &zoekt.ListOptions{Minimal: true})
	if err != nil {
		if ctx.Err() == nil {
			// Only hard fail if the user specified index:only
			if args.PatternInfo.Index == query.Only {
				return nil, errors.New("index:only failed since indexed search is not available yet")
			}

			log15.Warn("zoektIndexedRepos failed", "error", err)
		}

		return &IndexedSubsetSearchRequest{
			Unindexed:        limitUnindexedRepos(args.Repos, maxUnindexedRepoRevSearchesPerQuery, onMissing),
			IndexUnavailable: true,
		}, ctx.Err()
	}

	tr.LogFields(log.Int("all_indexed_set.size", len(list.Minimal)))

	// Split based on indexed vs unindexed
	indexed, searcherRepos := zoektIndexedRepos(list.Minimal, args.Repos, filter)

	tr.LogFields(
		log.Int("indexed.size", len(indexed.repoRevs)),
		log.Int("searcher_repos.size", len(searcherRepos)),
	)

	// Disable unindexed search
	if args.PatternInfo.Index == query.Only {
		searcherRepos = limitUnindexedRepos(searcherRepos, 0, onMissing)
	}

	q, err := search.QueryToZoektQuery(args.PatternInfo, typ == search.SymbolRequest)
	if err != nil {
		return nil, err
	}

	return &IndexedSubsetSearchRequest{
		Args: &search.ZoektParameters{
			Query:          q,
			Typ:            typ,
			FileMatchLimit: args.PatternInfo.FileMatchLimit,
			Select:         args.PatternInfo.Select,
			Zoekt:          args.Zoekt,
		},

		Unindexed: limitUnindexedRepos(searcherRepos, maxUnindexedRepoRevSearchesPerQuery, onMissing),
		RepoRevs:  indexed,

		DisableUnindexedSearch: args.PatternInfo.Index == query.Only,
	}, nil
}

// zoektGlobalQuery constructs a query that searches the entire universe of indexed repositories.
//
// We construct 2 Zoekt queries. One query for public repos and one query for
// private repos.
//
// We only have to search "HEAD", because global queries, per definition, don't
// have a repo: filter and consequently no rev: filter. This makes the code a bit
// simpler because we don't have to resolve revisions before sending off (global)
// requests to Zoekt.
func zoektGlobalQuery(q zoektquery.Q, repoOptions search.RepoOptions, userPrivateRepos []types.RepoName) zoektquery.Q {
	var qs []zoektquery.Q

	// Public or Any
	if repoOptions.Visibility == query.Public || repoOptions.Visibility == query.Any {
		rc := zoektquery.RcOnlyPublic
		apply := func(f zoektquery.RawConfig, b bool) {
			if !b {
				return
			}
			rc |= f
		}
		apply(zoektquery.RcOnlyArchived, repoOptions.OnlyArchived)
		apply(zoektquery.RcNoArchived, repoOptions.NoArchived)
		apply(zoektquery.RcOnlyForks, repoOptions.OnlyForks)
		apply(zoektquery.RcNoForks, repoOptions.NoForks)

		qs = append(qs, zoektquery.NewAnd(&zoektquery.Branch{Pattern: "HEAD", Exact: true}, rc, q))
	}

	// Private or Any
	if (repoOptions.Visibility == query.Private || repoOptions.Visibility == query.Any) && len(userPrivateRepos) > 0 {
		privateRepoSet := make(map[string][]string, len(userPrivateRepos))
		head := []string{"HEAD"}
		for _, r := range userPrivateRepos {
			privateRepoSet[string(r.Name)] = head
		}
		qs = append(qs, zoektquery.NewAnd(&zoektquery.RepoBranches{Set: privateRepoSet}, q))
	}

	return zoektquery.Simplify(zoektquery.NewOr(qs...))
}

func doZoektSearchGlobal(ctx context.Context, q zoektquery.Q, typ search.IndexedRequestType, client zoekt.Streamer, fileMatchLimit int32, selector filter.SelectPath, c streaming.Sender) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	k := ResultCountFactor(0, fileMatchLimit, true)
	searchOpts := SearchOpts(ctx, k, fileMatchLimit)

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

	// PERF: if we are going to be selecting to repo results only anyways, we can
	// just ask zoekt for only results of type repo.
	if selector.Root() == filter.Repository {
		repoList, err := client.List(ctx, q, nil)
		if err != nil {
			return err
		}

		matches := make([]result.Match, 0, len(repoList.Repos))
		for _, repo := range repoList.Repos {
			matches = append(matches, &result.RepoMatch{
				Name: api.RepoName(repo.Repository.Name),
				ID:   api.RepoID(repo.Repository.ID),
			})
		}

		c.Send(streaming.SearchEvent{
			Results: matches,
			Stats:   streaming.Stats{}, // TODO
		})
		return nil
	}

	return client.StreamSearch(ctx, q, &searchOpts, backend.ZoektStreamFunc(func(event *zoekt.SearchResult) {
		sendMatches(event, func(file *zoekt.FileMatch) (types.RepoName, []string) {
			repo := types.RepoName{
				ID:   api.RepoID(file.RepositoryID),
				Name: api.RepoName(file.Repository),
			}
			return repo, []string{""}
		}, typ, c)
	}))
}

// zoektSearch searches repositories using zoekt.
func zoektSearch(ctx context.Context, repos *IndexedRepoRevs, q zoektquery.Q, typ search.IndexedRequestType, client zoekt.Streamer, fileMatchLimit int32, selector filter.SelectPath, since func(t time.Time) time.Duration, c streaming.Sender) error {
	if len(repos.repoRevs) == 0 {
		return nil
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	finalQuery := zoektquery.NewAnd(&zoektquery.RepoBranches{Set: repos.repoBranches}, q)

	k := ResultCountFactor(len(repos.repoBranches), fileMatchLimit, false)
	searchOpts := SearchOpts(ctx, k, fileMatchLimit)

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

	// PERF: if we are going to be selecting to repo results only anyways, we can just ask
	// zoekt for only results of type repo.
	if selector.Root() == filter.Repository {
		return zoektSearchReposOnly(ctx, client, finalQuery, c, func() map[api.RepoID]*search.RepositoryRevisions {
			repoRevMap := make(map[api.RepoID]*search.RepositoryRevisions, len(repos.repoRevs))
			for _, r := range repos.repoRevs {
				repoRevMap[r.Repo.ID] = r
			}
			return repoRevMap
		})
	}

	foundResults := atomic.Bool{}
	err := client.StreamSearch(ctx, finalQuery, &searchOpts, backend.ZoektStreamFunc(func(event *zoekt.SearchResult) {
		foundResults.CAS(false, event.FileCount != 0 || event.MatchCount != 0)
		sendMatches(event, repos.getRepoInputRev, typ, c)
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

func sendMatches(event *zoekt.SearchResult, getRepoInputRev repoRevFunc, typ search.IndexedRequestType, c streaming.Sender) {
	files := event.Files
	limitHit := event.FilesSkipped+event.ShardsSkipped > 0

	if len(files) == 0 {
		c.Send(streaming.SearchEvent{
			Stats: streaming.Stats{IsLimitHit: limitHit},
		})
		return
	}

	matches := make([]result.Match, 0, len(files))
	for _, file := range files {
		repo, inputRevs := getRepoInputRev(&file)

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

// zoektSearchReposOnly is used when select:repo is set, in which case we can ask zoekt
// only for the repos that contain matches for the query. This is a performance optimization,
// and not required for proper function of select:repo.
func zoektSearchReposOnly(ctx context.Context, client zoekt.Streamer, query zoektquery.Q, c streaming.Sender, getRepoRevMap func() map[api.RepoID]*search.RepositoryRevisions) error {
	repoList, err := client.List(ctx, query, &zoekt.ListOptions{Minimal: true})
	if err != nil {
		return err
	}

	repoRevMap := getRepoRevMap()
	if repoRevMap == nil {
		return nil
	}

	matches := make([]result.Match, 0, len(repoList.Minimal))
	for id := range repoList.Minimal {
		rev, ok := repoRevMap[api.RepoID(id)]
		if !ok {
			continue
		}

		matches = append(matches, &result.RepoMatch{
			Name: rev.Repo.Name,
			ID:   rev.Repo.ID,
		})
	}

	c.Send(streaming.SearchEvent{
		Results: matches,
		Stats:   streaming.Stats{}, // TODO
	})
	return nil
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

func zoektFileMatchToSymbolResults(repoName types.RepoName, inputRev string, file *zoekt.FileMatch) []*result.SymbolMatch {
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
		repoRevs:     make(map[string]*search.RepositoryRevisions, len(revs)),
		repoBranches: make(map[string][]string, len(revs)),
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
