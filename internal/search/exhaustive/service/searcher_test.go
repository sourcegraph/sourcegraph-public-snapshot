package service

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/hexops/autogold/v2"
	"github.com/sourcegraph/log/logtest"
	"github.com/sourcegraph/zoekt"
	"github.com/stretchr/testify/require"
	"golang.org/x/exp/slices"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbmocks"
	"github.com/sourcegraph/sourcegraph/internal/endpoint"
	"github.com/sourcegraph/sourcegraph/internal/featureflag"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/internal/search"
	searchbackend "github.com/sourcegraph/sourcegraph/internal/search/backend"
	"github.com/sourcegraph/sourcegraph/internal/search/client"
	types2 "github.com/sourcegraph/sourcegraph/internal/search/exhaustive/types"
	"github.com/sourcegraph/sourcegraph/internal/search/job"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
	"github.com/sourcegraph/sourcegraph/internal/search/searcher"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/iterator"
)

func TestBackendFake(t *testing.T) {
	testNewSearcher(t, context.Background(), NewSearcherFake(), newSearcherTestCase{
		Query:        "1@rev1 1@rev2 2@rev3",
		WantRefSpecs: "RepositoryRevSpec{1@spec} RepositoryRevSpec{2@spec}",
		WantRepoRevs: "RepositoryRevision{1@rev1} RepositoryRevision{1@rev2} RepositoryRevision{2@rev3}",
		WantCSV: autogold.Expect(`repo,revspec,revision
1,spec,rev1
1,spec,rev2
2,spec,rev3
`),
	})
}

type newSearcherTestCase struct {
	Query        string
	WantRefSpecs string
	WantRepoRevs string
	WantCSV      autogold.Value
}

func TestFromSearchClient(t *testing.T) {
	repoMocks := []repoMock{{
		ID:   1,
		Name: "foo1",
		Branches: map[string]string{
			"HEAD": "commitfoo0",
			"dev1": "commitfoo1",
			"dev2": "commitfoo2",
		},
	}, {
		ID:   2,
		Name: "bar2",
		Branches: map[string]string{
			"HEAD": "commitbar0",
			"dev1": "commitbar1",
		},
	}, {
		ID:       3,
		Name:     "empty3",
		Branches: map[string]string{},
	}}

	ctx := featureflag.WithFlags(context.Background(), featureflag.NewMemoryStore(nil, nil, nil))
	mock := mockSearchClient(t, repoMocks)
	newSearcher := FromSearchClient(mock)

	do := func(name string, tc newSearcherTestCase) {
		t.Run(name, func(t *testing.T) {
			testNewSearcher(t, ctx, newSearcher, tc)
		})
	}

	// NOTE: our search stack calls gitserver twice per non-HEAD revision we
	// search. Converting a RefSpec into a RepoRev we validate the refspec
	// exists (or expand a glob). Then at actual search time we resolve it
	// again to find the actual commit to search.

	do("global", newSearcherTestCase{
		Query:        "content",
		WantRefSpecs: "RepositoryRevSpec{1@HEAD} RepositoryRevSpec{2@HEAD} RepositoryRevSpec{3@HEAD}",
		WantRepoRevs: "RepositoryRevision{1@HEAD} RepositoryRevision{2@HEAD} RepositoryRevision{3@HEAD}",
		WantCSV: autogold.Expect(`repository,revision,file_path,match_count,first_match_url
foo1,commitfoo0,,1,/foo1@commitfoo0/-/blob/?L2
bar2,commitbar0,,1,/bar2@commitbar0/-/blob/?L2
`),
	})

	do("explicit type:file", newSearcherTestCase{
		Query:        "content type:file",
		WantRefSpecs: "RepositoryRevSpec{1@HEAD} RepositoryRevSpec{2@HEAD} RepositoryRevSpec{3@HEAD}",
		WantRepoRevs: "RepositoryRevision{1@HEAD} RepositoryRevision{2@HEAD} RepositoryRevision{3@HEAD}",
		WantCSV: autogold.Expect(`repository,revision,file_path,match_count,first_match_url
foo1,commitfoo0,,1,/foo1@commitfoo0/-/blob/?L2
bar2,commitbar0,,1,/bar2@commitbar0/-/blob/?L2
`),
	})

	do("repo", newSearcherTestCase{
		Query:        "repo:foo content",
		WantRefSpecs: "RepositoryRevSpec{1@HEAD}",
		WantRepoRevs: "RepositoryRevision{1@HEAD}",
		WantCSV: autogold.Expect(`repository,revision,file_path,match_count,first_match_url
foo1,commitfoo0,,1,/foo1@commitfoo0/-/blob/?L2
`),
	})

	do("rev", newSearcherTestCase{
		Query:        "repo:foo rev:dev1 content",
		WantRefSpecs: "RepositoryRevSpec{1@dev1}",
		WantRepoRevs: "RepositoryRevision{1@dev1}",
		WantCSV: autogold.Expect(`repository,revision,file_path,match_count,first_match_url
foo1,commitfoo1,,1,/foo1@commitfoo1/-/blob/?L2
`),
	})

	do("glob", newSearcherTestCase{
		Query:        "repo:foo rev:*refs/heads/dev* content",
		WantRefSpecs: "RepositoryRevSpec{1@*refs/heads/dev*}",
		WantRepoRevs: "RepositoryRevision{1@dev1} RepositoryRevision{1@dev2}",
		WantCSV: autogold.Expect(`repository,revision,file_path,match_count,first_match_url
foo1,commitfoo1,,1,/foo1@commitfoo1/-/blob/?L2
foo1,commitfoo2,,1,/foo1@commitfoo2/-/blob/?L2
`),
	})

	do("global", newSearcherTestCase{
		Query:        "repo:. rev:*refs/heads/dev* content",
		WantRefSpecs: "RepositoryRevSpec{1@*refs/heads/dev*} RepositoryRevSpec{2@*refs/heads/dev*} RepositoryRevSpec{3@*refs/heads/dev*}",
		WantRepoRevs: "RepositoryRevision{1@dev1} RepositoryRevision{1@dev2} RepositoryRevision{2@dev1}",
		WantCSV: autogold.Expect(`repository,revision,file_path,match_count,first_match_url
foo1,commitfoo1,,1,/foo1@commitfoo1/-/blob/?L2
foo1,commitfoo2,,1,/foo1@commitfoo2/-/blob/?L2
bar2,commitbar1,,1,/bar2@commitbar1/-/blob/?L2
`),
	})

	do("notglob", newSearcherTestCase{
		Query:        "repo:foo rev:*refs/heads/dev*:*!refs/heads/dev1 content",
		WantRefSpecs: "RepositoryRevSpec{1@*refs/heads/dev*:*!refs/heads/dev1}",
		WantRepoRevs: "RepositoryRevision{1@dev2}",
		WantCSV: autogold.Expect(`repository,revision,file_path,match_count,first_match_url
foo1,commitfoo2,,1,/foo1@commitfoo2/-/blob/?L2
`),
	})

	do("nomatchglob", newSearcherTestCase{
		Query:        "repo:foo rev:*refs/heads/doesnotmatch* content",
		WantRefSpecs: "RepositoryRevSpec{1@*refs/heads/doesnotmatch*}",
	})

	do("norepos", newSearcherTestCase{
		Query: "repo:doesnotmatch content",
	})

	do("missingrev", newSearcherTestCase{
		Query:        "repo:foo rev:dev1:missing content",
		WantRefSpecs: "RepositoryRevSpec{1@dev1:missing}",
		WantRepoRevs: "RepositoryRevision{1@dev1}",
		WantCSV: autogold.Expect(`repository,revision,file_path,match_count,first_match_url
foo1,commitfoo1,,1,/foo1@commitfoo1/-/blob/?L2
`),
	})
}

