package inttests

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	inttests "github.com/sourcegraph/sourcegraph/internal/gitserver/integration_tests"
)

func TestGetCommits(t *testing.T) {
	t.Parallel()
	inttests.InitGitserver()
	ctx := actor.WithActor(context.Background(), &actor.Actor{
		UID: 1,
	})

	repo1 := inttests.MakeGitRepository(t, getGitCommandsWithFiles("file1", "file2")...)
	repo2 := inttests.MakeGitRepository(t, getGitCommandsWithFiles("file3", "file4")...)
	repo3 := inttests.MakeGitRepository(t, getGitCommandsWithFiles("file5", "file6")...)

	repoCommits := []api.RepoCommit{
		{Repo: repo1, CommitID: api.CommitID("HEAD")},                                     // HEAD (file2)
		{Repo: repo1, CommitID: api.CommitID("HEAD~1")},                                   // HEAD~1 (file1)
		{Repo: repo2, CommitID: api.CommitID("67762ad757dd26cac4145f2b744fd93ad10a48e0")}, // HEAD (file4)
		{Repo: repo2, CommitID: api.CommitID("2b988222e844b570959a493f5b07ec020b89e122")}, // HEAD~1 (file3)
		{Repo: repo3, CommitID: api.CommitID("01bed0a")},                                  // abbrev HEAD (file6)
		{Repo: repo3, CommitID: api.CommitID("unresolvable")},                             // unresolvable
		{Repo: api.RepoName("unresolvable"), CommitID: api.CommitID("deadbeef")},          // unresolvable
	}

	t.Run("with sub-repo permissions", func(t *testing.T) {
		expectedCommits := []*gitdomain.Commit{
			{
				ID:        "2ba4dd2b9a27ec125fea7d72e12b9824ead18631",
				Author:    gitdomain.Signature{Name: "a", Email: "a@a.com", Date: *mustParseDate("2006-01-02T15:04:05Z", t)},
				Committer: &gitdomain.Signature{Name: "a", Email: "a@a.com", Date: *mustParseDate("2006-01-02T15:04:05Z", t)},
				Message:   "commit2",
				Parents:   []api.CommitID{"d38233a79e037d2ab8170b0d0bc0aa438473e6da"},
			},
			nil, // file 1
			{
				ID:        "67762ad757dd26cac4145f2b744fd93ad10a48e0",
				Author:    gitdomain.Signature{Name: "a", Email: "a@a.com", Date: *mustParseDate("2006-01-02T15:04:05Z", t)},
				Committer: &gitdomain.Signature{Name: "a", Email: "a@a.com", Date: *mustParseDate("2006-01-02T15:04:05Z", t)},
				Message:   "commit2",
				Parents:   []api.CommitID{"2b988222e844b570959a493f5b07ec020b89e122"},
			},
			nil, // file 3
			{
				ID:        "01bed0ae660668c57539cecaacb4c33d77609f43",
				Author:    gitdomain.Signature{Name: "a", Email: "a@a.com", Date: *mustParseDate("2006-01-02T15:04:05Z", t)},
				Committer: &gitdomain.Signature{Name: "a", Email: "a@a.com", Date: *mustParseDate("2006-01-02T15:04:05Z", t)},
				Message:   "commit2",
				Parents:   []api.CommitID{"d6ce2e76d171569d81c0afdc4573f461cec17d45"},
			},
			nil,
			nil,
		}

		commits, err := gitserver.NewTestClient(http.DefaultClient, inttests.GitserverAddresses).GetCommits(ctx, getTestSubRepoPermsChecker("file1", "file3"), repoCommits, true)
		if err != nil {
			t.Fatalf("unexpected error calling getCommits: %s", err)
		}
		if diff := cmp.Diff(expectedCommits, commits); diff != "" {
			t.Errorf("unexpected commits (-want +got):\n%s", diff)
		}
	})
}

// get a test sub-repo permissions checker which allows access to all files (so should be a no-op)
func getTestSubRepoPermsChecker(noAccessPaths ...string) authz.SubRepoPermissionChecker {
	checker := authz.NewMockSubRepoPermissionChecker()
	checker.EnabledFunc.SetDefaultHook(func() bool {
		return true
	})
	checker.PermissionsFunc.SetDefaultHook(func(ctx context.Context, i int32, content authz.RepoContent) (authz.Perms, error) {
		for _, noAccessPath := range noAccessPaths {
			if content.Path == noAccessPath {
				return authz.None, nil
			}
		}
		return authz.Read, nil
	})
	return checker
}

func getGitCommandsWithFiles(fileName1, fileName2 string) []string {
	return []string{
		fmt.Sprintf("touch %s", fileName1),
		fmt.Sprintf("git add %s", fileName1),
		"GIT_COMMITTER_NAME=a GIT_COMMITTER_EMAIL=a@a.com GIT_COMMITTER_DATE=2006-01-02T15:04:05Z git commit -m commit1 --author='a <a@a.com>' --date 2006-01-02T15:04:05Z",
		fmt.Sprintf("touch %s", fileName2),
		fmt.Sprintf("git add %s", fileName2),
		"GIT_COMMITTER_NAME=a GIT_COMMITTER_EMAIL=a@a.com GIT_COMMITTER_DATE=2006-01-02T15:04:05Z git commit -m commit2 --author='a <a@a.com>' --date 2006-01-02T15:04:05Z",
	}
}

func mustParseDate(s string, t *testing.T) *time.Time {
	t.Helper()
	date, err := time.Parse(time.RFC3339, s)
	if err != nil {
		t.Fatalf("unexpected error parsing date string: %s", err)
	}
	return &date
}

func TestHead(t *testing.T) {
	inttests.InitGitserver()
	client := gitserver.NewTestClient(http.DefaultClient, inttests.GitserverAddresses)

	t.Run("with sub-repo permissions", func(t *testing.T) {
		gitCommands := []string{
			"touch file",
			"git add file",
			"GIT_COMMITTER_NAME=a GIT_COMMITTER_EMAIL=a@a.com GIT_COMMITTER_DATE=2006-01-02T15:04:05Z git commit -m foo --author='a <a@a.com>' --date 2006-01-02T15:04:05Z",
		}
		repo := inttests.MakeGitRepository(t, gitCommands...)
		ctx := actor.WithActor(context.Background(), &actor.Actor{
			UID: 1,
		})
		checker := getTestSubRepoPermsChecker("file")
		// call Head() when user doesn't have access to view the commit
		_, exists, err := client.Head(ctx, checker, repo)
		if err != nil {
			t.Fatal(err)
		}
		if exists {
			t.Fatalf("exists should be false since the user doesn't have access to view the commit")
		}
		readAllChecker := getTestSubRepoPermsChecker()
		// call Head() when user has access to view the commit; should return expected commit
		head, exists, err := client.Head(ctx, readAllChecker, repo)
		if err != nil {
			t.Fatal(err)
		}
		wantHead := "46619ad353dbe4ed4108ebde9aa59ef676994a0b"
		if head != wantHead {
			t.Fatalf("Want %q, got %q", wantHead, head)
		}
		if !exists {
			t.Fatal("Should exist")
		}
	})
}
