package backend

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	"context"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	vcstest "sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs/testing"
)

func TestRepos_ResolveRev_noRevSpecified_getsDefaultBranch(t *testing.T) {
	ctx := testContext()

	want := strings.Repeat("a", 40)

	var calledVCSRepoResolveRevision bool
	Mocks.Repos.MockOpenVCS(t, 1, vcstest.MockRepository{
		ResolveRevision_: func(ctx context.Context, rev string) (api.CommitID, error) {
			calledVCSRepoResolveRevision = true
			return api.CommitID(want), nil
		},
	})

	// (no rev/branch specified)
	commitID, err := Repos.ResolveRev(ctx, 1, "")
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
	Mocks.Repos.MockOpenVCS(t, 1, vcstest.MockRepository{
		ResolveRevision_: func(ctx context.Context, rev string) (api.CommitID, error) {
			calledVCSRepoResolveRevision = true
			return api.CommitID(want), nil
		},
	})

	commitID, err := Repos.ResolveRev(ctx, 1, "b")
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
	Mocks.Repos.MockOpenVCS(t, 1, vcstest.MockRepository{
		ResolveRevision_: func(ctx context.Context, rev string) (api.CommitID, error) {
			calledVCSRepoResolveRevision = true
			return api.CommitID(want), nil
		},
	})

	commitID, err := Repos.ResolveRev(ctx, 1, strings.Repeat("a", 40))
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
	Mocks.Repos.MockOpenVCS(t, 1, vcstest.MockRepository{
		ResolveRevision_: func(ctx context.Context, rev string) (api.CommitID, error) {
			calledVCSRepoResolveRevision = true
			return "", errors.New("x")
		},
	})

	_, err := Repos.ResolveRev(ctx, 1, strings.Repeat("a", 40))
	if !reflect.DeepEqual(err, want) {
		t.Fatalf("got err %v, want %v", err, want)
	}
	if !calledVCSRepoResolveRevision {
		t.Error("!calledVCSRepoResolveRevision")
	}
}
