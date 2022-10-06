package uploads

import (
	"context"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/derision-test/glock"
	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/shared"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func TestUnknownCommitsJanitor(t *testing.T) {
	resolveRevisionFunc := func(commit string) error {
		return nil
	}

	testUnknownCommitsJanitor(t, resolveRevisionFunc, []updateInvocation{
		{1, "foo-x", false},
		{1, "foo-y", false},
		{1, "foo-z", false},
		{2, "bar-x", false},
		{2, "bar-y", false},
		{2, "bar-z", false},
		{3, "baz-x", false},
		{3, "baz-y", false},
		{3, "baz-z", false},
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
		{1, "foo-x", false},
		{1, "foo-y", true},
		{1, "foo-z", false},
		{2, "bar-x", true},
		{2, "bar-y", false},
		{2, "bar-z", false},
		{3, "baz-x", false},
		{3, "baz-y", false},
		{3, "baz-z", true},
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
		{1, "foo-x", false},
		{1, "foo-y", false},
		{1, "foo-z", false},
		{2, "bar-x", false},
		{2, "bar-y", false},
		{2, "bar-z", false},
		{3, "baz-x", false},
		{3, "baz-y", false},
		{3, "baz-z", false},
	})
}

type updateInvocation struct {
	RepositoryID int
	Commit       string
	Delete       bool
}

var testSourcedCommits = []shared.SourcedCommits{
	{RepositoryID: 1, RepositoryName: "foo", Commits: []string{"foo-x", "foo-y", "foo-z"}},
	{RepositoryID: 2, RepositoryName: "bar", Commits: []string{"bar-x", "bar-y", "bar-z"}},
	{RepositoryID: 3, RepositoryName: "baz", Commits: []string{"baz-x", "baz-y", "baz-z"}},
}

func testUnknownCommitsJanitor(t *testing.T, resolveRevisionFunc func(commit string) error, expectedCalls []updateInvocation) {
	gitserverClient := NewMockGitserverClient()
	gitserverClient.ResolveRevisionFunc.SetDefaultHook(func(ctx context.Context, i int, spec string) (api.CommitID, error) {
		return api.CommitID(spec), resolveRevisionFunc(spec)
	})

	store := NewMockStore()
	lsifStore := NewMockLsifStore()
	store.GetStaleSourcedCommitsFunc.SetDefaultReturn(testSourcedCommits, nil)
	autoIndexingSvc := NewMockAutoIndexingService()
	janitor := &Service{
		store:           store,
		lsifstore:       lsifStore,
		gitserverClient: gitserverClient,
		autoIndexingSvc: autoIndexingSvc,
		clock:           glock.NewRealClock(),
		logger:          logtest.Scoped(t),
		operations:      newOperations(&observation.TestContext),
		janitorMetrics:  newJanitorMetrics(&observation.TestContext),
	}

	if err := janitor.handleCleanup(
		context.Background(), janitorConfig{
			minimumTimeSinceLastCheck:      1 * time.Hour,
			commitResolverBatchSize:        10,
			auditLogMaxAge:                 1 * time.Hour,
			commitResolverMaximumCommitLag: 1 * time.Hour,
			uploadTimeout:                  1 * time.Hour,
		}); err != nil {
		t.Fatalf("unexpected error running janitor: %s", err)
	}

	var sanitizedCalls []updateInvocation
	for _, call := range store.UpdateSourcedCommitsFunc.History() {
		sanitizedCalls = append(sanitizedCalls, updateInvocation{
			RepositoryID: call.Arg1,
			Commit:       call.Arg2,
			Delete:       false,
		})
	}
	for _, call := range store.DeleteSourcedCommitsFunc.History() {
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
