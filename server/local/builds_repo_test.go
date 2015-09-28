package local

import (
	"reflect"
	"testing"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sourcegraph/go-vcs/vcs"
	"sourcegraph.com/sqs/pbtypes"
)

func TestBuildsService_GetRepoBuildInfo_none(t *testing.T) {
	var s builds
	ctx, mock := testContext()

	testRepoRevSpec := sourcegraph.RepoRevSpec{RepoSpec: sourcegraph.RepoSpec{URI: "r"}, Rev: "r", CommitID: "c"}

	calledReposListCommits := mock.servers.Repos.MockListCommits(t)
	mock.stores.Builds.GetFirstInCommitOrder_ = func(context.Context, string, []string, bool) (*sourcegraph.Build, int, error) {
		return nil, -1, nil
	}

	if _, err := s.GetRepoBuildInfo(ctx, &sourcegraph.BuildsGetRepoBuildInfoOp{Repo: testRepoRevSpec}); grpc.Code(err) != codes.NotFound {
		t.Errorf("got err %v, want NotFound", err)
	}
	if !*calledReposListCommits {
		t.Error("!calledReposListCommits")
	}
}

func TestBuildsService_GetRepoBuildInfo_exactAndSuccessful(t *testing.T) {
	var s builds
	ctx, mock := testContext()

	time123 := pbtypes.NewTimestamp(time.Unix(123, 0).In(time.UTC))
	testRepoRevSpec := sourcegraph.RepoRevSpec{RepoSpec: sourcegraph.RepoSpec{URI: "r"}, Rev: "r"}
	exactCommit := &vcs.Commit{ID: "c", Author: vcs.Signature{Date: time123}}
	want := &sourcegraph.RepoBuildInfo{
		Exact:                &sourcegraph.Build{Attempt: 123, Repo: "r", Success: true, CommitID: "c"},
		LastSuccessful:       &sourcegraph.Build{Attempt: 123, Repo: "r", Success: true, CommitID: "c"},
		LastSuccessfulCommit: exactCommit,
	}

	calledReposGetCommit := mock.servers.Repos.MockGetCommit_Return_NoCheck(t, exactCommit)
	mock.stores.Builds.GetFirstInCommitOrder_ = func(context.Context, string, []string, bool) (*sourcegraph.Build, int, error) {
		return &sourcegraph.Build{Attempt: 123, Repo: "r", Success: true, CommitID: "c"}, 0, nil
	}

	buildInfo, err := s.GetRepoBuildInfo(ctx, &sourcegraph.BuildsGetRepoBuildInfoOp{Repo: testRepoRevSpec})
	if err != nil {
		t.Fatal(err)
	}
	if !*calledReposGetCommit {
		t.Error("!calledReposGetCommit")
	}
	if !reflect.DeepEqual(buildInfo, want) {
		t.Errorf("got %+v, want %+v", buildInfo, want)
	}
}

func TestBuildsService_GetRepoBuildInfo_oldSuccessful(t *testing.T) {
	var s builds
	ctx, mock := testContext()

	testRepoRevSpec := sourcegraph.RepoRevSpec{RepoSpec: sourcegraph.RepoSpec{URI: "r"}, Rev: "r"}
	exactCommit := &vcs.Commit{ID: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}
	lastSuccessfulCommit := &vcs.Commit{ID: "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"}
	want := &sourcegraph.RepoBuildInfo{
		Exact:                nil,
		LastSuccessful:       &sourcegraph.Build{Attempt: 123, Repo: "r", Success: true, CommitID: "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"},
		LastSuccessfulCommit: lastSuccessfulCommit,
		CommitsBehind:        1,
	}

	mock.servers.Repos.ListCommits_ = func(ctx context.Context, opt *sourcegraph.ReposListCommitsOp) (*sourcegraph.CommitList, error) {
		return &sourcegraph.CommitList{Commits: []*vcs.Commit{exactCommit, lastSuccessfulCommit}}, nil
	}
	mock.servers.Repos.GetCommit_ = func(ctx context.Context, opt *sourcegraph.RepoRevSpec) (*vcs.Commit, error) {
		return exactCommit, nil
	}
	mock.stores.Builds.GetFirstInCommitOrder_ = func(_ context.Context, repo string, commits []string, successfullOnly bool) (*sourcegraph.Build, int, error) {
		for i, commit := range commits {
			if commit == string(lastSuccessfulCommit.ID) {
				return &sourcegraph.Build{Attempt: 123, Repo: repo, Success: successfullOnly, CommitID: string(lastSuccessfulCommit.ID)}, i, nil
			}
		}
		return nil, 0, nil
	}

	buildInfo, err := s.GetRepoBuildInfo(ctx, &sourcegraph.BuildsGetRepoBuildInfoOp{Repo: testRepoRevSpec})
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(buildInfo, want) {
		t.Errorf("got %+v, want %+v", buildInfo, want)
	}
}

