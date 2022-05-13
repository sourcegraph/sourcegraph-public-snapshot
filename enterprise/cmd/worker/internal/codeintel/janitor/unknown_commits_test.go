package janitor

import (
	"context"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/derision-test/glock"
	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/stores/dbstore"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/vcs/git"
)

func TestUnknownCommitsJanitor(t *testing.T) {
	resolveRevisionFunc := func(commit string) error {
		return nil
	}

	testUnknownCommitsJanitor(t, resolveRevisionFunc, []updateInvocation{
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

	testUnknownCommitsJanitor(t, resolveRevisionFunc, []updateInvocation{
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

	testUnknownCommitsJanitor(t, resolveRevisionFunc, []updateInvocation{
		{1, "foo-x", false}, {1, "foo-y", false}, {1, "foo-z", false},
		{2, "bar-x", false}, {2, "bar-y", false}, {2, "bar-z", false},
		{3, "baz-x", false}, {3, "baz-y", false}, {3, "baz-z", false},
	})
}

type updateInvocation struct {
	RepositoryID int
	Commit       string
	Delete       bool
}

var testSourcedCommits = []dbstore.SourcedCommits{
	{RepositoryID: 1, RepositoryName: "foo", Commits: []string{"foo-x", "foo-y", "foo-z"}},
	{RepositoryID: 2, RepositoryName: "bar", Commits: []string{"bar-x", "bar-y", "bar-z"}},
	{RepositoryID: 3, RepositoryName: "baz", Commits: []string{"baz-x", "baz-y", "baz-z"}},
}

func testUnknownCommitsJanitor(t *testing.T, resolveRevisionFunc func(commit string) error, expectedCalls []updateInvocation) {
	gitserver.Mocks.ResolveRevision = func(spec string, opt gitserver.ResolveRevisionOptions) (api.CommitID, error) {
		return api.CommitID(spec), resolveRevisionFunc(spec)
	}
	defer git.ResetMocks()

	dbStore := NewMockDBStore()
	dbStore.TransactFunc.SetDefaultReturn(dbStore, nil)
	dbStore.DoneFunc.SetDefaultHook(func(err error) error { return err })
	dbStore.StaleSourcedCommitsFunc.SetDefaultReturn(testSourcedCommits, nil)
	clock := glock.NewMockClock()
	janitor := newJanitor(dbStore, time.Minute, 100, time.Minute, newMetrics(&observation.TestContext), clock)

	if err := janitor.Handle(context.Background()); err != nil {
		t.Fatalf("unexpected error running janitor: %s", err)
	}

	var sanitizedCalls []updateInvocation
	for _, call := range dbStore.UpdateSourcedCommitsFunc.History() {
		sanitizedCalls = append(sanitizedCalls, updateInvocation{
			RepositoryID: call.Arg1,
			Commit:       call.Arg2,
			Delete:       false,
		})
	}
	for _, call := range dbStore.DeleteSourcedCommitsFunc.History() {
		sanitizedCalls = append(sanitizedCalls, updateInvocation{
			RepositoryID: call.Arg1,
			Commit:       call.Arg2,
			Delete:       true,
		})
	}
	sort.Slice(sanitizedCalls, func(i, j int) bool {
		if sanitizedCalls[i].RepositoryID < sanitizedCalls[j].RepositoryID {
			return true
		}

		return sanitizedCalls[i].RepositoryID == sanitizedCalls[j].RepositoryID && sanitizedCalls[i].Commit < sanitizedCalls[j].Commit
	})

	if diff := cmp.Diff(expectedCalls, sanitizedCalls); diff != "" {
		t.Errorf("unexpected calls to UpdateSourcedCommits and DeleteSourcedCommits (-want +got):\n%s", diff)
	}
}
