package git

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var (
	fileWithAccess    = "file-with-access"
	fileWithoutAccess = "file-without-access"
)

func TestLogPartsPerCommitInSync(t *testing.T) {
	require.Equal(t, 2*partsPerCommitBasic, strings.Count(logFormatWithoutRefs, "%"),
		"Expected (2 * %0d) %% signs in log format string (%0d fields, %0d %%x00 separators)",
		partsPerCommitBasic)
}

func TestRepository_GetCommit(t *testing.T) {
	ctx := actor.WithActor(context.Background(), &actor.Actor{
		UID: 1,
	})
	db := database.NewMockDB()
	gitCommands := []string{
		"GIT_COMMITTER_NAME=a GIT_COMMITTER_EMAIL=a@a.com GIT_COMMITTER_DATE=2006-01-02T15:04:05Z git commit --allow-empty -m foo --author='a <a@a.com>' --date 2006-01-02T15:04:05Z",
		"GIT_COMMITTER_NAME=c GIT_COMMITTER_EMAIL=c@c.com GIT_COMMITTER_DATE=2006-01-02T15:04:07Z git commit --allow-empty -m bar --author='a <a@a.com>' --date 2006-01-02T15:04:06Z",
	}
	gitCommandsWithFiles := getGitCommandsWithFiles(fileWithAccess, fileWithoutAccess)

	oldRunCommitLog := runCommitLog

	type testCase struct {
		repo                  api.RepoName
		id                    api.CommitID
		wantCommit            *gitdomain.Commit
		noEnsureRevision      bool
		revisionNotFoundError bool
	}

	runGetCommitTests := func(checker authz.SubRepoPermissionChecker, tests map[string]testCase) {
		for label, test := range tests {
			t.Run(label, func(t *testing.T) {
				var noEnsureRevision bool
				t.Cleanup(func() {
					runCommitLog = oldRunCommitLog
				})
				runCommitLog = func(ctx context.Context, cmd gitserver.GitCommand, opt CommitsOptions) ([]*wrappedCommit, error) {
					// Track the value of NoEnsureRevision we pass to gitserver
					noEnsureRevision = opt.NoEnsureRevision
					return oldRunCommitLog(ctx, cmd, opt)
				}

				resolveRevisionOptions := gitserver.ResolveRevisionOptions{
					NoEnsureRevision: test.noEnsureRevision,
				}
				commit, err := GetCommit(ctx, db, test.repo, test.id, resolveRevisionOptions, checker)
				if err != nil {
					if test.revisionNotFoundError {
						if !errors.HasType(err, &gitdomain.RevisionNotFoundError{}) {
							t.Errorf("%s: GetCommit: expected a RevisionNotFoundError, got %s", label, err)
						}
						return
					}
					t.Errorf("%s: GetCommit: %s", label, err)
				}

				if !CommitsEqual(commit, test.wantCommit) {
					t.Errorf("%s: got commit == %+v, want %+v", label, commit, test.wantCommit)
					return
				}

				// Test that trying to get a nonexistent commit returns RevisionNotFoundError.
				if _, err := GetCommit(ctx, db, test.repo, NonExistentCommitID, resolveRevisionOptions, checker); !errors.HasType(err, &gitdomain.RevisionNotFoundError{}) {
					t.Errorf("%s: for nonexistent commit: got err %v, want RevisionNotFoundError", label, err)
				}

				if noEnsureRevision != test.noEnsureRevision {
					t.Fatalf("Expected %t, got %t", test.noEnsureRevision, noEnsureRevision)
				}
			})
		}
	}

	wantGitCommit := &gitdomain.Commit{
		ID:        "b266c7e3ca00b1a17ad0b1449825d0854225c007",
		Author:    gitdomain.Signature{Name: "a", Email: "a@a.com", Date: MustParseTime(time.RFC3339, "2006-01-02T15:04:06Z")},
		Committer: &gitdomain.Signature{Name: "c", Email: "c@c.com", Date: MustParseTime(time.RFC3339, "2006-01-02T15:04:07Z")},
		Message:   "bar",
		Parents:   []api.CommitID{"ea167fe3d76b1e5fd3ed8ca44cbd2fe3897684f8"},
	}
	tests := map[string]testCase{
		"git cmd with NoEnsureRevision false": {
			repo:             MakeGitRepository(t, gitCommands...),
			id:               "b266c7e3ca00b1a17ad0b1449825d0854225c007",
			wantCommit:       wantGitCommit,
			noEnsureRevision: false,
		},
		"git cmd with NoEnsureRevision true": {
			repo:             MakeGitRepository(t, gitCommands...),
			id:               "b266c7e3ca00b1a17ad0b1449825d0854225c007",
			wantCommit:       wantGitCommit,
			noEnsureRevision: true,
		},
	}
	// Run basic tests w/o sub-repo permissions checker
	runGetCommitTests(nil, tests)
	checker := getTestSubRepoPermsChecker(fileWithoutAccess)
	// Add test cases with file names for sub-repo permissions testing
	tests["with sub-repo permissions and access to file"] = testCase{
		repo: MakeGitRepository(t, gitCommandsWithFiles...),
		id:   "da50eed82c8ff3c17bb642000d8aad9d434283c1",
		wantCommit: &gitdomain.Commit{
			ID:        "da50eed82c8ff3c17bb642000d8aad9d434283c1",
			Author:    gitdomain.Signature{Name: "a", Email: "a@a.com", Date: MustParseTime(time.RFC3339, "2006-01-02T15:04:05Z")},
			Committer: &gitdomain.Signature{Name: "a", Email: "a@a.com", Date: MustParseTime(time.RFC3339, "2006-01-02T15:04:05Z")},
			Message:   "commit1",
		},
		noEnsureRevision: true,
	}
	tests["with sub-repo permissions and NO access to file"] = testCase{
		repo:                  MakeGitRepository(t, gitCommandsWithFiles...),
		id:                    "ee7773505e98390e809cbf518b2a92e4748b0187",
		wantCommit:            &gitdomain.Commit{},
		noEnsureRevision:      true,
		revisionNotFoundError: true,
	}
	// Run test w/ sub-repo permissions filtering
	runGetCommitTests(checker, tests)
}

func TestRepository_HasCommitAfter(t *testing.T) {
	t.Parallel()
	ctx := actor.WithActor(context.Background(), &actor.Actor{
		UID: 1,
	})

	db := database.NewMockDB()

	testCases := []struct {
		label                 string
		commitDates           []string
		after                 string
		revspec               string
		want, wantSubRepoTest bool
	}{
		{
			label: "after specific date",
			commitDates: []string{
				"2006-01-02T15:04:05Z",
				"2007-01-02T15:04:05Z",
				"2008-01-02T15:04:05Z",
			},
			after:           "2006-01-02T15:04:05Z",
			revspec:         "master",
			want:            true,
			wantSubRepoTest: true,
		},
		{
			label: "after 1 year ago",
			commitDates: []string{
				"2016-01-02T15:04:05Z",
				"2017-01-02T15:04:05Z",
				"2017-01-02T15:04:06Z",
			},
			after:           "1 year ago",
			revspec:         "master",
			want:            false,
			wantSubRepoTest: false,
		},
		{
			label: "after too recent date",
			commitDates: []string{
				"2006-01-02T15:04:05Z",
				"2007-01-02T15:04:05Z",
				"2008-01-02T15:04:05Z",
			},
			after:           "2010-01-02T15:04:05Z",
			revspec:         "HEAD",
			want:            false,
			wantSubRepoTest: false,
		},
		{
			label: "commit 1 second after",
			commitDates: []string{
				"2006-01-02T15:04:05Z",
				"2007-01-02T15:04:05Z",
				"2007-01-02T15:04:06Z",
			},
			after:           "2007-01-02T15:04:05Z",
			revspec:         "HEAD",
			want:            true,
			wantSubRepoTest: false,
		},
		{
			label: "after 10 years ago",
			commitDates: []string{
				"2016-01-02T15:04:05Z",
				"2017-01-02T15:04:05Z",
				"2017-01-02T15:04:06Z",
			},
			after:           "10 years ago",
			revspec:         "HEAD",
			want:            true,
			wantSubRepoTest: true,
		},
	}

	t.Run("basic", func(t *testing.T) {
		for _, tc := range testCases {
			t.Run(tc.label, func(t *testing.T) {
				gitCommands := make([]string, len(tc.commitDates))
				for i, date := range tc.commitDates {
					gitCommands[i] = fmt.Sprintf("GIT_COMMITTER_NAME=a GIT_COMMITTER_EMAIL=a@a.com GIT_COMMITTER_DATE=%s git commit --allow-empty -m foo --author='a <a@a.com>'", date)
				}

				repo := MakeGitRepository(t, gitCommands...)
				got, err := HasCommitAfter(ctx, db, repo, tc.after, tc.revspec, nil)
				if err != nil || got != tc.want {
					t.Errorf("got %t hascommitafter, want %t", got, tc.want)
				}
			})
		}
	})

	t.Run("with sub-repo permissions", func(t *testing.T) {
		for _, tc := range testCases {
			t.Run(tc.label, func(t *testing.T) {
				gitCommands := make([]string, len(tc.commitDates))
				for i, date := range tc.commitDates {
					fileName := fmt.Sprintf("file%d", i)
					gitCommands = append(gitCommands, fmt.Sprintf("touch %s", fileName), fmt.Sprintf("git add %s", fileName))
					gitCommands = append(gitCommands, fmt.Sprintf("GIT_COMMITTER_NAME=a GIT_COMMITTER_EMAIL=a@a.com GIT_COMMITTER_DATE=%s git commit -m commit%d --author='a <a@a.com>'", date, i))
				}
				// Case where user can't view commit 2, but can view commits 0 and 1. In each test case the result should match the case where no sub-repo perms enabled
				checker := getTestSubRepoPermsChecker("file2")
				repo := MakeGitRepository(t, gitCommands...)
				got, err := HasCommitAfter(ctx, db, repo, tc.after, tc.revspec, checker)
				if err != nil {
					t.Errorf("got error: %s", err)
				}
				if got != tc.want {
					t.Errorf("got %t hascommitafter, want %t", got, tc.want)
				}

				// Case where user can't view commit 1 or commit 2, which will mean in some cases since HasCommitAfter will be false due to those commits not being visible.
				checker = getTestSubRepoPermsChecker("file1", "file2")
				repo = MakeGitRepository(t, gitCommands...)
				got, err = HasCommitAfter(ctx, db, repo, tc.after, tc.revspec, checker)
				if err != nil {
					t.Errorf("got error: %s", err)
				}
				if got != tc.wantSubRepoTest {
					t.Errorf("got %t hascommitafter, want %t", got, tc.wantSubRepoTest)
				}
			})
		}
	})
}

