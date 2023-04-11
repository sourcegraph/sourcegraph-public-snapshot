package janitor

import (
	"context"
	"strings"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
)

func TestShouldDeleteRecordsForCommit(t *testing.T) {
	resolveRevisionFunc := func(commit string) error {
		return nil
	}

	testShouldDeleteRecordsForCommit(t, resolveRevisionFunc, []updateInvocation{
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

func TestShouldDeleteRecordsForCommitUnknownCommit(t *testing.T) {
	resolveRevisionFunc := func(commit string) error {
		if commit == "foo-y" || commit == "bar-x" || commit == "baz-z" {
			return &gitdomain.RevisionNotFoundError{}
		}

		return nil
	}

	testShouldDeleteRecordsForCommit(t, resolveRevisionFunc, []updateInvocation{
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

func TestShouldDeleteRecordsForCommitUnknownRepository(t *testing.T) {
	resolveRevisionFunc := func(commit string) error {
		if strings.HasPrefix(commit, "foo-") {
			return &gitdomain.RepoNotExistError{}
		}

		return nil
	}

	testShouldDeleteRecordsForCommit(t, resolveRevisionFunc, []updateInvocation{
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

type sourcedCommits struct {
	RepositoryID   int
	RepositoryName string
	Commits        []string
}

var testSourcedCommits = []sourcedCommits{
	{RepositoryID: 1, RepositoryName: "foo", Commits: []string{"foo-x", "foo-y", "foo-z"}},
	{RepositoryID: 2, RepositoryName: "bar", Commits: []string{"bar-x", "bar-y", "bar-z"}},
	{RepositoryID: 3, RepositoryName: "baz", Commits: []string{"baz-x", "baz-y", "baz-z"}},
}

func testShouldDeleteRecordsForCommit(t *testing.T, resolveRevisionFunc func(commit string) error, expectedCalls []updateInvocation) {
	gitserverClient := gitserver.NewMockClient()
	gitserverClient.ResolveRevisionFunc.SetDefaultHook(func(ctx context.Context, _ api.RepoName, spec string, opts gitserver.ResolveRevisionOptions) (api.CommitID, error) {
		return api.CommitID(spec), resolveRevisionFunc(spec)
	})

	for _, sc := range testSourcedCommits {
		for _, commit := range sc.Commits {
			shouldDelete, err := shouldDeleteRecordsForCommit(context.Background(), gitserverClient, sc.RepositoryName, commit)
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