// func TestBuildsService_GetRepoBuildInfo_exactOptionSuccessful(t *testing.T) {
// 	time123 := time.Unix(123, 0).In(time.UTC)
// 	testRepoRevSpec := sourcegraph.RepoRevSpec{RepoSpec: sourcegraph.RepoSpec{URI: "r"}, Rev: "r"}
// 	want := &sourcegraph.RepoBuildInfo{
// 		Exact:                &sourcegraph.Build{BID: 123, Repo: "r", Success: true, CommitID: "c"},
// 		LastSuccessful:       &sourcegraph.Build{BID: 123, Repo: "r", Success: true, CommitID: "c"},
// 		LastSuccessfulCommit: &vcs.Commit{Commit: &vcs.Commit{ID: "c", Author: vcs.Signature{Date: time123}}},
// 	}

// 	var calledGetCommit bool
// 	var s builds
// 	ctx, mock := testContext()
//
// 	svc.Repos(ctx).(*MockReposService).Get_ = func(repoSpec sourcegraph.RepoSpec) (*sourcegraph.Repo, error) {
// 		if repoSpec != testRepoRevSpec.RepoSpec {
// 			t.Errorf("got RepoSpec %+v, want %+v", repoSpec, testRepoRevSpec.RepoSpec)
// 		}
// 		return &sourcegraph.Repo{URI: repoSpec.URI}, nil, nil
// 	}
// 	svc.Repos(ctx).(*MockReposService).GetCommit_ = func(repoRevSpec sourcegraph.RepoRevSpec, opt *sourcegraph.RepoGetCommitOptions) (*vcs.Commit, error) {
// 		if repoRevSpec.RepoSpec != testRepoRevSpec.RepoSpec {
// 			t.Errorf("got RepoSpec %+v, want %+v", repoRevSpec.RepoSpec, testRepoRevSpec.RepoSpec)
// 		}
// 		calledGetCommit = true
// 		return &vcs.Commit{Commit: &vcs.Commit{ID: "c", Author: vcs.Signature{Date: time123}}}, nil, nil
// 	}

// 	store.BuildsFromContext(ctx) = &mockstore.Builds{
// 		GetFirstInCommitOrder_: func(context.Context, sourcegraph.RepoSpec, []string, bool) (*sourcegraph.Build, int, error) {
// 			return &sourcegraph.Build{BID: 123, Repo: "r", Success: true, CommitID: "c"}, 0, nil
// 		},
// 	}

