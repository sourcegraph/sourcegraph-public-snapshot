package background

import (
	"context"
	"strings"
	"testing"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/autoindexing/shared"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
)

func TestShouldDeleteUploadsForCommit(t *testing.T) {
	resolveRevisionFunc := func(commit string) error {
		return nil
	}

	testShouldDeleteUploadsForCommit(t, resolveRevisionFunc, []updateInvocation{
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

func TestShouldDeleteUploadsForCommitUnknownCommit(t *testing.T) {
	resolveRevisionFunc := func(commit string) error {
		if commit == "foo-y" || commit == "bar-x" || commit == "baz-z" {
			return &gitdomain.RevisionNotFoundError{}
		}

		return nil
	}

	testShouldDeleteUploadsForCommit(t, resolveRevisionFunc, []updateInvocation{
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

func TestShouldDeleteUploadsForCommitUnknownRepository(t *testing.T) {
	resolveRevisionFunc := func(commit string) error {
		if strings.HasPrefix(commit, "foo-") {
			return &gitdomain.RepoNotExistError{}
		}

		return nil
	}

	testShouldDeleteUploadsForCommit(t, resolveRevisionFunc, []updateInvocation{
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

func testShouldDeleteUploadsForCommit(t *testing.T, resolveRevisionFunc func(commit string) error, expectedCalls []updateInvocation) {
	gitserverClient := NewMockGitserverClient()
	gitserverClient.ResolveRevisionFunc.SetDefaultHook(func(ctx context.Context, i int, spec string) (api.CommitID, error) {
		return api.CommitID(spec), resolveRevisionFunc(spec)
	})

	job := janitorJob{gitserverClient: gitserverClient}

	for _, sc := range testSourcedCommits {
		for _, commit := range sc.Commits {
			shouldDelete, err := job.shouldDeleteUploadsForCommit(context.Background(), sc.RepositoryID, commit)
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			for _, c := range expectedCalls {
				if c.RepositoryID == sc.RepositoryID && c.Commit == commit {
					if shouldDelete != c.Delete {
						t.Fatalf("unexpected result for %d@%s (want=%v have=%v)", sc.RepositoryID, commit, c.Delete, shouldDelete)
					}
				}
			}
		}
	}
}