func TestRepository_FirstEverCommit(t *testing.T) {
	t.Parallel()
	ctx := actor.WithActor(context.Background(), &actor.Actor{
		UID: 1,
	})

	db := database.NewMockDB()

	testCases := []struct {
		commitDates []string
		want        string
	}{
		{
			commitDates: []string{
				"2006-01-02T15:04:05Z",
				"2007-01-02T15:04:05Z",
				"2008-01-02T15:04:05Z",
			},
			want: "2006-01-02T15:04:05Z",
		},
		{
			commitDates: []string{
				"2007-01-02T15:04:05Z", // Don't think this is possible, but if it is we still want the first commit (not strictly "oldest")
				"2006-01-02T15:04:05Z",
				"2007-01-02T15:04:06Z",
			},
			want: "2007-01-02T15:04:05Z",
		},
	}
	t.Run("basic", func(t *testing.T) {
		for _, tc := range testCases {
			gitCommands := make([]string, len(tc.commitDates))
			for i, date := range tc.commitDates {
				gitCommands[i] = fmt.Sprintf("GIT_COMMITTER_NAME=a GIT_COMMITTER_EMAIL=a@a.com GIT_COMMITTER_DATE=%s git commit --allow-empty -m foo --author='a <a@a.com>'", date)
			}

			repo := MakeGitRepository(t, gitCommands...)
			gotCommit, err := FirstEverCommit(ctx, db, repo, nil)
			if err != nil {
				t.Fatal(err)
			}
			got := gotCommit.Committer.Date.Format(time.RFC3339)
			if got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		}
	})

	t.Run("with sub-repo permissions", func(t *testing.T) {
		checkerWithoutAccessFirstCommit := getTestSubRepoPermsChecker("file0")
		checkerWithAccessFirstCommit := getTestSubRepoPermsChecker("file1")
		for _, tc := range testCases {
			gitCommands := make([]string, 0, len(tc.commitDates))
			for i, date := range tc.commitDates {
				fileName := fmt.Sprintf("file%d", i)
				gitCommands = append(gitCommands, fmt.Sprintf("touch %s", fileName))
				gitCommands = append(gitCommands, fmt.Sprintf("git add %s", fileName))
				gitCommands = append(gitCommands, fmt.Sprintf("GIT_COMMITTER_NAME=a GIT_COMMITTER_EMAIL=a@a.com GIT_COMMITTER_DATE=%s git commit -m foo --author='a <a@a.com>'", date))
			}

			repo := MakeGitRepository(t, gitCommands...)
			// Try to get first commit when user doesn't have permission to view
			_, err := FirstEverCommit(ctx, db, repo, checkerWithoutAccessFirstCommit)
			if !errors.HasType(err, &gitdomain.RevisionNotFoundError{}) {
				t.Errorf("expected a RevisionNotFoundError since the user does not have access to view this commit, got :%s", err)
			}
			// Try to get first commit when user does have permission to view, should succeed
			gotCommit, err := FirstEverCommit(ctx, db, repo, checkerWithAccessFirstCommit)
			if err != nil {
				t.Fatal(err)
			}
			got := gotCommit.Committer.Date.Format(time.RFC3339)
			if got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
			// Internal actor should always have access and ignore sub-repo permissions
			newCtx := actor.WithActor(context.Background(), &actor.Actor{
				UID:      1,
				Internal: true,
			})
			gotCommit, err = FirstEverCommit(newCtx, db, repo, checkerWithoutAccessFirstCommit)
			if err != nil {
				t.Fatal(err)
			}
			got = gotCommit.Committer.Date.Format(time.RFC3339)
			if got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		}
	})
}

