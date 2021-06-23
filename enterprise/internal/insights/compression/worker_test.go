package compression

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"golang.org/x/time/rate"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/internal/vcs/git"
)

func TestCommitIndexer_indexAll(t *testing.T) {
	ctx := context.Background()
	commitStore := NewMockCommitStore()

	maxHistorical := time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC)

	indexer := CommitIndexer{
		limiter:           rate.NewLimiter(10, 1),
		commitStore:       commitStore,
		maxHistoricalTime: maxHistorical,
		background:        context.Background(),
	}

	// Testing a scenario with 3 repos
	// "repo-one" has commits but has disabled indexing
	// "really-big-repo" has commits and has enabled indexing, it should update
	// "no-commits" has no commits but is enabled, and will not update the index but will update the metadata
	commits := map[string][]*git.Commit{
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
	indexer.getRepoID = mockIds(map[string]int{"repo-one": 1, "really-big-repo": 2, "no-commits": 3})
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
			if call.Arg1 != 3 {
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

// mockIterator generates iterator methods given a list of repo names for test scenarios
func mockIterator(repos []string) func(ctx context.Context, each func(repoName string) error) error {
	return func(ctx context.Context, each func(repoName string) error) error {
		for _, repo := range repos {
			err := each(repo)
			if err != nil {
				return err
			}
		}
		return nil
	}
}

// commit build a fake commit for test scenarios
func commit(ref string, commitTime string) *git.Commit {
	t, _ := time.Parse(time.RFC3339, commitTime)

	return &git.Commit{
		ID:        api.CommitID(ref),
		Committer: &git.Signature{Date: t},
	}
}

func mockCommits(commits map[string][]*git.Commit) func(ctx context.Context, name api.RepoName, after time.Time) ([]*git.Commit, error) {
	return func(ctx context.Context, name api.RepoName, after time.Time) ([]*git.Commit, error) {
		return commits[(string(name))], nil
	}
}

func mockIds(ids map[string]int) func(ctx context.Context, name api.RepoName) (*types.Repo, error) {
	return func(ctx context.Context, name api.RepoName) (*types.Repo, error) {
		id := ids[string(name)]
		return &types.Repo{ID: api.RepoID(id)}, nil
	}
}
