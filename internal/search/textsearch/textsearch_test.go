package textsearch

import (
	"context"
	"crypto/md5"
	"encoding/binary"
	"fmt"
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"
	"golang.org/x/sync/errgroup"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/endpoint"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/internal/search"
	searchbackend "github.com/sourcegraph/sourcegraph/internal/search/backend"
	"github.com/sourcegraph/sourcegraph/internal/search/job"
	"github.com/sourcegraph/sourcegraph/internal/search/limits"
	"github.com/sourcegraph/sourcegraph/internal/search/query"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
	"github.com/sourcegraph/sourcegraph/internal/search/searcher"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming"
	zoektutil "github.com/sourcegraph/sourcegraph/internal/search/zoekt"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestRepoSubsetTextSearch(t *testing.T) {
	searcher.MockSearchFilesInRepo = func(ctx context.Context, repo types.MinimalRepo, gitserverRepo api.RepoName, rev string, info *search.TextPatternInfo, fetchTimeout time.Duration, stream streaming.Sender) (limitHit bool, err error) {
		repoName := repo.Name
		switch repoName {
		case "foo/one":
			stream.Send(streaming.SearchEvent{
				Results: []result.Match{&result.FileMatch{
					File: result.File{
						Repo:     repo,
						InputRev: &rev,
						Path:     "main.go",
					},
				}},
			})
			return false, nil
		case "foo/two":
			stream.Send(streaming.SearchEvent{
				Results: []result.Match{&result.FileMatch{
					File: result.File{
						Repo:     repo,
						InputRev: &rev,
						Path:     "main.go",
					},
				}},
			})
			return false, nil
		case "foo/empty":
			return false, nil
		case "foo/cloning":
			return false, &gitdomain.RepoNotExistError{Repo: repoName, CloneInProgress: true}
		case "foo/missing":
			return false, &gitdomain.RepoNotExistError{Repo: repoName}
		case "foo/missing-database":
			return false, &errcode.Mock{Message: "repo not found: foo/missing-database", IsNotFound: true}
		case "foo/timedout":
			return false, context.DeadlineExceeded
		case "foo/no-rev":
			// TODO we do not specify a rev when searching "foo/no-rev", so it
			// is treated as an empty repository. We need to test the fatal
			// case of trying to search a revision which doesn't exist.
			return false, &gitdomain.RevisionNotFoundError{Repo: repoName, Spec: "missing"}
		default:
			return false, errors.New("Unexpected repo")
		}
	}
	defer func() { searcher.MockSearchFilesInRepo = nil }()

	zoekt := &searchbackend.FakeSearcher{}

	q, err := query.ParseLiteral("foo")
	if err != nil {
		t.Fatal(err)
	}
	repoRevs := makeRepositoryRevisions("foo/one", "foo/two", "foo/empty", "foo/cloning", "foo/missing", "foo/missing-database", "foo/timedout", "foo/no-rev")

	patternInfo := &search.TextPatternInfo{
		FileMatchLimit: limits.DefaultMaxSearchResults,
		Pattern:        "foo",
	}

	matches, common, err := RunRepoSubsetTextSearch(
		context.Background(),
		patternInfo,
		repoRevs,
		q,
		zoekt,
		endpoint.Static("test"),
		search.DefaultMode,
		false,
	)
	if err != nil {
		t.Fatal(err)
	}
	if len(matches) != 2 {
		t.Errorf("expected two results, got %d", len(matches))
	}
	repoNames := map[api.RepoID]string{}
	for _, rr := range repoRevs {
		repoNames[rr.Repo.ID] = string(rr.Repo.Name)
	}
	assertReposStatus(t, repoNames, common.Status, map[string]search.RepoStatus{
		"foo/cloning":          search.RepoStatusCloning,
		"foo/missing":          search.RepoStatusMissing,
		"foo/missing-database": search.RepoStatusMissing,
		"foo/timedout":         search.RepoStatusTimedout,
	})

	// If we specify a rev and it isn't found, we fail the whole search since
	// that should be checked earlier.
	_, _, err = RunRepoSubsetTextSearch(
		context.Background(),
		patternInfo,
		makeRepositoryRevisions("foo/no-rev@dev"),
		q,
		zoekt,
		endpoint.Static("test"),
		search.DefaultMode,
		false,
	)
	if !errors.HasType(err, &gitdomain.RevisionNotFoundError{}) {
		t.Fatalf("searching non-existent rev expected to fail with RevisionNotFoundError got: %v", err)
	}
}

func TestSearchFilesInReposStream(t *testing.T) {
	searcher.MockSearchFilesInRepo = func(ctx context.Context, repo types.MinimalRepo, gitserverRepo api.RepoName, rev string, info *search.TextPatternInfo, fetchTimeout time.Duration, stream streaming.Sender) (limitHit bool, err error) {
		repoName := repo.Name
		switch repoName {
		case "foo/one":
			stream.Send(streaming.SearchEvent{
				Results: []result.Match{&result.FileMatch{
					File: result.File{
						Repo:     repo,
						InputRev: &rev,
						Path:     "main.go",
					},
				}},
			})
			return false, nil
		case "foo/two":
			stream.Send(streaming.SearchEvent{
				Results: []result.Match{&result.FileMatch{
					File: result.File{
						Repo:     repo,
						InputRev: &rev,
						Path:     "main.go",
					},
				}},
			})
			return false, nil
		case "foo/three":
			stream.Send(streaming.SearchEvent{
				Results: []result.Match{&result.FileMatch{
					File: result.File{
						Repo:     repo,
						InputRev: &rev,
						Path:     "main.go",
					},
				}},
			})
			return false, nil
		default:
			return false, errors.New("Unexpected repo")
		}
	}
	defer func() { searcher.MockSearchFilesInRepo = nil }()

	zoekt := &searchbackend.FakeSearcher{}

	q, err := query.ParseLiteral("foo")
	if err != nil {
		t.Fatal(err)
	}

	patternInfo := &search.TextPatternInfo{
		FileMatchLimit: limits.DefaultMaxSearchResults,
		Pattern:        "foo",
	}

	matches, _, err := RunRepoSubsetTextSearch(
		context.Background(),
		patternInfo,
		makeRepositoryRevisions("foo/one", "foo/two", "foo/three"),
		q,
		zoekt,
		endpoint.Static("test"),
		search.DefaultMode,
		false,
	)
	if err != nil {
		t.Fatal(err)
	}

	if len(matches) != 3 {
		t.Errorf("expected three results, got %d", len(matches))
	}
}

func assertReposStatus(t *testing.T, repoNames map[api.RepoID]string, got search.RepoStatusMap, want map[string]search.RepoStatus) {
	t.Helper()
	gotM := map[string]search.RepoStatus{}
	got.Iterate(func(id api.RepoID, mask search.RepoStatus) {
		name := repoNames[id]
		if name == "" {
			name = fmt.Sprintf("UNKNOWNREPO{ID=%d}", id)
		}
		gotM[name] = mask
	})
	if diff := cmp.Diff(want, gotM); diff != "" {
		t.Errorf("RepoStatusMap mismatch (-want +got):\n%s", diff)
	}
}

func TestSearchFilesInRepos_multipleRevsPerRepo(t *testing.T) {
	searcher.MockSearchFilesInRepo = func(ctx context.Context, repo types.MinimalRepo, gitserverRepo api.RepoName, rev string, info *search.TextPatternInfo, fetchTimeout time.Duration, stream streaming.Sender) (limitHit bool, err error) {
		repoName := repo.Name
		switch repoName {
		case "foo":
			stream.Send(streaming.SearchEvent{
				Results: []result.Match{&result.FileMatch{
					File: result.File{
						Repo:     repo,
						CommitID: api.CommitID(rev),
						Path:     "main.go",
					},
				}},
			})
			return false, nil
		default:
			panic("unexpected repo")
		}
	}
	defer func() { searcher.MockSearchFilesInRepo = nil }()

	trueVal := true
	conf.Mock(&conf.Unified{SiteConfiguration: schema.SiteConfiguration{
		ExperimentalFeatures: &schema.ExperimentalFeatures{SearchMultipleRevisionsPerRepository: &trueVal},
	}})
	defer conf.Mock(nil)

	zoekt := &searchbackend.FakeSearcher{}

	q, err := query.ParseLiteral("foo")
	if err != nil {
		t.Fatal(err)
	}

	patternInfo := &search.TextPatternInfo{
		FileMatchLimit: limits.DefaultMaxSearchResults,
		Pattern:        "foo",
	}

	repos := makeRepositoryRevisions("foo@master:mybranch:*refs/heads/")
	repos[0].ListRefs = func(context.Context, database.DB, api.RepoName) ([]gitdomain.Ref, error) {
		return []gitdomain.Ref{{Name: "refs/heads/branch3"}, {Name: "refs/heads/branch4"}}, nil
	}

	matches, _, err := RunRepoSubsetTextSearch(
		context.Background(),
		patternInfo,
		repos,
		q,
		zoekt,
		endpoint.Static("test"),
		search.DefaultMode,
		false,
	)
	if err != nil {
		t.Fatal(err)
	}

	matchKeys := make([]result.Key, len(matches))
	for i, match := range matches {
		matchKeys[i] = match.Key()
	}
	sort.Slice(matchKeys, func(i, j int) bool { return matchKeys[i].Less(matchKeys[j]) })

	wantResultKeys := []result.Key{
		{Repo: "foo", Commit: "branch3", Path: "main.go"},
		{Repo: "foo", Commit: "branch4", Path: "main.go"},
		{Repo: "foo", Commit: "master", Path: "main.go"},
		{Repo: "foo", Commit: "mybranch", Path: "main.go"},
	}
	if !reflect.DeepEqual(matchKeys, wantResultKeys) {
		t.Errorf("got %v, want %v", matchKeys, wantResultKeys)
	}
}

func makeRepositoryRevisions(repos ...string) []*search.RepositoryRevisions {
	r := make([]*search.RepositoryRevisions, len(repos))
	for i, repospec := range repos {
		repoName, revs := search.ParseRepositoryRevisions(repospec)
		if len(revs) == 0 {
			// treat empty list as preferring master
			revs = []search.RevisionSpecifier{{RevSpec: ""}}
		}
		r[i] = &search.RepositoryRevisions{Repo: mkRepos(repoName)[0], Revs: revs}
	}
	return r
}

func mkRepos(names ...string) []types.MinimalRepo {
	var repos []types.MinimalRepo
	for _, name := range names {
		sum := md5.Sum([]byte(name))
		id := api.RepoID(binary.BigEndian.Uint64(sum[:]))
		if id < 0 {
			id = -(id / 2)
		}
		if id == 0 {
			id++
		}
		repos = append(repos, types.MinimalRepo{ID: id, Name: api.RepoName(name)})
	}
	return repos
}

func TestFileMatch_Limit(t *testing.T) {
	tests := []struct {
		numHunkRanges       int
		numSymbolMatches    int
		limit               int
		expNumHunkRanges    int
		expNumSymbolMatches int
		expRemainingLimit   int
		wantLimitHit        bool
	}{
		{
			numHunkRanges:     3,
			numSymbolMatches:  0,
			limit:             1,
			expNumHunkRanges:  1,
			expRemainingLimit: 0,
			wantLimitHit:      true,
		},
		{
			numHunkRanges:       0,
			numSymbolMatches:    3,
			limit:               1,
			expNumSymbolMatches: 1,
			expRemainingLimit:   0,
			wantLimitHit:        true,
		},
		{
			numHunkRanges:     3,
			numSymbolMatches:  0,
			limit:             5,
			expNumHunkRanges:  3,
			expRemainingLimit: 2,
			wantLimitHit:      false,
		},
		{
			numHunkRanges:       0,
			numSymbolMatches:    3,
			limit:               5,
			expNumSymbolMatches: 3,
			expRemainingLimit:   2,
			wantLimitHit:        false,
		},
		{
			numHunkRanges:     3,
			numSymbolMatches:  0,
			limit:             3,
			expNumHunkRanges:  3,
			expRemainingLimit: 0,
			wantLimitHit:      false,
		},
		{
			numHunkRanges:       0,
			numSymbolMatches:    3,
			limit:               3,
			expNumSymbolMatches: 3,
			expRemainingLimit:   0,
			wantLimitHit:        false,
		},
		{
			// An empty FileMatch should still count against the limit
			numHunkRanges:       0,
			numSymbolMatches:    0,
			limit:               1,
			expNumSymbolMatches: 0,
			expNumHunkRanges:    0,
			wantLimitHit:        false,
		},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			fileMatch := &result.FileMatch{
				File:         result.File{},
				ChunkMatches: result.ChunkMatches{{Ranges: make(result.Ranges, tt.numHunkRanges)}},
				Symbols:      make([]*result.SymbolMatch, tt.numSymbolMatches),
				LimitHit:     false,
			}

			got := fileMatch.Limit(tt.limit)

			require.Equal(t, tt.expNumHunkRanges, fileMatch.ChunkMatches.MatchCount())
			require.Equal(t, tt.expNumSymbolMatches, len(fileMatch.Symbols))
			require.Equal(t, tt.expRemainingLimit, got)
			require.Equal(t, tt.wantLimitHit, fileMatch.LimitHit)
		})
	}
}

