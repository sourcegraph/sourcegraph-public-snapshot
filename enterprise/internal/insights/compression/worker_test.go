package compression

import (
	"context"
	"database/sql"
	"math"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"golang.org/x/time/rate"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/schema"
)

var ops *operations = newOperations(&observation.TestContext)

func TestCommitIndexer_indexAll(t *testing.T) {
	ctx := context.Background()
	commitStore := NewMockCommitStore()

	maxHistorical := time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC)
	clock := func() time.Time { return time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC) }

	indexer := CommitIndexer{
		limiter:           rate.NewLimiter(10, 1),
		commitStore:       commitStore,
		maxHistoricalTime: maxHistorical,
		background:        context.Background(),
		operations:        ops,
		clock:             clock,
	}

	// Testing a scenario with 3 repos
	// "repo-one" has commits but has disabled indexing
	// "really-big-repo" has commits and has enabled indexing, it should update
	// "no-commits" has no commits but is enabled, and will not update the index but will update the metadata
	commits := map[string][]*gitdomain.Commit{
		"repo-one": {
			commit("ref1", "2020-05-01T00:00:00+00:00"),
			commit("ref2", "2020-05-10T00:00:00+00:00"),
			commit("ref3", "2020-05-15T00:00:00+00:00"),
			commit("ref4", "2020-05-20T00:00:00+00:00"),
		},
		"really-big-repo": {
			commit("bigref1", "1999-04-01T00:00:00+00:00"),
			commit("bigref2", "1999-04-03T00:00:00+00:00"),
			commit("bigref3", "1999-04-06T00:00:00+00:00"),
			commit("bigref4", "1999-04-09T00:00:00+00:00"),
		},
		"no-commits": {},
	}
	indexer.getCommits = mockCommits(commits)
	indexer.allReposIterator = mockIterator([]string{"repo-one", "really-big-repo", "no-commits"})

	commitStore.GetMetadataFunc.PushReturn(CommitIndexMetadata{
		RepoId:        1,
		Enabled:       false,
		LastIndexedAt: time.Now(),
	}, nil)

	commitStore.GetMetadataFunc.PushReturn(CommitIndexMetadata{
		RepoId:        2,
		Enabled:       true,
		LastIndexedAt: time.Now(),
	}, nil)

	commitStore.GetMetadataFunc.PushReturn(CommitIndexMetadata{
		RepoId:        3,
		Enabled:       true,
		LastIndexedAt: time.Now(),
	}, nil)

	t.Run("multi_repository", func(t *testing.T) {
		pageSize := 0
		conf.Mock(&conf.Unified{
			SiteConfiguration: schema.SiteConfiguration{
				InsightsCommitIndexerPageSize: pageSize,
			},
		})
		defer conf.Mock(nil)
		err := indexer.indexAll(ctx)
		if err != nil {
			t.Fatal(err)
		}

		// Three repos get metadata, one is disabled, the other two are enabled
		if got, want := len(commitStore.GetMetadataFunc.history), 3; got != want {
			t.Errorf("got GetMetadata invocations: %v want %v", got, want)
		}

		// Only one repository should actually update any commits
		if got, want := len(commitStore.InsertCommitsFunc.history), 1; got != want {
			t.Errorf("got InsertCommits invocations: %v want %v", got, want)
		} else {
			call := commitStore.InsertCommitsFunc.history[0]
			for i, got := range call.Arg2 {
				if diff := cmp.Diff(commits["really-big-repo"][i], got); diff != "" {
					t.Errorf("unexpected commit\n%s", diff)
				}
			}
		}

		// One repository had no commits, so only the timestamp would get updated
		if got, want := len(commitStore.UpsertMetadataStampFunc.history), 1; got != want {
			t.Errorf("got UpsertMetadataStamp invocations: %v want %v", got, want)
		} else {
			call := commitStore.UpsertMetadataStampFunc.history[0]
			if call.Arg1 != 2 {
				t.Errorf("unexpected repository for UpsertMetadataStamp repo_id: %v", call.Arg1)
			}
		}
	})
}