func TestHead(t *testing.T) {
	t.Parallel()

	t.Run("basic", func(t *testing.T) {
		gitCommands := []string{
			"GIT_COMMITTER_NAME=a GIT_COMMITTER_EMAIL=a@a.com GIT_COMMITTER_DATE=2006-01-02T15:04:05Z git commit --allow-empty -m foo --author='a <a@a.com>' --date 2006-01-02T15:04:05Z",
		}
		repo := MakeGitRepository(t, gitCommands...)
		ctx := context.Background()

		head, exists, err := Head(ctx, database.NewMockDB(), repo, nil)
		if err != nil {
			t.Fatal(err)
		}
		wantHead := "ea167fe3d76b1e5fd3ed8ca44cbd2fe3897684f8"
		if head != wantHead {
			t.Fatalf("Want %q, got %q", wantHead, head)
		}
		if !exists {
			t.Fatal("Should exist")
		}
	})

	t.Run("with sub-repo permissions", func(t *testing.T) {
		db := database.NewMockDB()
		gitCommands := []string{
			"touch file",
			"git add file",
			"GIT_COMMITTER_NAME=a GIT_COMMITTER_EMAIL=a@a.com GIT_COMMITTER_DATE=2006-01-02T15:04:05Z git commit -m foo --author='a <a@a.com>' --date 2006-01-02T15:04:05Z",
		}
		repo := MakeGitRepository(t, gitCommands...)
		ctx := actor.WithActor(context.Background(), &actor.Actor{
			UID: 1,
		})
		checker := getTestSubRepoPermsChecker("file")
		// call Head() when user doesn't have access to view the commit
		_, exists, err := Head(ctx, db, repo, checker)
		if err != nil {
			t.Fatal(err)
		}
		if exists {
			t.Fatalf("exists should be false since the user doesn't have access to view the commit")
		}
		readAllChecker := getTestSubRepoPermsChecker()
		// call Head() when user has access to view the commit; should return expected commit
		head, exists, err := Head(ctx, db, repo, readAllChecker)
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

func TestCommitExists(t *testing.T) {
	t.Parallel()
	ctx := actor.WithActor(context.Background(), &actor.Actor{
		UID: 1,
	})
	db := database.NewMockDB()
	testCommitExists := func(label string, gitCommands []string, commitID, nonExistentCommitID api.CommitID, checker authz.SubRepoPermissionChecker) {
		t.Run(label, func(t *testing.T) {
			repo := MakeGitRepository(t, gitCommands...)

			exists, err := CommitExists(ctx, db, repo, commitID, checker)
			if err != nil {
				t.Fatal(err)
			}
			if !exists {
				t.Fatal("Should exist")
			}

			exists, err = CommitExists(ctx, db, repo, nonExistentCommitID, checker)
			if err != nil {
				t.Fatal(err)
			}
			if exists {
				t.Fatal("Should not exist")
			}
		})
	}

	gitCommands := []string{
		"GIT_COMMITTER_NAME=a GIT_COMMITTER_EMAIL=a@a.com GIT_COMMITTER_DATE=2006-01-02T15:04:05Z git commit --allow-empty -m foo --author='a <a@a.com>' --date 2006-01-02T15:04:05Z",
	}
	testCommitExists("basic", gitCommands, "ea167fe3d76b1e5fd3ed8ca44cbd2fe3897684f8", NonExistentCommitID, nil)
	gitCommandsWithFiles := getGitCommandsWithFiles(fileWithAccess, fileWithoutAccess)
	commitIDWithAccess := api.CommitID("da50eed82c8ff3c17bb642000d8aad9d434283c1")
	commitIDWithoutAccess := api.CommitID("ee7773505e98390e809cbf518b2a92e4748b0187")
	// Test that the commit ID the user has access to exists, and CommitExists returns false for the commit ID the user
	// doesn't have access to (since a file was modified in the commit that the user doesn't have permissions to view)
	testCommitExists("with sub-repo permissions filtering", gitCommandsWithFiles, commitIDWithAccess, commitIDWithoutAccess, getTestSubRepoPermsChecker(fileWithoutAccess))
}

func TestRepository_Commits(t *testing.T) {
	t.Parallel()
	ctx := actor.WithActor(context.Background(), &actor.Actor{
		UID: 1,
	})

	// TODO(sqs): test CommitsOptions.Base

	gitCommands := []string{
		"GIT_COMMITTER_NAME=a GIT_COMMITTER_EMAIL=a@a.com GIT_COMMITTER_DATE=2006-01-02T15:04:05Z git commit --allow-empty -m foo --author='a <a@a.com>' --date 2006-01-02T15:04:05Z",
		"GIT_COMMITTER_NAME=c GIT_COMMITTER_EMAIL=c@c.com GIT_COMMITTER_DATE=2006-01-02T15:04:07Z git commit --allow-empty -m bar --author='a <a@a.com>' --date 2006-01-02T15:04:06Z",
	}
	wantGitCommits := []*gitdomain.Commit{
		{
			ID:        "b266c7e3ca00b1a17ad0b1449825d0854225c007",
			Author:    gitdomain.Signature{Name: "a", Email: "a@a.com", Date: MustParseTime(time.RFC3339, "2006-01-02T15:04:06Z")},
			Committer: &gitdomain.Signature{Name: "c", Email: "c@c.com", Date: MustParseTime(time.RFC3339, "2006-01-02T15:04:07Z")},
			Message:   "bar",
			Parents:   []api.CommitID{"ea167fe3d76b1e5fd3ed8ca44cbd2fe3897684f8"},
		},
		{
			ID:        "ea167fe3d76b1e5fd3ed8ca44cbd2fe3897684f8",
			Author:    gitdomain.Signature{Name: "a", Email: "a@a.com", Date: MustParseTime(time.RFC3339, "2006-01-02T15:04:05Z")},
			Committer: &gitdomain.Signature{Name: "a", Email: "a@a.com", Date: MustParseTime(time.RFC3339, "2006-01-02T15:04:05Z")},
			Message:   "foo",
			Parents:   nil,
		},
	}
	tests := map[string]struct {
		repo        api.RepoName
		id          api.CommitID
		wantCommits []*gitdomain.Commit
		wantTotal   uint
	}{
		"git cmd": {
			repo:        MakeGitRepository(t, gitCommands...),
			id:          "b266c7e3ca00b1a17ad0b1449825d0854225c007",
			wantCommits: wantGitCommits,
			wantTotal:   2,
		},
	}
	runCommitsTests := func(checker authz.SubRepoPermissionChecker) {
		for label, test := range tests {
			t.Run(label, func(t *testing.T) {
				testCommits(ctx, label, test.repo, CommitsOptions{Range: string(test.id)}, checker, test.wantTotal, test.wantCommits, t)

				// Test that trying to get a nonexistent commit returns RevisionNotFoundError.
				if _, err := Commits(ctx, database.NewMockDB(), test.repo, CommitsOptions{Range: string(NonExistentCommitID)}, nil); !errors.HasType(err, &gitdomain.RevisionNotFoundError{}) {
					t.Errorf("%s: for nonexistent commit: got err %v, want RevisionNotFoundError", label, err)
				}
			})
		}
	}
	runCommitsTests(nil)
	checker := getTestSubRepoPermsChecker()
	runCommitsTests(checker)
}

func TestCommits_SubRepoPerms(t *testing.T) {
	t.Parallel()
	ctx := actor.WithActor(context.Background(), &actor.Actor{
		UID: 1,
	})
	gitCommands := []string{
		"touch file1",
		"git add file1",
		"GIT_COMMITTER_NAME=a GIT_COMMITTER_EMAIL=a@a.com GIT_COMMITTER_DATE=2006-01-02T15:04:05Z git commit -m commit1 --author='a <a@a.com>' --date 2006-01-02T15:04:05Z",
		"touch file2",
		"git add file2",
		"touch file2.2",
		"git add file2.2",
		"GIT_COMMITTER_NAME=c GIT_COMMITTER_EMAIL=c@c.com GIT_COMMITTER_DATE=2006-01-02T15:04:07Z git commit -m commit2 --author='a <a@a.com>' --date 2006-01-02T15:04:06Z",
		"touch file3",
		"git add file3",
		"GIT_COMMITTER_NAME=c GIT_COMMITTER_EMAIL=c@c.com GIT_COMMITTER_DATE=2006-01-02T15:04:07Z git commit -m commit3 --author='a <a@a.com>' --date 2006-01-02T15:04:07Z",
	}

	tests := map[string]struct {
		repo          api.RepoName
		wantCommits   []*gitdomain.Commit
		opt           CommitsOptions
		wantTotal     uint
		noAccessPaths []string
	}{
		"if no read perms on file should filter out commit": {
			repo:      MakeGitRepository(t, gitCommands...),
			wantTotal: 1,
			wantCommits: []*gitdomain.Commit{
				{
					ID:        "d38233a79e037d2ab8170b0d0bc0aa438473e6da",
					Author:    gitdomain.Signature{Name: "a", Email: "a@a.com", Date: MustParseTime(time.RFC3339, "2006-01-02T15:04:05Z")},
					Committer: &gitdomain.Signature{Name: "a", Email: "a@a.com", Date: MustParseTime(time.RFC3339, "2006-01-02T15:04:05Z")},
					Message:   "commit1",
				},
			},
			noAccessPaths: []string{"file2", "file3"},
		},
		"sub-repo perms with path (w/ no access) specified should return no commits": {
			repo:      MakeGitRepository(t, gitCommands...),
			wantTotal: 1,
			opt: CommitsOptions{
				Path: "file2",
			},
			wantCommits:   []*gitdomain.Commit{},
			noAccessPaths: []string{"file2", "file3"},
		},
		"sub-repo perms with path (w/ access) specified should return that commit": {
			repo:      MakeGitRepository(t, gitCommands...),
			wantTotal: 1,
			opt: CommitsOptions{
				Path: "file1",
			},
			wantCommits: []*gitdomain.Commit{
				{
					ID:        "d38233a79e037d2ab8170b0d0bc0aa438473e6da",
					Author:    gitdomain.Signature{Name: "a", Email: "a@a.com", Date: MustParseTime(time.RFC3339, "2006-01-02T15:04:05Z")},
					Committer: &gitdomain.Signature{Name: "a", Email: "a@a.com", Date: MustParseTime(time.RFC3339, "2006-01-02T15:04:05Z")},
					Message:   "commit1",
				},
			},
			noAccessPaths: []string{"file2", "file3"},
		},
	}

	for label, test := range tests {
		t.Run(label, func(t *testing.T) {
			checker := getTestSubRepoPermsChecker(test.noAccessPaths...)
			commits, err := Commits(ctx, database.NewMockDB(), test.repo, test.opt, checker)
			if err != nil {
				t.Errorf("%s: Commits(): %s", label, err)
				return
			}

			if len(commits) != len(test.wantCommits) {
				t.Errorf("%s: got %d commits, want %d", label, len(commits), len(test.wantCommits))
			}

			checkCommits(t, label, commits, test.wantCommits)
		})
	}
}

func TestCommits_SubRepoPerms_ReturnNCommits(t *testing.T) {
	t.Parallel()
	ctx := actor.WithActor(context.Background(), &actor.Actor{
		UID: 1,
	})
	gitCommands := []string{
		"touch file1",
		"git add file1",
		"GIT_COMMITTER_NAME=c GIT_COMMITTER_EMAIL=c@c.com GIT_COMMITTER_DATE=2006-01-02T15:04:01Z git commit -m commit1 --author='a <a@a.com>' --date 2006-01-02T15:04:01Z",
		"touch file2",
		"git add file2",
		"GIT_COMMITTER_NAME=c GIT_COMMITTER_EMAIL=c@c.com GIT_COMMITTER_DATE=2006-01-02T15:04:02Z git commit -m commit2 --author='a <a@a.com>' --date 2006-01-02T15:04:02Z",
		"echo foo > file1",
		"git add file1",
		"GIT_COMMITTER_NAME=c GIT_COMMITTER_EMAIL=c@c.com GIT_COMMITTER_DATE=2006-01-02T15:04:03Z git commit -m commit3 --author='a <a@a.com>' --date 2006-01-02T15:04:03Z",
		"echo asdf > file1",
		"git add file1",
		"GIT_COMMITTER_NAME=c GIT_COMMITTER_EMAIL=c@c.com GIT_COMMITTER_DATE=2006-01-02T15:04:04Z git commit -m commit4 --author='a <a@a.com>' --date 2006-01-02T15:04:04Z",
		"echo bar > file1",
		"git add file1",
		"GIT_COMMITTER_NAME=c GIT_COMMITTER_EMAIL=c@c.com GIT_COMMITTER_DATE=2006-01-02T15:04:05Z git commit -m commit5 --author='a <a@a.com>' --date 2006-01-02T15:04:05Z",
		"echo asdf2 > file2",
		"git add file2",
		"GIT_COMMITTER_NAME=c GIT_COMMITTER_EMAIL=c@c.com GIT_COMMITTER_DATE=2006-01-02T15:04:06Z git commit -m commit6 --author='a <a@a.com>' --date 2006-01-02T15:04:06Z",
		"echo bazz > file1",
		"git add file1",
		"GIT_COMMITTER_NAME=c GIT_COMMITTER_EMAIL=c@c.com GIT_COMMITTER_DATE=2006-01-02T15:04:07Z git commit -m commit7 --author='a <a@a.com>' --date 2006-01-02T15:04:07Z",
		"echo bazz > file2",
		"git add file2",
		"GIT_COMMITTER_NAME=c GIT_COMMITTER_EMAIL=c@c.com GIT_COMMITTER_DATE=2006-01-02T15:04:08Z git commit -m commit8 --author='a <a@a.com>' --date 2006-01-02T15:04:08Z",
	}

	tests := map[string]struct {
		repo          api.RepoName
		wantCommits   []*gitdomain.Commit
		opt           CommitsOptions
		wantTotal     uint
		noAccessPaths []string
	}{
		"return the requested number of commits": {
			repo:      MakeGitRepository(t, gitCommands...),
			wantTotal: 3,
			opt: CommitsOptions{
				N: 3,
			},
			wantCommits: []*gitdomain.Commit{
				{
					ID:        "61dbc35f719c53810904a2d359309d4e1e98a6be",
					Author:    gitdomain.Signature{Name: "a", Email: "a@a.com", Date: MustParseTime(time.RFC3339, "2006-01-02T15:04:07Z")},
					Committer: &gitdomain.Signature{Name: "c", Email: "c@c.com", Date: MustParseTime(time.RFC3339, "2006-01-02T15:04:07Z")},
					Message:   "commit7",
					Parents:   []api.CommitID{"66566c8aa223f3e1b94ebe09e6cdb14c3a5bfb36"},
				},
				{
					ID:        "2e6b2c94293e9e339f781b2a2f7172e15460f88c",
					Author:    gitdomain.Signature{Name: "a", Email: "a@a.com", Date: MustParseTime(time.RFC3339, "2006-01-02T15:04:05Z")},
					Committer: &gitdomain.Signature{Name: "c", Email: "c@c.com", Date: MustParseTime(time.RFC3339, "2006-01-02T15:04:05Z")},
					Parents: []api.CommitID{
						"9a7ec70986d657c4c86d6ac476f0c5181ece509a",
					},
					Message: "commit5",
				},
				{
					ID:        "9a7ec70986d657c4c86d6ac476f0c5181ece509a",
					Author:    gitdomain.Signature{Name: "a", Email: "a@a.com", Date: MustParseTime(time.RFC3339, "2006-01-02T15:04:04Z")},
					Committer: &gitdomain.Signature{Name: "c", Email: "c@c.com", Date: MustParseTime(time.RFC3339, "2006-01-02T15:04:04Z")},
					Message:   "commit4",
					Parents: []api.CommitID{
						"f3fa8cf6ec56d0469402523385d6ca4b7cb222d8",
					},
				},
			},
			noAccessPaths: []string{"file2"},
		},
	}

	for label, test := range tests {
		t.Run(label, func(t *testing.T) {
			checker := getTestSubRepoPermsChecker(test.noAccessPaths...)
			commits, err := Commits(ctx, database.NewMockDB(), test.repo, test.opt, checker)
			if err != nil {
				t.Errorf("%s: Commits(): %s", label, err)
				return
			}

			if diff := cmp.Diff(test.wantCommits, commits); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}

func TestRepository_Commits_options(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	gitCommands := []string{
		"GIT_COMMITTER_NAME=a GIT_COMMITTER_EMAIL=a@a.com GIT_COMMITTER_DATE=2006-01-02T15:04:05Z git commit --allow-empty -m foo --author='a <a@a.com>' --date 2006-01-02T15:04:05Z",
		"GIT_COMMITTER_NAME=c GIT_COMMITTER_EMAIL=c@c.com GIT_COMMITTER_DATE=2006-01-02T15:04:07Z git commit --allow-empty -m bar --author='a <a@a.com>' --date 2006-01-02T15:04:06Z",
		"GIT_COMMITTER_NAME=c GIT_COMMITTER_EMAIL=c@c.com GIT_COMMITTER_DATE=2006-01-02T15:04:08Z git commit --allow-empty -m qux --author='a <a@a.com>' --date 2006-01-02T15:04:08Z",
	}
	wantGitCommits := []*gitdomain.Commit{
		{
			ID:        "b266c7e3ca00b1a17ad0b1449825d0854225c007",
			Author:    gitdomain.Signature{Name: "a", Email: "a@a.com", Date: MustParseTime(time.RFC3339, "2006-01-02T15:04:06Z")},
			Committer: &gitdomain.Signature{Name: "c", Email: "c@c.com", Date: MustParseTime(time.RFC3339, "2006-01-02T15:04:07Z")},
			Message:   "bar",
			Parents:   []api.CommitID{"ea167fe3d76b1e5fd3ed8ca44cbd2fe3897684f8"},
		},
	}
	wantGitCommits2 := []*gitdomain.Commit{
		{
			ID:        "ade564eba4cf904492fb56dcd287ac633e6e082c",
			Author:    gitdomain.Signature{Name: "a", Email: "a@a.com", Date: MustParseTime(time.RFC3339, "2006-01-02T15:04:08Z")},
			Committer: &gitdomain.Signature{Name: "c", Email: "c@c.com", Date: MustParseTime(time.RFC3339, "2006-01-02T15:04:08Z")},
			Message:   "qux",
			Parents:   []api.CommitID{"b266c7e3ca00b1a17ad0b1449825d0854225c007"},
		},
	}
	tests := map[string]struct {
		repo        api.RepoName
		opt         CommitsOptions
		wantCommits []*gitdomain.Commit
		wantTotal   uint
	}{
		"git cmd": {
			repo:        MakeGitRepository(t, gitCommands...),
			opt:         CommitsOptions{Range: "ade564eba4cf904492fb56dcd287ac633e6e082c", N: 1, Skip: 1},
			wantCommits: wantGitCommits,
			wantTotal:   1,
		},
		"git cmd Head": {
			repo: MakeGitRepository(t, gitCommands...),
			opt: CommitsOptions{
				Range: "b266c7e3ca00b1a17ad0b1449825d0854225c007...ade564eba4cf904492fb56dcd287ac633e6e082c",
			},
			wantCommits: wantGitCommits2,
			wantTotal:   1,
		},
		"before": {
			repo: MakeGitRepository(t, gitCommands...),
			opt: CommitsOptions{
				Before: "2006-01-02T15:04:07Z",
				Range:  "HEAD",
				N:      1,
			},
			wantCommits: []*gitdomain.Commit{
				{
					ID:        "b266c7e3ca00b1a17ad0b1449825d0854225c007",
					Author:    gitdomain.Signature{Name: "a", Email: "a@a.com", Date: MustParseTime(time.RFC3339, "2006-01-02T15:04:06Z")},
					Committer: &gitdomain.Signature{Name: "c", Email: "c@c.com", Date: MustParseTime(time.RFC3339, "2006-01-02T15:04:07Z")},
					Message:   "bar",
					Parents:   []api.CommitID{"ea167fe3d76b1e5fd3ed8ca44cbd2fe3897684f8"},
				},
			},
			wantTotal: 1,
		},
	}
	runCommitsTests := func(checker authz.SubRepoPermissionChecker) {
		for label, test := range tests {
			t.Run(label, func(t *testing.T) {
				testCommits(ctx, label, test.repo, test.opt, checker, test.wantTotal, test.wantCommits, t)
			})
		}
	}
	runCommitsTests(nil)
	checker := getTestSubRepoPermsChecker()
	runCommitsTests(checker)
}

func TestRepository_Commits_options_path(t *testing.T) {
	t.Parallel()
	ctx := actor.WithActor(context.Background(), &actor.Actor{
		UID: 1,
	})

	gitCommands := []string{
		"GIT_COMMITTER_NAME=a GIT_COMMITTER_EMAIL=a@a.com GIT_COMMITTER_DATE=2006-01-02T15:04:05Z git commit --allow-empty -m commit1 --author='a <a@a.com>' --date 2006-01-02T15:04:05Z",
		"touch file1",
		"touch --date=2006-01-02T15:04:05Z file1 || touch -t " + Times[0] + " file1",
		"git add file1",
		"GIT_COMMITTER_NAME=a GIT_COMMITTER_EMAIL=a@a.com GIT_COMMITTER_DATE=2006-01-02T15:04:05Z git commit -m commit2 --author='a <a@a.com>' --date 2006-01-02T15:04:05Z",
		"GIT_COMMITTER_NAME=c GIT_COMMITTER_EMAIL=c@c.com GIT_COMMITTER_DATE=2006-01-02T15:04:07Z git commit --allow-empty -m commit3 --author='a <a@a.com>' --date 2006-01-02T15:04:06Z",
	}
	wantGitCommits := []*gitdomain.Commit{
		{
			ID:        "546a3ef26e581624ef997cb8c0ba01ee475fc1dc",
			Author:    gitdomain.Signature{Name: "a", Email: "a@a.com", Date: MustParseTime(time.RFC3339, "2006-01-02T15:04:05Z")},
			Committer: &gitdomain.Signature{Name: "a", Email: "a@a.com", Date: MustParseTime(time.RFC3339, "2006-01-02T15:04:05Z")},
			Message:   "commit2",
			Parents:   []api.CommitID{"a04652fa1998a0a7d2f2f77ecb7021de943d3aab"},
		},
	}
	tests := map[string]struct {
		repo        api.RepoName
		opt         CommitsOptions
		wantCommits []*gitdomain.Commit
		wantTotal   uint
	}{
		"git cmd Path 0": {
			repo: MakeGitRepository(t, gitCommands...),
			opt: CommitsOptions{
				Range: "master",
				Path:  "doesnt-exist",
			},
			wantCommits: nil,
			wantTotal:   0,
		},
		"git cmd Path 1": {
			repo: MakeGitRepository(t, gitCommands...),
			opt: CommitsOptions{
				Range: "master",
				Path:  "file1",
			},
			wantCommits: wantGitCommits,
			wantTotal:   1,
		},
	}

	runCommitsTest := func(checker authz.SubRepoPermissionChecker) {
		for label, test := range tests {
			t.Run(label, func(t *testing.T) {
				testCommits(ctx, label, test.repo, test.opt, checker, test.wantTotal, test.wantCommits, t)
			})
		}
	}
	runCommitsTest(nil)
	checker := getTestSubRepoPermsChecker()
	runCommitsTest(checker)
}

func TestMessage(t *testing.T) {
	t.Run("Body", func(t *testing.T) {
		tests := map[gitdomain.Message]string{
			"hello":                 "",
			"hello\n":               "",
			"hello\n\n":             "",
			"hello\nworld":          "world",
			"hello\n\nworld":        "world",
			"hello\n\nworld\nfoo":   "world\nfoo",
			"hello\n\nworld\nfoo\n": "world\nfoo",
		}
		for input, want := range tests {
			got := input.Body()
			if got != want {
				t.Errorf("got %q, want %q", got, want)
			}
		}
	})
}

func TestParseCommitsUniqueToBranch(t *testing.T) {
	commits, err := parseCommitsUniqueToBranch([]string{
		"c165bfff52e9d4f87891bba497e3b70fea144d89:2020-08-04T08:23:30-05:00",
		"f73ee8ed601efea74f3b734eeb073307e1615606:2020-04-16T16:06:21-04:00",
		"6057f7ed8d331c82030c713b650fc8fd2c0c2347:2020-04-16T16:20:26-04:00",
		"7886287b8758d1baf19cf7b8253856128369a2a7:2020-04-16T16:55:58-04:00",
		"b69f89473bbcc04dc52cafaf6baa504e34791f5a:2020-04-20T12:10:49-04:00",
		"172b7fcf8b8c49b37b231693433586c2bfd1619e:2020-04-20T12:37:36-04:00",
		"5bc35c78fb5fb388891ca944cd12d85fd6dede95:2020-05-05T12:53:18-05:00",
	})
	if err != nil {
		t.Fatalf("unexpected error parsing commits: %s", err)
	}

	expectedCommits := map[string]time.Time{
		"c165bfff52e9d4f87891bba497e3b70fea144d89": *mustParseDate("2020-08-04T08:23:30-05:00", t),
		"f73ee8ed601efea74f3b734eeb073307e1615606": *mustParseDate("2020-04-16T16:06:21-04:00", t),
		"6057f7ed8d331c82030c713b650fc8fd2c0c2347": *mustParseDate("2020-04-16T16:20:26-04:00", t),
		"7886287b8758d1baf19cf7b8253856128369a2a7": *mustParseDate("2020-04-16T16:55:58-04:00", t),
		"b69f89473bbcc04dc52cafaf6baa504e34791f5a": *mustParseDate("2020-04-20T12:10:49-04:00", t),
		"172b7fcf8b8c49b37b231693433586c2bfd1619e": *mustParseDate("2020-04-20T12:37:36-04:00", t),
		"5bc35c78fb5fb388891ca944cd12d85fd6dede95": *mustParseDate("2020-05-05T12:53:18-05:00", t),
	}
	if diff := cmp.Diff(expectedCommits, commits); diff != "" {
		t.Errorf("unexpected commits (-want +got):\n%s", diff)
	}
}

func TestParseBranchesContaining(t *testing.T) {
	names := parseBranchesContaining([]string{
		"refs/tags/v0.7.0",
		"refs/tags/v0.5.1",
		"refs/tags/v1.1.4",
		"refs/heads/symbols", "refs/heads/bl/symbols",
		"refs/tags/v1.2.0",
		"refs/tags/v1.1.0",
		"refs/tags/v0.10.0",
		"refs/tags/v1.0.0",
		"refs/heads/garo/index-specific-files",
		"refs/heads/bl/symbols-2",
		"refs/tags/v1.3.1",
		"refs/tags/v0.5.2",
		"refs/tags/v1.1.2",
		"refs/tags/v0.8.0",
		"refs/heads/ef/wtf",
		"refs/tags/v1.5.0",
		"refs/tags/v0.9.0",
		"refs/heads/garo/go-and-typescript-lsif-indexing",
		"refs/heads/master",
		"refs/heads/sg/document-symbols",
		"refs/tags/v1.1.1",
		"refs/tags/v1.4.0",
		"refs/heads/nsc/bump-go-version",
		"refs/heads/nsc/random",
		"refs/heads/nsc/markupcontent",
		"refs/tags/v0.6.0",
		"refs/tags/v1.1.3",
		"refs/tags/v0.5.3",
		"refs/tags/v1.3.0",
	})

	expectedNames := []string{
		"bl/symbols",
		"bl/symbols-2",
		"ef/wtf",
		"garo/go-and-typescript-lsif-indexing",
		"garo/index-specific-files",
		"master",
		"nsc/bump-go-version",
		"nsc/markupcontent",
		"nsc/random",
		"sg/document-symbols",
		"symbols",
		"v0.10.0",
		"v0.5.1",
		"v0.5.2",
		"v0.5.3",
		"v0.6.0",
		"v0.7.0",
		"v0.8.0",
		"v0.9.0",
		"v1.0.0",
		"v1.1.0",
		"v1.1.1",
		"v1.1.2",
		"v1.1.3",
		"v1.1.4",
		"v1.2.0",
		"v1.3.0",
		"v1.3.1",
		"v1.4.0",
		"v1.5.0",
	}
	if diff := cmp.Diff(expectedNames, names); diff != "" {
		t.Errorf("unexpected names (-want +got):\n%s", diff)
	}
}

func TestParseRefDescriptions(t *testing.T) {
	refDescriptions, err := parseRefDescriptions(bytes.Join([][]byte{
		[]byte("66a7ac584740245fc523da443a3f540a52f8af72\x00refs/heads/bl/symbols\x00 \x002021-01-18T16:46:51-08:00"),
		[]byte("58537c06cf7ba8a562a3f5208fb7a8efbc971d0e\x00refs/heads/bl/symbols-2\x00 \x002021-02-24T06:21:20-08:00"),
		[]byte("a40716031ae97ee7c5cdf1dec913567a4a7c50c8\x00refs/heads/ef/wtf\x00 \x002021-02-10T10:50:08-06:00"),
		[]byte("e2e283fdaf6ea4a419cdbad142bbfd4b730080f8\x00refs/heads/garo/go-and-typescript-lsif-indexing\x00 \x002020-04-29T16:45:46+00:00"),
		[]byte("c485d92c3d2065041bf29b3fe0b55ffac7e66b2a\x00refs/heads/garo/index-specific-files\x00 \x002021-03-01T13:09:42-08:00"),
		[]byte("ce30aee6cc56f39d0ac6fee03c4c151c08a8cd2e\x00refs/heads/master\x00*\x002021-06-16T11:51:09-07:00"),
		[]byte("ec5cfc8ab33370c698273b1a097af73ea289c92b\x00refs/heads/nsc/bump-go-version\x00 \x002021-03-12T22:33:17+00:00"),
		[]byte("22b2c4f734f62060cae69da856fe3854defdcc87\x00refs/heads/nsc/markupcontent\x00 \x002021-05-03T23:50:02+01:00"),
		[]byte("9df3358a18792fa9dbd40d506f2e0ad23fc11ee8\x00refs/heads/nsc/random\x00 \x002021-02-10T16:29:06+00:00"),
		[]byte("a02b85b63345a1406d7a19727f7a5472c976e053\x00refs/heads/sg/document-symbols\x00 \x002021-04-08T15:33:03-07:00"),
		[]byte("234b0a484519129b251164ecb0674ec27d154d2f\x00refs/heads/symbols\x00 \x002021-01-01T22:51:55-08:00"),
		[]byte("6b5ae2e0ce568a7641174072271d109d7d0977c7\x00refs/tags/v0.0.0\x00 \x00"),
		[]byte("c165bfff52e9d4f87891bba497e3b70fea144d89\x00refs/tags/v0.10.0\x00 \x002020-08-04T08:23:30-05:00"),
		[]byte("f73ee8ed601efea74f3b734eeb073307e1615606\x00refs/tags/v0.5.1\x00 \x002020-04-16T16:06:21-04:00"),
		[]byte("6057f7ed8d331c82030c713b650fc8fd2c0c2347\x00refs/tags/v0.5.2\x00 \x002020-04-16T16:20:26-04:00"),
		[]byte("7886287b8758d1baf19cf7b8253856128369a2a7\x00refs/tags/v0.5.3\x00 \x002020-04-16T16:55:58-04:00"),
		[]byte("b69f89473bbcc04dc52cafaf6baa504e34791f5a\x00refs/tags/v0.6.0\x00 \x002020-04-20T12:10:49-04:00"),
		[]byte("172b7fcf8b8c49b37b231693433586c2bfd1619e\x00refs/tags/v0.7.0\x00 \x002020-04-20T12:37:36-04:00"),
		[]byte("5bc35c78fb5fb388891ca944cd12d85fd6dede95\x00refs/tags/v0.8.0\x00 \x002020-05-05T12:53:18-05:00"),
		[]byte("14faa49ef098df9488536ca3c9b26d79e6bec4d6\x00refs/tags/v0.9.0\x00 \x002020-07-14T14:26:40-05:00"),
		[]byte("0a82af8b6914d8c81326eee5f3a7e1d1106547f1\x00refs/tags/v1.0.0\x00 \x002020-08-19T19:33:39-05:00"),
		[]byte("262defb72b96261a7d56b000d438c5c7ec6d0f3e\x00refs/tags/v1.1.0\x00 \x002020-08-21T14:15:44-05:00"),
		[]byte("806b96eb544e7e632a617c26402eccee6d67faed\x00refs/tags/v1.1.1\x00 \x002020-08-21T16:02:35-05:00"),
		[]byte("5d8865d6feacb4fce3313cade2c61dc29c6271e6\x00refs/tags/v1.1.2\x00 \x002020-08-22T13:45:26-05:00"),
		[]byte("8c45a5635cf0a4968cc8c9dac2d61c388b53251e\x00refs/tags/v1.1.3\x00 \x002020-08-25T10:10:46-05:00"),
		[]byte("fc212da31ce157ef0795e934381509c5a50654f6\x00refs/tags/v1.1.4\x00 \x002020-08-26T14:02:47-05:00"),
		[]byte("4fd8b2c3522df32ffc8be983d42c3a504cc75fbc\x00refs/tags/v1.2.0\x00 \x002020-09-07T09:52:43-05:00"),
		[]byte("9741f54aa0f14be1103b00c89406393ea4d8a08a\x00refs/tags/v1.3.0\x00 \x002021-02-10T23:21:31+00:00"),
		[]byte("b358977103d2d66e2a3fc5f8081075c2834c4936\x00refs/tags/v1.3.1\x00 \x002021-02-24T20:16:45+00:00"),
		[]byte("2882ad236da4b649b4c1259d815bf1a378e3b92f\x00refs/tags/v1.4.0\x00 \x002021-05-13T10:41:02-05:00"),
		[]byte("340b84452286c18000afad9b140a32212a82840a\x00refs/tags/v1.5.0\x00 \x002021-05-20T18:41:41-05:00"),
	}, []byte("\n")))
	if err != nil {
		t.Fatalf("unexpected error parsing ref descriptions: %s", err)
	}

	makeBranch := func(name, createdDate string, isDefaultBranch bool) gitdomain.RefDescription {
		return gitdomain.RefDescription{Name: name, Type: gitdomain.RefTypeBranch, IsDefaultBranch: isDefaultBranch, CreatedDate: mustParseDate(createdDate, t)}
	}

	makeTag := func(name, createdDate string) gitdomain.RefDescription {
		return gitdomain.RefDescription{Name: name, Type: gitdomain.RefTypeTag, IsDefaultBranch: false, CreatedDate: mustParseDate(createdDate, t)}
	}

	expectedRefDescriptions := map[string][]gitdomain.RefDescription{
		"66a7ac584740245fc523da443a3f540a52f8af72": {makeBranch("bl/symbols", "2021-01-18T16:46:51-08:00", false)},
		"58537c06cf7ba8a562a3f5208fb7a8efbc971d0e": {makeBranch("bl/symbols-2", "2021-02-24T06:21:20-08:00", false)},
		"a40716031ae97ee7c5cdf1dec913567a4a7c50c8": {makeBranch("ef/wtf", "2021-02-10T10:50:08-06:00", false)},
		"e2e283fdaf6ea4a419cdbad142bbfd4b730080f8": {makeBranch("garo/go-and-typescript-lsif-indexing", "2020-04-29T16:45:46+00:00", false)},
		"c485d92c3d2065041bf29b3fe0b55ffac7e66b2a": {makeBranch("garo/index-specific-files", "2021-03-01T13:09:42-08:00", false)},
		"ce30aee6cc56f39d0ac6fee03c4c151c08a8cd2e": {makeBranch("master", "2021-06-16T11:51:09-07:00", true)},
		"ec5cfc8ab33370c698273b1a097af73ea289c92b": {makeBranch("nsc/bump-go-version", "2021-03-12T22:33:17+00:00", false)},
		"22b2c4f734f62060cae69da856fe3854defdcc87": {makeBranch("nsc/markupcontent", "2021-05-03T23:50:02+01:00", false)},
		"9df3358a18792fa9dbd40d506f2e0ad23fc11ee8": {makeBranch("nsc/random", "2021-02-10T16:29:06+00:00", false)},
		"a02b85b63345a1406d7a19727f7a5472c976e053": {makeBranch("sg/document-symbols", "2021-04-08T15:33:03-07:00", false)},
		"234b0a484519129b251164ecb0674ec27d154d2f": {makeBranch("symbols", "2021-01-01T22:51:55-08:00", false)},
		"6b5ae2e0ce568a7641174072271d109d7d0977c7": {gitdomain.RefDescription{Name: "v0.0.0", Type: gitdomain.RefTypeTag, IsDefaultBranch: false}},
		"c165bfff52e9d4f87891bba497e3b70fea144d89": {makeTag("v0.10.0", "2020-08-04T08:23:30-05:00")},
		"f73ee8ed601efea74f3b734eeb073307e1615606": {makeTag("v0.5.1", "2020-04-16T16:06:21-04:00")},
		"6057f7ed8d331c82030c713b650fc8fd2c0c2347": {makeTag("v0.5.2", "2020-04-16T16:20:26-04:00")},
		"7886287b8758d1baf19cf7b8253856128369a2a7": {makeTag("v0.5.3", "2020-04-16T16:55:58-04:00")},
		"b69f89473bbcc04dc52cafaf6baa504e34791f5a": {makeTag("v0.6.0", "2020-04-20T12:10:49-04:00")},
		"172b7fcf8b8c49b37b231693433586c2bfd1619e": {makeTag("v0.7.0", "2020-04-20T12:37:36-04:00")},
		"5bc35c78fb5fb388891ca944cd12d85fd6dede95": {makeTag("v0.8.0", "2020-05-05T12:53:18-05:00")},
		"14faa49ef098df9488536ca3c9b26d79e6bec4d6": {makeTag("v0.9.0", "2020-07-14T14:26:40-05:00")},
		"0a82af8b6914d8c81326eee5f3a7e1d1106547f1": {makeTag("v1.0.0", "2020-08-19T19:33:39-05:00")},
		"262defb72b96261a7d56b000d438c5c7ec6d0f3e": {makeTag("v1.1.0", "2020-08-21T14:15:44-05:00")},
		"806b96eb544e7e632a617c26402eccee6d67faed": {makeTag("v1.1.1", "2020-08-21T16:02:35-05:00")},
		"5d8865d6feacb4fce3313cade2c61dc29c6271e6": {makeTag("v1.1.2", "2020-08-22T13:45:26-05:00")},
		"8c45a5635cf0a4968cc8c9dac2d61c388b53251e": {makeTag("v1.1.3", "2020-08-25T10:10:46-05:00")},
		"fc212da31ce157ef0795e934381509c5a50654f6": {makeTag("v1.1.4", "2020-08-26T14:02:47-05:00")},
		"4fd8b2c3522df32ffc8be983d42c3a504cc75fbc": {makeTag("v1.2.0", "2020-09-07T09:52:43-05:00")},
		"9741f54aa0f14be1103b00c89406393ea4d8a08a": {makeTag("v1.3.0", "2021-02-10T23:21:31+00:00")},
		"b358977103d2d66e2a3fc5f8081075c2834c4936": {makeTag("v1.3.1", "2021-02-24T20:16:45+00:00")},
		"2882ad236da4b649b4c1259d815bf1a378e3b92f": {makeTag("v1.4.0", "2021-05-13T10:41:02-05:00")},
		"340b84452286c18000afad9b140a32212a82840a": {makeTag("v1.5.0", "2021-05-20T18:41:41-05:00")},
	}
	if diff := cmp.Diff(expectedRefDescriptions, refDescriptions); diff != "" {
		t.Errorf("unexpected ref descriptions (-want +got):\n%s", diff)
	}
}

func TestFilterRefDescriptions(t *testing.T) {
	ctx := actor.WithActor(context.Background(), &actor.Actor{
		UID: 1,
	})
	gitCommands := append(getGitCommandsWithFiles("file1", "file2"), getGitCommandsWithFiles("file3", "file4")...)
	repo := MakeGitRepository(t, gitCommands...)

	refDescriptions := map[string][]gitdomain.RefDescription{
		"d38233a79e037d2ab8170b0d0bc0aa438473e6da": {},
		"2775e60f523d3151a2a34ffdc659f500d0e73022": {},
		"2ba4dd2b9a27ec125fea7d72e12b9824ead18631": {},
		"9019942b8b92d5a70a7f546d97c451621c5059a6": {},
	}

	checker := getTestSubRepoPermsChecker("file3")
	filtered := filterRefDescriptions(ctx, database.NewMockDB(), repo, refDescriptions, checker)
	expectedRefDescriptions := map[string][]gitdomain.RefDescription{
		"d38233a79e037d2ab8170b0d0bc0aa438473e6da": {},
		"2ba4dd2b9a27ec125fea7d72e12b9824ead18631": {},
		"9019942b8b92d5a70a7f546d97c451621c5059a6": {},
	}
	if diff := cmp.Diff(expectedRefDescriptions, filtered); diff != "" {
		t.Errorf("unexpected ref descriptions (-want +got):\n%s", diff)
	}
}

func TestRefDescriptions(t *testing.T) {
	t.Parallel()
	ctx := actor.WithActor(context.Background(), &actor.Actor{
		UID: 1,
	})
	db := database.NewMockDB()
	gitCommands := append(getGitCommandsWithFiles("file1", "file2"), "git checkout -b my-other-branch")
	gitCommands = append(gitCommands, getGitCommandsWithFiles("file1-b2", "file2-b2")...)
	gitCommands = append(gitCommands, "git checkout -b my-branch-no-access")
	gitCommands = append(gitCommands, getGitCommandsWithFiles("file", "file-with-no-access")...)
	repo := MakeGitRepository(t, gitCommands...)

	makeBranch := func(name, createdDate string, isDefaultBranch bool) gitdomain.RefDescription {
		return gitdomain.RefDescription{Name: name, Type: gitdomain.RefTypeBranch, IsDefaultBranch: isDefaultBranch, CreatedDate: mustParseDate(createdDate, t)}
	}

	t.Run("basic", func(t *testing.T) {
		refDescriptions, err := RefDescriptions(ctx, db, repo, nil)
		if err != nil {
			t.Errorf("err calling RefDescriptions: %s", err)
		}
		expectedRefDescriptions := map[string][]gitdomain.RefDescription{
			"2ba4dd2b9a27ec125fea7d72e12b9824ead18631": {makeBranch("master", "2006-01-02T15:04:05Z", false)},
			"9d7a382983098eed6cf911bd933dfacb13116e42": {makeBranch("my-other-branch", "2006-01-02T15:04:05Z", false)},
			"7cf006d0599531db799c08d3b00d7fd06da33015": {makeBranch("my-branch-no-access", "2006-01-02T15:04:05Z", true)},
		}
		if diff := cmp.Diff(expectedRefDescriptions, refDescriptions); diff != "" {
			t.Errorf("unexpected ref descriptions (-want +got):\n%s", diff)
		}
	})

	t.Run("with sub-repo enabled", func(t *testing.T) {
		checker := getTestSubRepoPermsChecker("file-with-no-access")
		refDescriptions, err := RefDescriptions(ctx, db, repo, checker)
		if err != nil {
			t.Errorf("err calling RefDescriptions: %s", err)
		}
		expectedRefDescriptions := map[string][]gitdomain.RefDescription{
			"2ba4dd2b9a27ec125fea7d72e12b9824ead18631": {makeBranch("master", "2006-01-02T15:04:05Z", false)},
			"9d7a382983098eed6cf911bd933dfacb13116e42": {makeBranch("my-other-branch", "2006-01-02T15:04:05Z", false)},
		}
		if diff := cmp.Diff(expectedRefDescriptions, refDescriptions); diff != "" {
			t.Errorf("unexpected ref descriptions (-want +got):\n%s", diff)
		}
	})
}

func TestCommitsUniqueToBranch(t *testing.T) {
	t.Parallel()
	ctx := actor.WithActor(context.Background(), &actor.Actor{
		UID: 1,
	})
	db := database.NewMockDB()
	gitCommands := append([]string{"git checkout -b my-branch"}, getGitCommandsWithFiles("file1", "file2")...)
	gitCommands = append(gitCommands, getGitCommandsWithFiles("file3", "file-with-no-access")...)
	repo := MakeGitRepository(t, gitCommands...)

	t.Run("basic", func(t *testing.T) {
		commits, err := CommitsUniqueToBranch(ctx, db, repo, "my-branch", true, &time.Time{}, nil)
		if err != nil {
			t.Errorf("err calling RefDescriptions: %s", err)
		}
		expectedCommits := map[string]time.Time{
			"2775e60f523d3151a2a34ffdc659f500d0e73022": *mustParseDate("2006-01-02T15:04:05-00:00", t),
			"2ba4dd2b9a27ec125fea7d72e12b9824ead18631": *mustParseDate("2006-01-02T15:04:05-00:00", t),
			"791ce7cd8ca2d855e12f47f8692a62bc42477edc": *mustParseDate("2006-01-02T15:04:05-00:00", t),
			"d38233a79e037d2ab8170b0d0bc0aa438473e6da": *mustParseDate("2006-01-02T15:04:05-00:00", t),
		}
		if diff := cmp.Diff(expectedCommits, commits); diff != "" {
			t.Errorf("unexpected ref descriptions (-want +got):\n%s", diff)
		}
	})

	t.Run("with sub-repo enabled", func(t *testing.T) {
		checker := getTestSubRepoPermsChecker("file-with-no-access")
		commits, err := CommitsUniqueToBranch(ctx, db, repo, "my-branch", true, &time.Time{}, checker)
		if err != nil {
			t.Errorf("err calling RefDescriptions: %s", err)
		}
		expectedCommits := map[string]time.Time{
			"2775e60f523d3151a2a34ffdc659f500d0e73022": *mustParseDate("2006-01-02T15:04:05-00:00", t),
			"2ba4dd2b9a27ec125fea7d72e12b9824ead18631": *mustParseDate("2006-01-02T15:04:05-00:00", t),
			"d38233a79e037d2ab8170b0d0bc0aa438473e6da": *mustParseDate("2006-01-02T15:04:05-00:00", t),
		}
		if diff := cmp.Diff(expectedCommits, commits); diff != "" {
			t.Errorf("unexpected ref descriptions (-want +got):\n%s", diff)
		}
	})
}

func TestCommitDate(t *testing.T) {
	t.Parallel()
	ctx := actor.WithActor(context.Background(), &actor.Actor{
		UID: 1,
	})
	db := database.NewMockDB()
	gitCommands := getGitCommandsWithFiles("file1", "file2")
	repo := MakeGitRepository(t, gitCommands...)

	t.Run("basic", func(t *testing.T) {
		_, date, commitExists, err := CommitDate(ctx, db, repo, "d38233a79e037d2ab8170b0d0bc0aa438473e6da", nil)
		if err != nil {
			t.Errorf("error fetching CommitDate: %s", err)
		}
		if !commitExists {
			t.Errorf("commit should exist")
		}
		if !date.Equal(time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC)) {
			t.Errorf("unexpected date: %s", date)
		}
	})

	t.Run("with sub-repo permissions enabled", func(t *testing.T) {
		checker := getTestSubRepoPermsChecker("file1")
		_, date, commitExists, err := CommitDate(ctx, db, repo, "d38233a79e037d2ab8170b0d0bc0aa438473e6da", checker)
		if err != nil {
			t.Errorf("error fetching CommitDate: %s", err)
		}
		if commitExists {
			t.Errorf("expect commit to not exist since the user doesn't have access")
		}
		if !date.IsZero() {
			t.Errorf("expected date to be empty, got: %s", date)
		}
	})
}

func TestGetCommits(t *testing.T) {
	t.Parallel()
	ctx := actor.WithActor(context.Background(), &actor.Actor{
		UID: 1,
	})
	db := database.NewMockDB()

	repo1 := MakeGitRepository(t, getGitCommandsWithFiles("file1", "file2")...)
	repo2 := MakeGitRepository(t, getGitCommandsWithFiles("file3", "file4")...)
	repo3 := MakeGitRepository(t, getGitCommandsWithFiles("file5", "file6")...)

	repoCommits := []api.RepoCommit{
		{Repo: repo1, CommitID: api.CommitID("HEAD")},                                     // HEAD (file2)
		{Repo: repo1, CommitID: api.CommitID("HEAD~1")},                                   // HEAD~1 (file1)
		{Repo: repo2, CommitID: api.CommitID("67762ad757dd26cac4145f2b744fd93ad10a48e0")}, // HEAD (file4)
		{Repo: repo2, CommitID: api.CommitID("2b988222e844b570959a493f5b07ec020b89e122")}, // HEAD~1 (file3)
		{Repo: repo3, CommitID: api.CommitID("01bed0a")},                                  // abbrev HEAD (file6)
		{Repo: repo3, CommitID: api.CommitID("unresolvable")},                             // unresolvable
		{Repo: api.RepoName("unresolvable"), CommitID: api.CommitID("deadbeef")},          // unresolvable
	}

	t.Run("basic", func(t *testing.T) {
		expectedCommits := []*gitdomain.Commit{
			{
				ID:        "2ba4dd2b9a27ec125fea7d72e12b9824ead18631",
				Author:    gitdomain.Signature{Name: "a", Email: "a@a.com", Date: *mustParseDate("2006-01-02T15:04:05Z", t)},
				Committer: &gitdomain.Signature{Name: "a", Email: "a@a.com", Date: *mustParseDate("2006-01-02T15:04:05Z", t)},
				Message:   "commit2",
				Parents:   []api.CommitID{"d38233a79e037d2ab8170b0d0bc0aa438473e6da"},
			},
			{
				ID:        "d38233a79e037d2ab8170b0d0bc0aa438473e6da",
				Author:    gitdomain.Signature{Name: "a", Email: "a@a.com", Date: *mustParseDate("2006-01-02T15:04:05Z", t)},
				Committer: &gitdomain.Signature{Name: "a", Email: "a@a.com", Date: *mustParseDate("2006-01-02T15:04:05Z", t)},
				Message:   "commit1",
			},
			{
				ID:        "67762ad757dd26cac4145f2b744fd93ad10a48e0",
				Author:    gitdomain.Signature{Name: "a", Email: "a@a.com", Date: *mustParseDate("2006-01-02T15:04:05Z", t)},
				Committer: &gitdomain.Signature{Name: "a", Email: "a@a.com", Date: *mustParseDate("2006-01-02T15:04:05Z", t)},
				Message:   "commit2",
				Parents:   []api.CommitID{"2b988222e844b570959a493f5b07ec020b89e122"},
			},
			{
				ID:        "2b988222e844b570959a493f5b07ec020b89e122",
				Author:    gitdomain.Signature{Name: "a", Email: "a@a.com", Date: *mustParseDate("2006-01-02T15:04:05Z", t)},
				Committer: &gitdomain.Signature{Name: "a", Email: "a@a.com", Date: *mustParseDate("2006-01-02T15:04:05Z", t)},
				Message:   "commit1",
			},
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

		commits, err := getCommits(ctx, db, repoCommits, true, nil)
		if err != nil {
			t.Fatalf("unexpected error calling getCommits: %s", err)
		}
		if diff := cmp.Diff(expectedCommits, commits); diff != "" {
			t.Errorf("unexpected commits (-want +got):\n%s", diff)
		}
	})

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

		commits, err := getCommits(ctx, db, repoCommits, true, getTestSubRepoPermsChecker("file1", "file3"))
		if err != nil {
			t.Fatalf("unexpected error calling getCommits: %s", err)
		}
		if diff := cmp.Diff(expectedCommits, commits); diff != "" {
			t.Errorf("unexpected commits (-want +got):\n%s", diff)
		}
	})
}

func testCommits(ctx context.Context, label string, repo api.RepoName, opt CommitsOptions, checker authz.SubRepoPermissionChecker, wantTotal uint, wantCommits []*gitdomain.Commit, t *testing.T) {
	t.Helper()
	db := database.NewMockDB()
	commits, err := Commits(ctx, db, repo, opt, checker)
	if err != nil {
		t.Errorf("%s: Commits(): %s", label, err)
		return
	}

	total, err := commitCount(ctx, db, repo, opt)
	if err != nil {
		t.Errorf("%s: commitCount(): %s", label, err)
		return
	}
	if total != wantTotal {
		t.Errorf("%s: got %d total commits, want %d", label, total, wantTotal)
	}
	if len(commits) != len(wantCommits) {
		t.Errorf("%s: got %d commits, want %d", label, len(commits), len(wantCommits))
	}
	checkCommits(t, label, commits, wantCommits)
}

func checkCommits(t *testing.T, label string, commits, wantCommits []*gitdomain.Commit) {
	t.Helper()
	for i := 0; i < len(commits) || i < len(wantCommits); i++ {
		var gotC, wantC *gitdomain.Commit
		if i < len(commits) {
			gotC = commits[i]
		}
		if i < len(wantCommits) {
			wantC = wantCommits[i]
		}
		if !CommitsEqual(gotC, wantC) {
			t.Errorf("%s: got commit %d == %+v, want %+v", label, i, gotC, wantC)
		}
	}
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
