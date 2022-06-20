package zoekt

import (
	"context"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/RoaringBitmap/roaring"
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
	"github.com/sourcegraph/sourcegraph/internal/search/job"
	"github.com/sourcegraph/sourcegraph/internal/search/query"
	searchrepos "github.com/sourcegraph/sourcegraph/internal/search/repos"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// IndexedRepoRevs creates both the Sourcegraph and Zoekt representation of a
// list of repository and refs to search.
type IndexedRepoRevs struct {
	// RepoRevs is the Sourcegraph representation of a the list of repoRevs
	// repository and revisions to search.
	RepoRevs map[api.RepoID]*search.RepositoryRevisions

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
	if _, ok := rb.RepoRevs[reporev.Repo.ID]; ok {
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

	// Assume for large searches they will mostly involve indexed
	// revisions, so just allocate that.
	var unindexed []search.RevisionSpecifier

	branches := make([]string, 0, len(reporev.Revs))
	reporev = reporev.Copy()
	indexed := reporev.Revs[:0]

	for _, rev := range reporev.Revs {
		found := false
		revSpec := rev.RevSpec
		if revSpec == "" {
			revSpec = "HEAD"
		}

		for _, branch := range repo.Branches {
			if branch.Name == revSpec {
				branches = append(branches, branch.Name)
				found = true
				break
			}
			// Check if rev is an abbrev commit SHA
			if len(revSpec) >= 4 && strings.HasPrefix(branch.Version, revSpec) {
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
		rb.RepoRevs[reporev.Repo.ID] = reporev
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
	repoRev := rb.RepoRevs[api.RepoID(file.RepositoryID)]

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

func PartitionRepos(
	ctx context.Context,
	repos []*search.RepositoryRevisions,
	zoektStreamer zoekt.Streamer,
	typ search.IndexedRequestType,
	useIndex query.YesNoOnly,
	containsRefGlobs bool,
) (indexed *IndexedRepoRevs, unindexed []*search.RepositoryRevisions, err error) {
	// Fallback to Unindexed if the query contains valid ref-globs.
	if containsRefGlobs {
		return nil, repos, nil
	}
	// Fallback to Unindexed if index:no
	if useIndex == query.No {
		return nil, repos, nil
	}

	tr, ctx := trace.New(ctx, "PartitionRepos", string(typ))
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

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
	list, err := zoektStreamer.List(ctx, &zoektquery.Const{Value: true}, &zoekt.ListOptions{Minimal: true})
	if err != nil {
		if ctx.Err() == nil {
			// Only hard fail if the user specified index:only
			if useIndex == query.Only {
				return nil, nil, errors.New("index:only failed since indexed search is not available yet")
			}

			log15.Warn("zoektIndexedRepos failed", "error", err)
		}

		return nil, repos, ctx.Err()
	}

	tr.LogFields(log.Int("all_indexed_set.size", len(list.Minimal)))

	// Split based on indexed vs unindexed
	indexed, unindexed = zoektIndexedRepos(list.Minimal, repos, filter)

	tr.LogFields(
		log.Int("indexed.size", len(indexed.RepoRevs)),
		log.Int("unindexed.size", len(unindexed)),
	)

	// Disable unindexed search
	if useIndex == query.Only {
		unindexed = unindexed[:0]
	}

	return indexed, unindexed, nil
}

func DoZoektSearchGlobal(ctx context.Context, client zoekt.Streamer, args *search.ZoektParameters, c streaming.Sender) error {
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

	return client.StreamSearch(ctx, args.Query, &searchOpts, backend.ZoektStreamFunc(func(event *zoekt.SearchResult) {
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
	if len(repos.RepoRevs) == 0 {
		return nil
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	brs := make([]zoektquery.BranchRepos, 0, len(repos.branchRepos))
	for _, br := range repos.branchRepos {
		brs = append(brs, *br)
	}

	finalQuery := zoektquery.NewAnd(&zoektquery.BranchesRepos{List: brs}, q)

	k := ResultCountFactor(len(repos.RepoRevs), fileMatchLimit, false)
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
		for _, r := range repos.RepoRevs {
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

		var hms result.ChunkMatches
		if typ != search.SymbolRequest {
			hms = zoektFileMatchToMultilineMatches(&file)
		}

		for _, inputRev := range inputRevs {
			inputRev := inputRev // copy so we can take the pointer

			var symbols []*result.SymbolMatch
			if typ == search.SymbolRequest {
				symbols = zoektFileMatchToSymbolResults(repo, inputRev, &file)
			}
			fm := result.FileMatch{
				ChunkMatches: hms,
				Symbols:      symbols,
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

func zoektFileMatchToMultilineMatches(file *zoekt.FileMatch) result.ChunkMatches {
	hms := make(result.ChunkMatches, 0, len(file.LineMatches))

	for _, l := range file.LineMatches {
		if l.FileName {
			continue
		}

		for _, m := range l.LineFragments {
			offset := utf8.RuneCount(l.Line[:m.LineOffset])
			length := utf8.RuneCount(l.Line[m.LineOffset : m.LineOffset+m.MatchLength])

			hms = append(hms, result.ChunkMatch{
				Content: string(l.Line),
				// zoekt line numbers are 1-based rather than 0-based so subtract 1
				ContentStart: result.Location{
					Offset: l.LineStart,
					Line:   l.LineNumber - 1,
					Column: 0,
				},
				Ranges: result.Ranges{{
					Start: result.Location{
						Offset: int(m.Offset),
						Line:   l.LineNumber - 1,
						Column: offset,
					},
					End: result.Location{
						Offset: int(m.Offset) + m.MatchLength,
						Line:   l.LineNumber - 1,
						Column: offset + length,
					},
				}},
			})
		}
	}

	return hms
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
				-1, // -1 means infer the column
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
		RepoRevs:    make(map[api.RepoID]*search.RepositoryRevisions, len(revs)),
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

type RepoSubsetTextSearchJob struct {
	Repos          *IndexedRepoRevs // the set of indexed repository revisions to search.
	Query          zoektquery.Q
	Typ            search.IndexedRequestType
	FileMatchLimit int32
	Select         filter.SelectPath
	Since          func(time.Time) time.Duration `json:"-"` // since if non-nil will be used instead of time.Since. For tests
}

// ZoektSearch is a job that searches repositories using zoekt.
func (z *RepoSubsetTextSearchJob) Run(ctx context.Context, clients job.RuntimeClients, stream streaming.Sender) (alert *search.Alert, err error) {
	_, ctx, stream, finish := job.StartSpan(ctx, stream, z)
	defer func() { finish(alert, err) }()

	if z.Repos == nil {
		return nil, nil
	}
	if len(z.Repos.RepoRevs) == 0 {
		return nil, nil
	}

	since := time.Since
	if z.Since != nil {
		since = z.Since
	}

	return nil, zoektSearch(ctx, z.Repos, z.Query, z.Typ, clients.Zoekt, z.FileMatchLimit, z.Select, since, stream)
}

func (*RepoSubsetTextSearchJob) Name() string {
	return "ZoektRepoSubsetTextSearchJob"
}

func (z *RepoSubsetTextSearchJob) Tags() []log.Field {
	tags := []log.Field{
		trace.Stringer("query", z.Query),
		log.String("type", string(z.Typ)),
		log.Int32("fileMatchLimit", z.FileMatchLimit),
		trace.Stringer("select", z.Select),
	}
	// z.Repos is nil for un-indexed search
	if z.Repos != nil {
		tags = append(tags, log.Int("numRepoRevs", len(z.Repos.RepoRevs)))
		tags = append(tags, log.Int("numBranchRepos", len(z.Repos.branchRepos)))
	}
	return tags
}

type GlobalTextSearchJob struct {
	GlobalZoektQuery *GlobalZoektQuery
	ZoektArgs        *search.ZoektParameters
	RepoOpts         search.RepoOptions
}

func (t *GlobalTextSearchJob) Run(ctx context.Context, clients job.RuntimeClients, stream streaming.Sender) (alert *search.Alert, err error) {
	_, ctx, stream, finish := job.StartSpan(ctx, stream, t)
	defer func() { finish(alert, err) }()

	userPrivateRepos := searchrepos.PrivateReposForActor(ctx, clients.DB, t.RepoOpts)
	t.GlobalZoektQuery.ApplyPrivateFilter(userPrivateRepos)
	t.ZoektArgs.Query = t.GlobalZoektQuery.Generate()

	return nil, DoZoektSearchGlobal(ctx, clients.Zoekt, t.ZoektArgs, stream)
}

func (*GlobalTextSearchJob) Name() string {
	return "ZoektGlobalTextSearchJob"
}

func (t *GlobalTextSearchJob) Tags() []log.Field {
	return []log.Field{
		trace.Stringer("query", t.GlobalZoektQuery.Query),
		trace.Printf("repoScope", "%q", t.GlobalZoektQuery.RepoScope),
		log.Bool("includePrivate", t.GlobalZoektQuery.IncludePrivate),
		log.String("type", string(t.ZoektArgs.Typ)),
		log.Int32("fileMatchLimit", t.ZoektArgs.FileMatchLimit),
		trace.Stringer("select", t.ZoektArgs.Select),
		trace.Stringer("repoOpts", &t.RepoOpts),
	}
}