func Test_getMetadata_InsertNewRecord(t *testing.T) {
	ctx := context.Background()
	commitStore := NewMockCommitStore()

	expected := CommitIndexMetadata{
		RepoId:        1,
		Enabled:       true,
		LastIndexedAt: time.Date(2018, 1, 1, 1, 1, 1, 1, time.UTC),
	}

	// this test will get no results from the metadata table and will insert one
	commitStore.GetMetadataFunc.PushReturn(CommitIndexMetadata{}, sql.ErrNoRows)
	commitStore.UpsertMetadataStampFunc.PushReturn(expected, nil)

	t.Run("create_new_metadata", func(t *testing.T) {
		metadata, err := getMetadata(ctx, 1, commitStore)
		if err != nil {
			t.Fatal(err)
		}

		if diff := cmp.Diff(expected, metadata); diff != "" {
			t.Errorf("unexpected metadata\n%s", diff)
		}

		if got, want := len(commitStore.GetMetadataFunc.history), 1; got != want {
			t.Errorf("unexpected GetMetadata invocations %v", 1)
		}

		if got, want := len(commitStore.UpsertMetadataStampFunc.history), 1; got != want {
			t.Errorf("unexpected UpsertMetadataStamp invocations %v", 1)
		}
	})
}

func Test_getMetadata_NoInsertRequired(t *testing.T) {
	ctx := context.Background()
	commitStore := NewMockCommitStore()

	expected := CommitIndexMetadata{
		RepoId:        1,
		Enabled:       true,
		LastIndexedAt: time.Date(2018, 1, 1, 1, 1, 1, 1, time.UTC),
	}
	// will get results immediately and will not insert a new row
	commitStore.GetMetadataFunc.PushReturn(expected, nil)

	t.Run("get_metadata", func(t *testing.T) {
		metadata, err := getMetadata(ctx, 1, commitStore)

		if err != nil {
			t.Fatal(err)
		}

		if diff := cmp.Diff(expected, metadata); diff != "" {
			t.Errorf("unexpected metadata\n%s", diff)
		}

		if got, want := len(commitStore.GetMetadataFunc.history), 1; got != want {
			t.Errorf("unexpected GetMetadata invocations %v", 1)
		}

		if got, want := len(commitStore.UpsertMetadataStampFunc.history), 0; got != want {
			t.Errorf("unexpected UpsertMetadataStamp invocations %v", 1)
		}
	})
}