// 	buildInfo, err := svc.Builds(ctx).GetRepoBuildInfo(ctx, testRepoRevSpec, &sourcegraph.BuildsGetRepoBuildInfoOptions{Exact: true})
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	if !calledGetCommit {
// 		t.Error("!calledGetCommit")
// 	}

// 	if !reflect.DeepEqual(buildInfo, want) {
// 		t.Errorf("got BuildInfo != want BuildInfo\n\ngot buildInfo ========\n%s\n\nwant buildInfo ========\n%s", asJSON(buildInfo), asJSON(want))
// 	}
// }

// func TestBuildsService_GetRepoBuildInfo_exactOptionFailure(t *testing.T) {
// 	time123 := time.Unix(123, 0).In(time.UTC)
// 	testRepoRevSpec := sourcegraph.RepoRevSpec{RepoSpec: sourcegraph.RepoSpec{URI: "r"}, Rev: "r"}
// 	want := &sourcegraph.RepoBuildInfo{
// 		Exact:                &sourcegraph.Build{BID: 123, Repo: "r", Success: false, CommitID: "c"},
// 		LastSuccessful:       nil,
// 		LastSuccessfulCommit: nil,
// 	}

// 	var calledGetCommit bool
// 	var s builds
// 	ctx, mock := testContext()
//
// 	svc.Repos(ctx).(*MockReposService).Get_ = func(repoSpec sourcegraph.RepoSpec) (*sourcegraph.Repo, error) {
// 		if repoSpec != testRepoRevSpec.RepoSpec {
// 			t.Errorf("got RepoSpec %+v, want %+v", repoSpec, testRepoRevSpec.RepoSpec)
// 		}
// 		return &sourcegraph.Repo{URI: repoSpec.URI}, nil, nil
// 	}
// 	svc.Repos(ctx).(*MockReposService).GetCommit_ = func(repoRevSpec sourcegraph.RepoRevSpec, opt *sourcegraph.RepoGetCommitOptions) (*vcs.Commit, error) {
// 		if repoRevSpec.RepoSpec != testRepoRevSpec.RepoSpec {
// 			t.Errorf("got RepoSpec %+v, want %+v", repoRevSpec.RepoSpec, testRepoRevSpec.RepoSpec)
// 		}
// 		calledGetCommit = true
// 		return &vcs.Commit{Commit: &vcs.Commit{ID: "c", Author: vcs.Signature{Date: time123}}}, nil, nil
// 	}

// 	store.BuildsFromContext(ctx) = &mockstore.Builds{
// 		GetFirstInCommitOrder_: func(context.Context, sourcegraph.RepoSpec, []string, bool) (*sourcegraph.Build, int, error) {
// 			return &sourcegraph.Build{BID: 123, Repo: "r", Success: false, CommitID: "c"}, 0, nil
// 		},
// 	}

// 	buildInfo, err := svc.Builds(ctx).GetRepoBuildInfo(ctx, testRepoRevSpec, &sourcegraph.BuildsGetRepoBuildInfoOptions{Exact: true})
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	if !calledGetCommit {
// 		t.Error("!calledGetCommit")
// 	}
// 	if buildInfo.LastSuccessful != nil {
// 		t.Error("buildInfo.LastSuccessful != nil, buildInfo is exact on a failed build.")
// 	}

// 	if !reflect.DeepEqual(buildInfo, want) {
// 		t.Errorf("got BuildInfo != want BuildInfo\n\ngot buildInfo ========\n%s\n\nwant buildInfo ========\n%s", asJSON(buildInfo), asJSON(want))
// 	}
// }

// func TestBuildsService_updateRepoBuildStatus_disabled(t *testing.T) {
// 	var s builds
// 	ctx, mock := testContext()

// 	testRepoRevSpec := sourcegraph.RepoRevSpec{
// 		RepoSpec: sourcegraph.RepoSpec{URI: "github.com/o/r"},
// 		Rev:      "c",
// 		CommitID: "c",
// 	}

// 	var calledRepoGet, calledRepoGetSettings, calledRepoCreateStatus bool
//
// 	svc.Repos(ctx).(*MockReposService).GetSettings_ = func(repoSpec sourcegraph.RepoSpec) (*sourcegraph.RepoSettings, error) {
// 		if repoSpec != testRepoRevSpec.RepoSpec {
// 			t.Errorf("got RepoSpec %+v, want %+v", repoSpec, testRepoRevSpec.RepoSpec)
// 		}
// 		calledRepoGetSettings = true
// 		return &sourcegraph.RepoSettings{ExternalCommitStatuses: github.Bool(false)}, nil, nil
// 	}
// 	svc.Repos(ctx).(*MockReposService).CreateStatus_ = func(repoRevSpec sourcegraph.RepoRevSpec, st sourcegraph.RepoStatus) (*sourcegraph.RepoStatus, error) {
// 		if repoRevSpec != testRepoRevSpec {
// 			t.Errorf("got RepoRevSpec %+v, want %+v", repoRevSpec, testRepoRevSpec)
// 		}
// 		calledRepoCreateStatus = true
// 		return &st, nil, nil
// 	}
// 	s.Base.Repos = &mockstore.Repos{
// 		Get_: func(ctx context.Context, repo string) (*sourcegraph.Repo, error) {
// 			if repoSpec.URI != testRepoRevSpec.RepoSpec.URI {
// 				t.Errorf("got RepoSpec.URI %+v, want %+v", repoSpec.URI, testRepoRevSpec.RepoSpec.URI)
// 			}
// 			calledRepoGet = true
// 			return &sourcegraph.Repo{URI: repoSpec.URI, GitHubID: 1}, nil
// 		},
// 	}

// 	b := &sourcegraph.Build{
// 		Repo:     testRepoRevSpec.URI,
// 		CommitID: testRepoRevSpec.CommitID,
// 		Success:  true,
// 	}

// 	if err := svc.Builds(ctx).(*buildsService).updateRepoStatusForBuild(b); err != nil {
// 		t.Fatal(err)
// 	}

// 	if !calledRepoGet {
// 		t.Error("!calledRepoGet")
// 	}
// 	if !calledRepoGetSettings {
// 		t.Error("!calledRepoGetSettings")
// 	}
// 	if calledRepoCreateStatus {
// 		t.Error("calledRepoCreateStatus but settings should have disabled statuses")
// 	}
// }

// func TestBuildsService_updateRepoBuildStatus_enabled(t *testing.T) {
// 	var s builds
// 	ctx, mock := testContext()

// 	testRepoRevSpec := sourcegraph.RepoRevSpec{
// 		RepoSpec: sourcegraph.RepoSpec{URI: "github.com/o/r"},
// 		Rev:      "c",
// 		CommitID: "c",
// 	}

// 	var calledRepoGet, calledRepoGetSettings, calledRepoCreateStatus bool
//
// 	svc.Repos(ctx).(*MockReposService).GetSettings_ = func(repoSpec sourcegraph.RepoSpec) (*sourcegraph.RepoSettings, error) {
// 		if repoSpec != testRepoRevSpec.RepoSpec {
// 			t.Errorf("got RepoSpec %+v, want %+v", repoSpec, testRepoRevSpec.RepoSpec)
// 		}
// 		calledRepoGetSettings = true
// 		return &sourcegraph.RepoSettings{ExternalCommitStatuses: github.Bool(true), LastAdminUID: github.Int(1)}, nil, nil
// 	}
// 	svc.Repos(ctx).(*MockReposService).CreateStatus_ = func(repoRevSpec sourcegraph.RepoRevSpec, st sourcegraph.RepoStatus) (*sourcegraph.RepoStatus, error) {
// 		if repoRevSpec != testRepoRevSpec {
// 			t.Errorf("got RepoRevSpec %+v, want %+v", repoRevSpec, testRepoRevSpec)
// 		}
// 		if want := "success"; *st.State != want {
// 			t.Errorf("got repo status state %q, want %q", *st.State, want)
// 		}
// 		calledRepoCreateStatus = true
// 		return &st, nil, nil
// 	}
// 	s.Base.Repos = &mockstore.Repos{
// 		Get_: func(ctx context.Context, repo string) (*sourcegraph.Repo, error) {
// 			if repoSpec.URI != testRepoRevSpec.RepoSpec.URI {
// 				t.Errorf("got RepoSpec.URI %+v, want %+v", repoSpec.URI, testRepoRevSpec.RepoSpec.URI)
// 			}
// 			calledRepoGet = true
// 			return &sourcegraph.Repo{URI: repoSpec.URI, GitHubID: 1}, nil
// 		},
// 	}

// 	b := &sourcegraph.Build{
// 		Repo:     testRepoRevSpec.URI,
// 		CommitID: testRepoRevSpec.CommitID,
// 		Success:  true,
// 	}

// 	if err := svc.Builds(ctx).(*buildsService).updateRepoStatusForBuild(b); err != nil {
// 		t.Fatal(err)
// 	}

// 	if !calledRepoGet {
// 		t.Error("!calledRepoGet")
// 	}
// 	if !calledRepoGetSettings {
// 		t.Error("!calledRepoGetSettings")
// 	}
// 	if !calledRepoCreateStatus {
// 		t.Error("!calledRepoCreateStatus")
// 	}
// }

// func makeCommitsList(commitIDs ...string) []*vcs.Commit {
// 	cs := make([]*vcs.Commit, len(commitIDs))
// 	for i, id := range commitIDs {
// 		cs[i] = &vcs.Commit{Commit: &vcs.Commit{ID: vcs.CommitID(id), Author: vcs.Signature{Date: time.Unix(int64(i), 0)}}}
// 	}
// 	return cs
// }
