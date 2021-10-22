package janitor

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/derision-test/glock"
	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/stores/dbstore"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/vcs/git"
)

func TestUnknownCommitsJanitor(t *testing.T) {
	resolveRevisionFunc := func(commit string) error {
		return nil
	}

	testUnknownCommitsJanitor(t, resolveRevisionFunc, []refreshCommitResolvabilityFuncInvocation{
		{1, "foo-x", false}, {1, "foo-y", false}, {1, "foo-z", false},
		{2, "bar-x", false}, {2, "bar-y", false}, {2, "bar-z", false},
		{3, "baz-x", false}, {3, "baz-y", false}, {3, "baz-z", false},
	})
}

func TestUnknownCommitsJanitorUnknownCommit(t *testing.T) {
	resolveRevisionFunc := func(commit string) error {
		if commit == "foo-y" || commit == "bar-x" || commit == "baz-z" {
			return &gitdomain.RevisionNotFoundError{}
		}

		return nil
	}

	testUnknownCommitsJanitor(t, resolveRevisionFunc, []refreshCommitResolvabilityFuncInvocation{
		{1, "foo-x", false}, {1, "foo-y", true}, {1, "foo-z", false},
		{2, "bar-x", true}, {2, "bar-y", false}, {2, "bar-z", false},
		{3, "baz-x", false}, {3, "baz-y", false}, {3, "baz-z", true},
	})
}

func TestUnknownCommitsJanitorUnknownRepository(t *testing.T) {
	resolveRevisionFunc := func(commit string) error {
		if strings.HasPrefix(commit, "foo-") {
			return &gitdomain.RepoNotExistError{}
		}

		return nil
	}

	testUnknownCommitsJanitor(t, resolveRevisionFunc, []refreshCommitResolvabilityFuncInvocation{
		{1, "foo-x", false}, {1, "foo-y", false}, {1, "foo-z", false},
		{2, "bar-x", false}, {2, "bar-y", false}, {2, "bar-z", false},
		{3, "baz-x", false}, {3, "baz-y", false}, {3, "baz-z", false},
	})
}

type refreshCommitResolvabilityFuncInvocation struct {
	RepositoryID int
	Commit       string
	Delete       bool
}

var testSourcedCommits = []dbstore.SourcedCommits{
	{RepositoryID: 1, RepositoryName: "foo", Commits: []string{"foo-x", "foo-y", "foo-z"}},
	{RepositoryID: 2, RepositoryName: "bar", Commits: []string{"bar-x", "bar-y", "bar-z"}},
	{RepositoryID: 3, RepositoryName: "baz", Commits: []string{"baz-x", "baz-y", "baz-z"}},
}

func testUnknownCommitsJanitor(t *testing.T, resolveRevisionFunc func(commit string) error, expectedCalls []refreshCommitResolvabilityFuncInvocation) {
	git.Mocks.ResolveRevision = func(spec string, opt git.ResolveRevisionOptions) (api.CommitID, error) {
		return api.CommitID(spec), resolveRevisionFunc(spec)
	}
	defer git.ResetMocks()

	dbStore := NewMockDBStore()
	dbStore.TransactFunc.SetDefaultReturn(dbStore, nil)
	dbStore.DoneFunc.SetDefaultHook(func(err error) error { return err })
	dbStore.StaleSourcedCommitsFunc.SetDefaultReturn(testSourcedCommits, nil)
	clock := glock.NewMockClock()
	janitor := newJanitor(dbStore, time.Minute, 100, newMetrics(&observation.TestContext), clock)

	if err := janitor.Handle(context.Background()); err != nil {
		t.Fatalf("unexpected error running janitor")
	}

	var sanitizedCalls []refreshCommitResolvabilityFuncInvocation
	for _, call := range dbStore.RefreshCommitResolvabilityFunc.History() {
		sanitizedCalls = append(sanitizedCalls, refreshCommitResolvabilityFuncInvocation{
			RepositoryID: call.Arg1,
			Commit:       call.Arg2,
			Delete:       call.Arg3,
		})
	}
	if diff := cmp.Diff(expectedCalls, sanitizedCalls); diff != "" {
		t.Errorf("unexpected calls to RefreshCommitResolvability (-want +got):\n%s", diff)
	}
}
