package backend

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/types"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/vcs/git"
)

func TestRepos_ResolveRev_noRevSpecified_getsDefaultBranch(t *testing.T) {
	ctx := testContext()

	want := strings.Repeat("a", 40)

	var calledVCSRepoResolveRevision bool
	Mocks.Repos.MockVCS(t, "a")
	git.Mocks.ResolveRevision = func(rev string, opt *git.ResolveRevisionOptions) (api.CommitID, error) {
		calledVCSRepoResolveRevision = true
		return api.CommitID(want), nil
	}
	defer git.ResetMocks()

	// (no rev/branch specified)
	commitID, err := Repos.ResolveRev(ctx, &types.Repo{URI: "a"}, "")
	if err != nil {
		t.Fatal(err)
	}
	if !calledVCSRepoResolveRevision {
		t.Error("!calledVCSRepoResolveRevision")
	}
	if string(commitID) != want {
		t.Errorf("got resolved commit %q, want %q", commitID, want)
	}
}

func TestRepos_ResolveRev_noCommitIDSpecified_resolvesRev(t *testing.T) {
	ctx := testContext()

	want := strings.Repeat("a", 40)

	var calledVCSRepoResolveRevision bool
	Mocks.Repos.MockVCS(t, "a")
	git.Mocks.ResolveRevision = func(rev string, opt *git.ResolveRevisionOptions) (api.CommitID, error) {
		calledVCSRepoResolveRevision = true
		return api.CommitID(want), nil
	}
	defer git.ResetMocks()

	commitID, err := Repos.ResolveRev(ctx, &types.Repo{URI: "a"}, "b")
	if err != nil {
		t.Fatal(err)
	}
	if !calledVCSRepoResolveRevision {
		t.Error("!calledVCSRepoResolveRevision")
	}
	if string(commitID) != want {
		t.Errorf("got resolved commit %q, want %q", commitID, want)
	}
}

func TestRepos_ResolveRev_commitIDSpecified_resolvesCommitID(t *testing.T) {
	ctx := testContext()

	want := strings.Repeat("a", 40)

	var calledVCSRepoResolveRevision bool
	Mocks.Repos.MockVCS(t, "a")
	git.Mocks.ResolveRevision = func(rev string, opt *git.ResolveRevisionOptions) (api.CommitID, error) {
		calledVCSRepoResolveRevision = true
		return api.CommitID(want), nil
	}
	defer git.ResetMocks()

	commitID, err := Repos.ResolveRev(ctx, &types.Repo{URI: "a"}, strings.Repeat("a", 40))
	if err != nil {
		t.Fatal(err)
	}
	if !calledVCSRepoResolveRevision {
		t.Error("!calledVCSRepoResolveRevision")
	}
	if string(commitID) != want {
		t.Errorf("got resolved commit %q, want %q", commitID, want)
	}
}

func TestRepos_ResolveRev_commitIDSpecified_failsToResolve(t *testing.T) {
	ctx := testContext()

	want := errors.New("x")

	var calledVCSRepoResolveRevision bool
	Mocks.Repos.MockVCS(t, "a")
	git.Mocks.ResolveRevision = func(rev string, opt *git.ResolveRevisionOptions) (api.CommitID, error) {
		calledVCSRepoResolveRevision = true
		return "", errors.New("x")
	}
	defer git.ResetMocks()

	_, err := Repos.ResolveRev(ctx, &types.Repo{URI: "a"}, strings.Repeat("a", 40))
	if !reflect.DeepEqual(err, want) {
		t.Fatalf("got err %v, want %v", err, want)
	}
	if !calledVCSRepoResolveRevision {
		t.Error("!calledVCSRepoResolveRevision")
	}
}