// RunRepoSubsetTextSearch is a convenience function that simulates the RepoSubsetTextSearch job.
func RunRepoSubsetTextSearch(
	ctx context.Context,
	patternInfo *search.TextPatternInfo,
	repos []*search.RepositoryRevisions,
	q query.Q,
	zoekt *searchbackend.FakeSearcher,
	searcherURLs *endpoint.Map,
	mode search.GlobalSearchMode,
	useFullDeadline bool,
) ([]*result.FileMatch, streaming.Stats, error) {
	notSearcherOnly := mode != search.SearcherOnly
	searcherArgs := &search.SearcherParameters{
		PatternInfo:     patternInfo,
		UseFullDeadline: useFullDeadline,
	}

	agg := streaming.NewAggregatingStream()

	indexed, unindexed, err := zoektutil.PartitionRepos(
		context.Background(),
		repos,
		zoekt,
		search.TextRequest,
		query.Yes,
		query.ContainsRefGlobs(q),
	)
	if err != nil {
		return nil, streaming.Stats{}, err
	}

	g, ctx := errgroup.WithContext(ctx)

	if notSearcherOnly {
		b, err := query.ToBasicQuery(q)
		if err != nil {
			return nil, streaming.Stats{}, err
		}

		types, _ := q.StringValues(query.FieldType)
		var resultTypes result.Types
		if len(types) == 0 {
			resultTypes = result.TypeFile | result.TypePath | result.TypeRepo
		} else {
			for _, t := range types {
				resultTypes = resultTypes.With(result.TypeFromString[t])
			}
		}

		typ := search.TextRequest
		zoektQuery, err := zoektutil.QueryToZoektQuery(b, resultTypes, nil, typ)
		if err != nil {
			return nil, streaming.Stats{}, err
		}

		zoektJob := &zoektutil.RepoSubsetTextSearchJob{
			Repos:          indexed,
			Query:          zoektQuery,
			Typ:            search.TextRequest,
			FileMatchLimit: patternInfo.FileMatchLimit,
			Select:         patternInfo.Select,
			Since:          nil,
		}

		// Run literal and regexp searches on indexed repositories.
		g.Go(func() error {
			_, err := zoektJob.Run(ctx, job.RuntimeClients{Zoekt: zoekt}, agg)
			return err
		})
	}

	// Concurrently run searcher for all unindexed repos regardless whether text or regexp.
	g.Go(func() error {
		searcherJob := &searcher.TextSearchJob{
			PatternInfo:     searcherArgs.PatternInfo,
			Repos:           unindexed,
			Indexed:         false,
			UseFullDeadline: searcherArgs.UseFullDeadline,
		}

		_, err := searcherJob.Run(ctx, job.RuntimeClients{SearcherURLs: searcherURLs, Zoekt: zoekt}, agg)
		return err
	})

	err = g.Wait()

	fms, fmErr := matchesToFileMatches(agg.Results)
	if fmErr != nil && err == nil {
		err = errors.Wrap(fmErr, "searchFilesInReposBatch failed to convert results")
	}
	return fms, agg.Stats, err
}

func matchesToFileMatches(matches []result.Match) ([]*result.FileMatch, error) {
	fms := make([]*result.FileMatch, 0, len(matches))
	for _, match := range matches {
		fm, ok := match.(*result.FileMatch)
		if !ok {
			return nil, errors.Errorf("expected only file match results")
		}
		fms = append(fms, fm)
	}
	return fms, nil
}