type repoMock struct {
	ID       int
	Name     string
	Branches map[string]string
}

// mockSearchClient returns a client which will return matches. This exercises
// more of the search code path to give a bit more confidence we are correctly
// calling Plan and Execute vs a dumb SearchClient mock.
//
// Note: for now we only support nicely mocking zoekt. This isn't good enough
// to gain confidence in how this all works, so will follow up with making it
// possible to mock searcher.
func mockSearchClient(t *testing.T, repoMocks []repoMock) client.SearchClient {
	db := dbmocks.NewMockDB()
	db.ReposFunc.SetDefaultReturn(mockRepoStore(repoMocks))

	return client.Mocked(job.RuntimeClients{
		Logger:       logtest.Scoped(t),
		DB:           db,
		Zoekt:        mockZoekt(repoMocks),
		Gitserver:    mockGitserver(repoMocks),
		SearcherURLs: mockSearcher(t, repoMocks),
	})
}

func mockGitserver(repoMocks []repoMock) *gitserver.MockClient {
	get := func(name api.RepoName) (repoMock, error) {
		for _, repo := range repoMocks {
			if name == api.RepoName(repo.Name) {
				return repo, nil
			}
		}
		return repoMock{}, &gitdomain.RepoNotExistError{Repo: name}
	}

	gsClient := gitserver.NewMockClient()
	gsClient.ResolveRevisionFunc.SetDefaultHook(func(_ context.Context, name api.RepoName, spec string, _ gitserver.ResolveRevisionOptions) (api.CommitID, error) {
		repo, err := get(name)
		if err != nil {
			return "", err
		}
		if spec == "" {
			// Normally in search we treat the empty string has HEAD. In our
			// case we want to ensure we are explicit so will fail if this
			// happens.
			return "", errors.New("empty spec used instead of HEAD")
		}
		for branch, commit := range repo.Branches {
			if spec == branch || spec == commit {
				return api.CommitID(commit), nil
			}
		}
		return "", &gitdomain.RevisionNotFoundError{}
	})
	gsClient.ListRefsFunc.SetDefaultHook(func(_ context.Context, name api.RepoName) ([]gitdomain.Ref, error) {
		repo, err := get(name)
		if err != nil {
			return nil, err
		}
		var refs []gitdomain.Ref
		for branch, commit := range repo.Branches {
			refs = append(refs, gitdomain.Ref{
				Name:     "refs/heads/" + branch,
				CommitID: api.CommitID(commit),
			})
		}
		slices.SortFunc(refs, func(a, b gitdomain.Ref) bool {
			return a.Name < b.Name
		})
		return refs, nil
	})
	return gsClient
}

