package local

import (
	"encoding/json"
	"reflect"
	"testing"

	"golang.org/x/net/context"

	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sourcegraph/go-vcs/vcs"
	vcstest "sourcegraph.com/sourcegraph/go-vcs/vcs/testing"
	localcli "src.sourcegraph.com/sourcegraph/server/local/cli"
)

type listCommitsTest struct {
	repoRev       sourcegraph.RepoRevSpec
	actualCommits []*vcs.Commit
	wantCommits   []*vcs.Commit
	cachedCommits []*vcs.Commit
	refreshCache  bool
}

func Test_ListCommits(t *testing.T) {
	original := localcli.Flags
	localcli.Flags.CommitLogCachePeriod = 1 // We're testing caching behavior in this test, so caching needs to be enabled (any non-zero cache period value will do).
	defer func() {
		localcli.Flags = original
	}()

	rr1 := sourcegraph.RepoRevSpec{
		RepoSpec: sourcegraph.RepoSpec{URI: "r"},
		Rev:      "master",
	}
	commits1 := []*vcs.Commit{
		{ID: vcs.CommitID("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")},
		{ID: vcs.CommitID("bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb")},
		{ID: vcs.CommitID("cccccccccccccccccccccccccccccccccccccccc")},
		{ID: vcs.CommitID("dddddddddddddddddddddddddddddddddddddddd")},
	}
	commits2 := []*vcs.Commit{
		{ID: vcs.CommitID("bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb")},
		{ID: vcs.CommitID("cccccccccccccccccccccccccccccccccccccccc")},
		{ID: vcs.CommitID("dddddddddddddddddddddddddddddddddddddddd")},
	}

	tests := []listCommitsTest{{
		// refreshCache == true, fetch actual
		repoRev:       rr1,
		actualCommits: commits1,
		wantCommits:   commits1,
		cachedCommits: nil,
		refreshCache:  true,
	}, {
		// refreshCache == false, hit cache
		repoRev:       rr1,
		actualCommits: nil,
		wantCommits:   commits1,
		cachedCommits: commits1,
		refreshCache:  false,
	}, {
		// hit stale cache
		repoRev:       rr1,
		actualCommits: commits1,
		wantCommits:   commits2,
		cachedCommits: commits2,
		refreshCache:  false,
	}, {
		// stale cache with refreshCache == true
		repoRev:       rr1,
		actualCommits: commits1,
		wantCommits:   commits1,
		cachedCommits: commits2,
		refreshCache:  true,
	}}

	for _, test := range tests {
		test_ListCommits(t, test)
	}
}

func test_ListCommits(t *testing.T, test listCommitsTest) {
	var s repos
	ctx, mock := testContext()

	if test.refreshCache { // these methods should only be called if RefreshCache == true
		mockRepo := vcstest.MockRepository{}
		mockRepo.ResolveRevision_ = func(spec string) (vcs.CommitID, error) {
			if spec != "master" {
				t.Fatalf("call to ResolveRevision with unexpected argument spec=%s", spec)
			}
			return test.actualCommits[0].ID, nil
		}
		mockRepo.Commits_ = func(opt vcs.CommitsOptions) ([]*vcs.Commit, uint, error) {
			if !(opt.Head == test.actualCommits[0].ID && opt.Base == "") {
				t.Fatalf("call to Commits with unexpected argument opt=%+v", opt)
			}
			return test.actualCommits, uint(len(test.actualCommits)), nil
		}
		mock.stores.RepoVCS.Open_ = func(ctx context.Context, repo string) (vcs.Repository, error) {
			return mockRepo, nil
		}
		mock.servers.RepoStatuses.Create_ = func(ctx context.Context, opt *sourcegraph.RepoStatusesCreateOp) (*sourcegraph.RepoStatus, error) {
			if !reflect.DeepEqual(test.repoRev, opt.Repo) {
				t.Fatalf("RepoStatuses.Create: want opt.Repo == %+v, got %+v", test.repoRev.RepoSpec, opt.Repo)
			}
			if opt.Status.Context != "graph_data_commit" {
				t.Fatalf("RepoStatuses.Create: want opt.Status.Context == \"graph_data_commit\", got %s", opt.Status.Context)
			}

			var commitList sourcegraph.CommitList
			if err := json.Unmarshal([]byte(opt.Status.Description), &commitList); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(test.actualCommits, commitList.Commits) {
				t.Fatalf("RepoStatuses.Create: want commit args %+v, got %+v", test.actualCommits, commitList.Commits)
			}

			// Update test.cachedCommits (in a real environment, this would actually update the repo
			// status in the repo status store
			test.cachedCommits = test.actualCommits

			return nil, nil // no-op because GetCombined_ is mocked, too
		}
	}

	mock.servers.RepoStatuses.GetCombined_ = func(context.Context, *sourcegraph.RepoRevSpec) (*sourcegraph.CombinedStatus, error) {
		b, err := json.Marshal(sourcegraph.CommitList{Commits: test.cachedCommits})
		if err != nil {
			t.Fatal(err)
		}
		return &sourcegraph.CombinedStatus{
			Statuses: []*sourcegraph.RepoStatus{{
				Context:     "graph_data_commit",
				Description: string(b),
			}},
		}, nil
	}

	commitList, err := s.ListCommits(ctx, &sourcegraph.ReposListCommitsOp{
		Repo: test.repoRev.RepoSpec,
		Opt: &sourcegraph.RepoListCommitsOptions{
			Head:         test.repoRev.Rev,
			RefreshCache: test.refreshCache,
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(test.wantCommits, commitList.Commits) {
		t.Errorf("want %+v, got %+v", test.wantCommits, commitList.Commits)
	}
}