func TestCommitIndexer_paging(t *testing.T) {
	ctx := context.Background()
	commitStore := NewMockCommitStore()

	maxHistorical := time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC)
	clock := func() time.Time { return time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC) }

	indexer := CommitIndexer{
		limiter:           rate.NewLimiter(10, 1),
		commitStore:       commitStore,
		maxHistoricalTime: maxHistorical,
		background:        context.Background(),
		operations:        ops,
		clock:             clock,
	}

	// Testing a scenario with 3 repos
	// "repo-one" has less than 1 page of commits it should update index and metadata
	// "really-big-repo" has 2 pages of commits, it should update index and metadata
	// "no-commits" has no commits but is enabled, and will not update the index but will update the metadata
	commits := map[string][]*gitdomain.Commit{
		"repo-one": {
			commit("ref1", "2020-05-01T00:00:00+00:00"),
			commit("ref2", "2020-05-10T00:00:00+00:00"),
		},
		"really-big-repo": {
			commit("bigref1", "1999-04-01T00:00:00+00:00"),
			commit("bigref2", "1999-04-03T00:00:00+00:00"),
			commit("bigref3", "1999-04-06T00:00:00+00:00"),
			commit("bigref4", "1999-04-09T00:00:00+00:00"),
		},
		"no-commits": {},
	}
	indexer.getCommits = mockCommits(commits)
	indexer.allReposIterator = mockIterator([]string{"repo-one", "really-big-repo", "no-commits"})

	commitStore.GetMetadataFunc.PushReturn(CommitIndexMetadata{
		RepoId:        1,
		Enabled:       true,
		LastIndexedAt: time.Now(),
	}, nil)

	commitStore.GetMetadataFunc.PushReturn(CommitIndexMetadata{
		RepoId:        2,
		Enabled:       true,
		LastIndexedAt: time.Now(),
	}, nil)

	commitStore.GetMetadataFunc.PushReturn(CommitIndexMetadata{
		RepoId:        2,
		Enabled:       true,
		LastIndexedAt: time.Now(),
	}, nil)

	commitStore.GetMetadataFunc.PushReturn(CommitIndexMetadata{
		RepoId:        3,
		Enabled:       true,
		LastIndexedAt: time.Now(),
	}, nil)

	t.Run("multi_repository_paging", func(t *testing.T) {
		pageSize := 3
		conf.Mock(&conf.Unified{
			SiteConfiguration: schema.SiteConfiguration{
				InsightsCommitIndexerPageSize: pageSize,
			},
		})
		defer conf.Mock(nil)
		err := indexer.indexAll(ctx)
		if err != nil {
			t.Fatal(err)
		}

		// Three enabled repos get metadata, repo 2 had 2 pages of commits so it makes 2 metadata calls
		if got, want := len(commitStore.GetMetadataFunc.history), 4; got != want {
			t.Errorf("got GetMetadata invocations: %v want %v", got, want)
		}

		// Repo 1 and 2 have commits, repo 1 has 1 page and repo 2 has 2 pages so 3 total calls
		if got, want := len(commitStore.InsertCommitsFunc.history), 3; got != want {
			t.Errorf("got InsertCommits invocations: %v want %v", got, want)
		} else {
			repo1Page1 := commitStore.InsertCommitsFunc.history[0]
			repo2Page1 := commitStore.InsertCommitsFunc.history[1]
			repo2Page2 := commitStore.InsertCommitsFunc.history[2]
			// All commits from repo-one because it was less than a page size
			checkCommits(t, commits["repo-one"], repo1Page1.Arg2)
			// One page of commits for really-big-repo because it was > one page
			checkCommits(t, commits["really-big-repo"][:pageSize+1], repo2Page1.Arg2)
			// The rest of the commits for really-big-repo
			checkCommits(t, commits["really-big-repo"][pageSize:], repo2Page2.Arg2)

			// "Current time" for repo-one because commit history did not fill a full page
			checkIndexedThough(t, clock().UTC(), repo1Page1.Arg3)
			// Time of last indexed commit because really big repo's first call was a full page
			checkIndexedThough(t, commits["really-big-repo"][pageSize-1].Committer.Date, repo2Page1.Arg3)
			// "Current time" for page two of really-big-repo because remaining commit history did not fill a full page
			checkIndexedThough(t, clock().UTC(), repo2Page2.Arg3)
		}

		// One repository had no commits, so only the timestamp would get updated
		if got, want := len(commitStore.UpsertMetadataStampFunc.history), 1; got != want {
			t.Errorf("got UpsertMetadataStamp invocations: %v want %v", got, want)
		} else {
			call := commitStore.UpsertMetadataStampFunc.history[0]
			if call.Arg1 != 2 {
				t.Errorf("unexpected repository for UpsertMetadataStamp repo_id: %v", call.Arg1)
			}
		}
	})
}

func checkCommits(t *testing.T, want []*gitdomain.Commit, got []*gitdomain.Commit) {
	for i, commit := range got {
		if diff := cmp.Diff(want[i], commit); diff != "" {
			t.Errorf("unexpected commit\n%s", diff)
		}
	}
}

func checkIndexedThough(t *testing.T, want time.Time, got time.Time) {
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("unexpected indexed through date\n%s", diff)
	}
}

// mockIterator generates iterator methods given a list of repo names for test scenarios
func mockIterator(repos []string) func(ctx context.Context, each func(repoName string, id api.RepoID) error) error {
	return func(ctx context.Context, each func(repoName string, id api.RepoID) error) error {
		for i, repo := range repos {
			err := each(repo, api.RepoID(i))
			if err != nil {
				return err
			}
		}
		return nil
	}
}

// commit build a fake commit for test scenarios
func commit(ref string, commitTime string) *gitdomain.Commit {
	t, _ := time.Parse(time.RFC3339, commitTime)

	return &gitdomain.Commit{
		ID:        api.CommitID(ref),
		Committer: &gitdomain.Signature{Date: t},
	}
}

func mockCommits(commits map[string][]*gitdomain.Commit) func(ctx context.Context, db database.DB, name api.RepoName, after time.Time, pageSize int, operation *observation.Operation) ([]*gitdomain.Commit, error) {
	repoPages := map[string]int{}
	return func(ctx context.Context, db database.DB, name api.RepoName, after time.Time, pageSize int, operation *observation.Operation) ([]*gitdomain.Commit, error) {
		if pageSize == 0 {
			pageSize = len(commits[(string(name))])
		}
		curentPage := repoPages[string(name)]
		repoPages[string(name)] = repoPages[string(name)] + 1
		startItem := curentPage * pageSize
		endItem := int(math.Min(float64(startItem+pageSize), float64(len(commits[(string(name))]))))
		return commits[(string(name))][startItem:endItem], nil
	}
}