func mockRepoStore(repoMocks []repoMock) *dbmocks.MockRepoStore {
	repos := dbmocks.NewMockRepoStore()
	repos.ListMinimalReposFunc.SetDefaultHook(func(_ context.Context, opts database.ReposListOptions) (resp []types.MinimalRepo, _ error) {
		for _, repo := range repoMocks {
			keep := true
			for _, pat := range opts.IncludePatterns {
				keep = keep && (strings.Contains(repo.Name, pat) || pat == ".")
			}
			if !keep {
				continue
			}
			if len(opts.IDs) > 0 && !slices.Contains(opts.IDs, api.RepoID(repo.ID)) {
				continue
			}

			resp = append(resp, types.MinimalRepo{
				ID:   api.RepoID(repo.ID),
				Name: api.RepoName(repo.Name),
			})
		}
		return
	})
	return repos
}

func mockZoekt(repoMocks []repoMock) *searchbackend.FakeStreamer {
	var matches []zoekt.FileMatch
	for _, repo := range repoMocks {
		matches = append(matches, zoekt.FileMatch{
			RepositoryID: uint32(repo.ID),
			Repository:   repo.Name,
		})
	}
	return &searchbackend.FakeStreamer{
		Repos: []*zoekt.RepoListEntry{},
		Results: []*zoekt.SearchResult{{
			Files: matches,
		}},
	}
}

func mockSearcher(t *testing.T, repoMocks []repoMock) *endpoint.Map {
	searcher.MockSearchFilesInRepo = func(
		ctx context.Context,
		repo types.MinimalRepo,
		gitserverRepo api.RepoName,
		rev string,
		info *search.TextPatternInfo,
		fetchTimeout time.Duration,
		stream streaming.Sender,
	) (limitHit bool, err error) {
		found := false
		for _, r := range repoMocks {
			if api.RepoID(r.ID) == repo.ID {
				found = true
				commit, ok := r.Branches[rev]
				if !ok {
					return false, &gitdomain.RevisionNotFoundError{Spec: rev}
				}

				stream.Send(streaming.SearchEvent{
					Results: result.Matches{&result.FileMatch{
						File: result.File{
							Repo:     repo,
							CommitID: api.CommitID(commit),
						},
						ChunkMatches: result.ChunkMatches{{
							Content:      "line1",
							ContentStart: result.Location{Line: 1},
							Ranges: result.Ranges{{
								Start: result.Location{1, 1, 1},
								End:   result.Location{3, 1, 3},
							}},
						}},
					}}})
			}
		}
		if !found {
			return false, &gitdomain.RepoNotExistError{}
		}
		return false, nil
	}
	t.Cleanup(func() {
		searcher.MockSearchFilesInRepo = nil
	})
	return endpoint.Static("test")
}

func testNewSearcher(t *testing.T, ctx context.Context, newSearcher NewSearcher, tc newSearcherTestCase) {
	assert := require.New(t)

	userID := int32(1)
	ctx = actor.WithActor(ctx, actor.FromMockUser(userID))

	searcher, err := newSearcher.NewSearch(ctx, userID, tc.Query)
	assert.NoError(err)

	// Test RepositoryRevSpecs
	refSpecs, err := iterator.Collect(searcher.RepositoryRevSpecs(ctx))
	assert.NoError(err)
	assert.Equal(tc.WantRefSpecs, joinStringer(refSpecs))

	// Test ResolveRepositoryRevSpec
	var repoRevs []types2.RepositoryRevision
	for _, refSpec := range refSpecs {
		repoRevsPart, err := searcher.ResolveRepositoryRevSpec(ctx, refSpec)
		assert.NoError(err)
		repoRevs = append(repoRevs, repoRevsPart...)
	}
	assert.Equal(tc.WantRepoRevs, joinStringer(repoRevs))

	// Test Search
	var csv csvBuffer
	for _, repoRev := range repoRevs {
		err := searcher.Search(ctx, repoRev, &csv)
		assert.NoError(err)
	}
	if tc.WantCSV != nil {
		tc.WantCSV.Equal(t, csv.buf.String())
	}
}
