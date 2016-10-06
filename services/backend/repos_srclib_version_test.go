package backend

import (
	"strings"
	"testing"

	"context"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph/legacyerr"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs"
	"sourcegraph.com/sourcegraph/sourcegraph/services/backend/internal/localstore"
	srclibstore "sourcegraph.com/sourcegraph/srclib/store"
)

const (
	c1 = "1111111111111111111111111111111111111111"
	c2 = "2222222222222222222222222222222222222222"
	c3 = "3333333333333333333333333333333333333333"
)

func TestReposService_GetSrclibDataVersionForPath_exact(t *testing.T) {
	var s repos
	ctx := testContext()

	calledReposGet := localstore.Mocks.Repos.MockGet_Path(t, 1, "r")
	calledVersions := localstore.GraphMockVersions(&localstore.Mocks.Graph, &srclibstore.Version{Repo: "r", CommitID: strings.Repeat("c", 40)})

	dataVer, err := s.GetSrclibDataVersionForPath(ctx, &sourcegraph.TreeEntrySpec{
		RepoRev: sourcegraph.RepoRevSpec{Repo: 1, CommitID: strings.Repeat("c", 40)},
		Path:    "p",
	})
	if err != nil {
		t.Fatal(err)
	}
	if want := (sourcegraph.SrclibDataVersion{CommitID: strings.Repeat("c", 40)}); *dataVer != want {
		t.Fatalf("got %+v, want %+v", *dataVer, want)
	}
	if !*calledReposGet {
		t.Error("!calledReposGet")
	}
	if !*calledVersions {
		t.Error("!calledVersions")
	}
}

func TestReposService_GetSrclibDataVersionForPath_lookback_versionNewerThanLastCommitThatChangedFile(t *testing.T) {
	testReposService_GetSrclibDataVersionForPath_lookback(t, c2, 1)
}

func TestReposService_GetSrclibDataVersionForPath_lookback_versionSameAsLastCommitThatChangedFile(t *testing.T) {
	testReposService_GetSrclibDataVersionForPath_lookback(t, c3, 2)
}

func testReposService_GetSrclibDataVersionForPath_lookback(t *testing.T, versionCommitID string, commitsBehind int32) {
	var s repos
	ctx := testContext()

	calledReposGet := localstore.Mocks.Repos.MockGet_Path(t, 1, "r")
	calledVersions := localstore.GraphMockVersionsFiltered(&localstore.Mocks.Graph, &srclibstore.Version{Repo: "r", CommitID: versionCommitID})
	var calledListCommitsWithPath, calledListCommitsNoPath bool
	Mocks.Repos.ListCommits = func(ctx context.Context, op *sourcegraph.ReposListCommitsOp) (*sourcegraph.CommitList, error) {
		if op.Opt.Path != "" {
			// Return the last commit that changed the file "p".
			calledListCommitsWithPath = true
			return &sourcegraph.CommitList{Commits: []*vcs.Commit{{ID: c3}}}, nil
		}
		// Return all commits between c3 and v (inclusive).
		calledListCommitsNoPath = true
		return &sourcegraph.CommitList{Commits: []*vcs.Commit{{ID: c1}, {ID: c2}, {ID: c3}}}, nil
	}

	dataVer, err := s.GetSrclibDataVersionForPath(ctx, &sourcegraph.TreeEntrySpec{
		RepoRev: sourcegraph.RepoRevSpec{Repo: 1, CommitID: c1},
		Path:    "p",
	})
	if err != nil {
		t.Fatal(err)
	}
	if want := (sourcegraph.SrclibDataVersion{CommitID: versionCommitID, CommitsBehind: commitsBehind}); *dataVer != want {
		t.Fatalf("got %+v, want %+v", *dataVer, want)
	}
	if !*calledReposGet {
		t.Error("!calledReposGet")
	}
	if !*calledVersions {
		t.Error("!calledVersions")
	}
	if !calledListCommitsWithPath {
		t.Error("!calledListCommitsWithPath")
	}
	if !calledListCommitsNoPath {
		t.Error("!calledListCommitsNoPath")
	}
}

func TestReposService_GetSrclibDataVersionForPath_notFoundNoVersionsNoCommits(t *testing.T) {
	var s repos
	ctx := testContext()

	calledReposGet := localstore.Mocks.Repos.MockGet_Path(t, 1, "r")
	calledVersions := localstore.GraphMockVersions(&localstore.Mocks.Graph)
	calledListCommits := Mocks.Repos.MockListCommits(t)

	_, err := s.GetSrclibDataVersionForPath(ctx, &sourcegraph.TreeEntrySpec{
		RepoRev: sourcegraph.RepoRevSpec{Repo: 1, CommitID: strings.Repeat("c", 40)},
		Path:    "p",
	})
	if legacyerr.ErrCode(err) != legacyerr.NotFound {
		t.Fatalf("got error %v, want NotFound", err)
	}
	if !*calledReposGet {
		t.Error("!calledReposGet")
	}
	if !*calledVersions {
		t.Error("!calledVersions")
	}
	if !*calledListCommits {
		t.Error("!calledListCommits")
	}
}

func TestReposService_GetSrclibDataVersionForPath_notFoundWrongVersionsNoCommits(t *testing.T) {
	var s repos
	ctx := testContext()

	calledReposGet := localstore.Mocks.Repos.MockGet_Path(t, 1, "r")
	calledVersions := localstore.GraphMockVersionsFiltered(&localstore.Mocks.Graph, &srclibstore.Version{Repo: "r", CommitID: "x"})
	calledListCommits := Mocks.Repos.MockListCommits(t)

	_, err := s.GetSrclibDataVersionForPath(ctx, &sourcegraph.TreeEntrySpec{
		RepoRev: sourcegraph.RepoRevSpec{Repo: 1, CommitID: strings.Repeat("c", 40)},
		Path:    "p",
	})
	if legacyerr.ErrCode(err) != legacyerr.NotFound {
		t.Fatalf("got error %v, want NotFound", err)
	}
	if !*calledReposGet {
		t.Error("!calledReposGet")
	}
	if !*calledVersions {
		t.Error("!calledVersions")
	}
	if !*calledListCommits {
		t.Error("!calledListCommits")
	}
}

func TestReposService_GetSrclibDataVersionForPath_notFoundNoVersionsWrongCommits(t *testing.T) {
	var s repos
	ctx := testContext()

	calledReposGet := localstore.Mocks.Repos.MockGet_Path(t, 1, "r")
	calledVersions := localstore.GraphMockVersions(&localstore.Mocks.Graph)
	calledListCommits := Mocks.Repos.MockListCommits(t, "x")

	_, err := s.GetSrclibDataVersionForPath(ctx, &sourcegraph.TreeEntrySpec{
		RepoRev: sourcegraph.RepoRevSpec{Repo: 1, CommitID: strings.Repeat("c", 40)},
		Path:    "p",
	})
	if legacyerr.ErrCode(err) != legacyerr.NotFound {
		t.Fatalf("got error %v, want NotFound", err)
	}
	if !*calledReposGet {
		t.Error("!calledReposGet")
	}
	if !*calledVersions {
		t.Error("!calledVersions")
	}
	if !*calledListCommits {
		t.Error("!calledListCommits")
	}
}
