package git

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
)

func TestRepository_GetCommit(t *testing.T) {
	ctx := context.Background()

	gitCommands := []string{
		"GIT_COMMITTER_NAME=a GIT_COMMITTER_EMAIL=a@a.com GIT_COMMITTER_DATE=2006-01-02T15:04:05Z git commit --allow-empty -m foo --author='a <a@a.com>' --date 2006-01-02T15:04:05Z",
		"GIT_COMMITTER_NAME=c GIT_COMMITTER_EMAIL=c@c.com GIT_COMMITTER_DATE=2006-01-02T15:04:07Z git commit --allow-empty -m bar --author='a <a@a.com>' --date 2006-01-02T15:04:06Z",
	}
	wantGitCommit := &gitdomain.Commit{
		ID:        "b266c7e3ca00b1a17ad0b1449825d0854225c007",
		Author:    gitdomain.Signature{Name: "a", Email: "a@a.com", Date: MustParseTime(time.RFC3339, "2006-01-02T15:04:06Z")},
		Committer: &gitdomain.Signature{Name: "c", Email: "c@c.com", Date: MustParseTime(time.RFC3339, "2006-01-02T15:04:07Z")},
		Message:   "bar",
		Parents:   []api.CommitID{"ea167fe3d76b1e5fd3ed8ca44cbd2fe3897684f8"},
	}
	tests := map[string]struct {
		repo             api.RepoName
		id               api.CommitID
		wantCommit       *gitdomain.Commit
		noEnsureRevision bool
	}{
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

	oldRunCommitLog := runCommitLog

	for label, test := range tests {
		var noEnsureRevision bool
		t.Cleanup(func() {
			runCommitLog = oldRunCommitLog
		})
		runCommitLog = func(ctx context.Context, cmd *gitserver.Cmd, opt CommitsOptions) ([]*wrappedCommit, error) {
			// Track the value of NoEnsureRevision we pass to gitserver
			noEnsureRevision = opt.NoEnsureRevision
			return oldRunCommitLog(ctx, cmd, opt)
		}

		resolveRevisionOptions := ResolveRevisionOptions{
			NoEnsureRevision: test.noEnsureRevision,
		}
		commit, err := GetCommit(ctx, test.repo, test.id, resolveRevisionOptions)
		if err != nil {
			t.Errorf("%s: GetCommit: %s", label, err)
			continue
		}

		if !CommitsEqual(commit, test.wantCommit) {
			t.Errorf("%s: got commit == %+v, want %+v", label, commit, test.wantCommit)
		}

		// Test that trying to get a nonexistent commit returns RevisionNotFoundError.
		if _, err := GetCommit(ctx, test.repo, NonExistentCommitID, resolveRevisionOptions); !errors.HasType(err, &gitdomain.RevisionNotFoundError{}) {
			t.Errorf("%s: for nonexistent commit: got err %v, want RevisionNotFoundError", label, err)
		}

		if noEnsureRevision != test.noEnsureRevision {
			t.Fatalf("Expected %t, got %t", test.noEnsureRevision, noEnsureRevision)
		}
	}
}

func TestRepository_HasCommitAfter(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	testCases := []struct {
		commitDates []string
		after       string
		revspec     string
		want        bool
	}{
		{
			commitDates: []string{
				"2006-01-02T15:04:05Z",
				"2007-01-02T15:04:05Z",
				"2008-01-02T15:04:05Z",
			},
			after:   "2006-01-02T15:04:05Z",
			revspec: "master",
			want:    true,
		},
		{
			commitDates: []string{
				"2016-01-02T15:04:05Z",
				"2017-01-02T15:04:05Z",
				"2017-01-02T15:04:06Z",
			},
			after:   "1 year ago",
			revspec: "master",
			want:    false,
		},
		{
			commitDates: []string{
				"2006-01-02T15:04:05Z",
				"2007-01-02T15:04:05Z",
				"2008-01-02T15:04:05Z",
			},
			after:   "2010-01-02T15:04:05Z",
			revspec: "HEAD",
			want:    false,
		},
		{
			commitDates: []string{
				"2006-01-02T15:04:05Z",
				"2007-01-02T15:04:05Z",
				"2007-01-02T15:04:06Z",
			},
			after:   "2007-01-02T15:04:05Z",
			revspec: "HEAD",
			want:    true,
		},
		{
			commitDates: []string{
				"2016-01-02T15:04:05Z",
				"2017-01-02T15:04:05Z",
				"2017-01-02T15:04:06Z",
			},
			after:   "10 years ago",
			revspec: "HEAD",
			want:    true,
		},
	}

	for _, tc := range testCases {
		gitCommands := make([]string, len(tc.commitDates))
		for i, date := range tc.commitDates {
			gitCommands[i] = fmt.Sprintf("GIT_COMMITTER_NAME=a GIT_COMMITTER_EMAIL=a@a.com GIT_COMMITTER_DATE=%s git commit --allow-empty -m foo --author='a <a@a.com>'", date)
		}

		repo := MakeGitRepository(t, gitCommands...)
		got, err := HasCommitAfter(ctx, repo, tc.after, tc.revspec)
		if err != nil || got != tc.want {
			t.Errorf("got %t hascommitafter, want %t", got, tc.want)
		}
	}
}

func TestRepository_FirstEverCommit(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

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
	for _, tc := range testCases {
		gitCommands := make([]string, len(tc.commitDates))
		for i, date := range tc.commitDates {
			gitCommands[i] = fmt.Sprintf("GIT_COMMITTER_NAME=a GIT_COMMITTER_EMAIL=a@a.com GIT_COMMITTER_DATE=%s git commit --allow-empty -m foo --author='a <a@a.com>'", date)
		}

		repo := MakeGitRepository(t, gitCommands...)
		gotCommit, err := FirstEverCommit(ctx, repo)
		if err != nil {
			t.Fatal(err)
		}
		got := gotCommit.Committer.Date.Format(time.RFC3339)
		if got != tc.want {
			t.Errorf("got %q, want %q", got, tc.want)
		}
	}
}

func TestHead(t *testing.T) {
	t.Parallel()

	gitCommands := []string{
		"GIT_COMMITTER_NAME=a GIT_COMMITTER_EMAIL=a@a.com GIT_COMMITTER_DATE=2006-01-02T15:04:05Z git commit --allow-empty -m foo --author='a <a@a.com>' --date 2006-01-02T15:04:05Z",
	}
	repo := MakeGitRepository(t, gitCommands...)
	ctx := context.Background()

	head, exists, err := Head(ctx, repo)
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
}

func TestCommitExists(t *testing.T) {
	t.Parallel()

	gitCommands := []string{
		"GIT_COMMITTER_NAME=a GIT_COMMITTER_EMAIL=a@a.com GIT_COMMITTER_DATE=2006-01-02T15:04:05Z git commit --allow-empty -m foo --author='a <a@a.com>' --date 2006-01-02T15:04:05Z",
	}
	repo := MakeGitRepository(t, gitCommands...)
	ctx := context.Background()

	wantCommit := api.CommitID("ea167fe3d76b1e5fd3ed8ca44cbd2fe3897684f8")
	exists, err := CommitExists(ctx, repo, wantCommit)
	if err != nil {
		t.Fatal(err)
	}
	if !exists {
		t.Fatal("Should exist")
	}

	exists, err = CommitExists(ctx, repo, NonExistentCommitID)
	if err != nil {
		t.Fatal(err)
	}
	if exists {
		t.Fatal("Should not exist")
	}
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
				if _, err := Commits(ctx, test.repo, CommitsOptions{Range: string(NonExistentCommitID)}, nil); !errors.HasType(err, &gitdomain.RevisionNotFoundError{}) {
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
	checker := authz.NewMockSubRepoPermissionChecker()
	checker.EnabledFunc.SetDefaultHook(func() bool {
		return true
	})
	checker.PermissionsFunc.SetDefaultHook(func(ctx context.Context, i int32, content authz.RepoContent) (authz.Perms, error) {
		if content.Path == "file2" || content.Path == "file3" {
			return authz.None, nil
		}
		return authz.Read, nil
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
		repo        api.RepoName
		wantCommits []*gitdomain.Commit
		opt         CommitsOptions
		wantTotal   uint
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
		},
		"sub-repo perms with path (w/ no access) specified should return no commits": {
			repo:      MakeGitRepository(t, gitCommands...),
			wantTotal: 1,
			opt: CommitsOptions{
				//Range: "master",
				Path: "file2",
			},
			wantCommits: []*gitdomain.Commit{},
		},
		"sub-repo perms with path (w/ access) specified should return that commit": {
			repo:      MakeGitRepository(t, gitCommands...),
			wantTotal: 1,
			opt: CommitsOptions{
				//Range: "master",
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
		},
	}

	for label, test := range tests {
		commits, err := Commits(ctx, test.repo, test.opt, checker)
		if err != nil {
			t.Errorf("%s: Commits(): %s", label, err)
			return
		}

		if len(commits) != len(test.wantCommits) {
			t.Errorf("%s: got %d commits, want %d", label, len(commits), len(test.wantCommits))
		}

		checkCommits(t, label, commits, test.wantCommits)
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
	mustParseDate := func(s string) time.Time {
		date, err := time.Parse(time.RFC3339, s)
		if err != nil {
			t.Fatalf("unexpected error parsing date string: %s", err)
		}

		return date
	}

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
		"c165bfff52e9d4f87891bba497e3b70fea144d89": mustParseDate("2020-08-04T08:23:30-05:00"),
		"f73ee8ed601efea74f3b734eeb073307e1615606": mustParseDate("2020-04-16T16:06:21-04:00"),
		"6057f7ed8d331c82030c713b650fc8fd2c0c2347": mustParseDate("2020-04-16T16:20:26-04:00"),
		"7886287b8758d1baf19cf7b8253856128369a2a7": mustParseDate("2020-04-16T16:55:58-04:00"),
		"b69f89473bbcc04dc52cafaf6baa504e34791f5a": mustParseDate("2020-04-20T12:10:49-04:00"),
		"172b7fcf8b8c49b37b231693433586c2bfd1619e": mustParseDate("2020-04-20T12:37:36-04:00"),
		"5bc35c78fb5fb388891ca944cd12d85fd6dede95": mustParseDate("2020-05-05T12:53:18-05:00"),
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
	refDescriptions, err := parseRefDescriptions([]string{
		"66a7ac584740245fc523da443a3f540a52f8af72:refs/heads/bl/symbols: :2021-01-18T16:46:51-08:00",
		"58537c06cf7ba8a562a3f5208fb7a8efbc971d0e:refs/heads/bl/symbols-2: :2021-02-24T06:21:20-08:00",
		"a40716031ae97ee7c5cdf1dec913567a4a7c50c8:refs/heads/ef/wtf: :2021-02-10T10:50:08-06:00",
		"e2e283fdaf6ea4a419cdbad142bbfd4b730080f8:refs/heads/garo/go-and-typescript-lsif-indexing: :2020-04-29T16:45:46+00:00",
		"c485d92c3d2065041bf29b3fe0b55ffac7e66b2a:refs/heads/garo/index-specific-files: :2021-03-01T13:09:42-08:00",
		"ce30aee6cc56f39d0ac6fee03c4c151c08a8cd2e:refs/heads/master:*:2021-06-16T11:51:09-07:00",
		"ec5cfc8ab33370c698273b1a097af73ea289c92b:refs/heads/nsc/bump-go-version: :2021-03-12T22:33:17+00:00",
		"22b2c4f734f62060cae69da856fe3854defdcc87:refs/heads/nsc/markupcontent: :2021-05-03T23:50:02+01:00",
		"9df3358a18792fa9dbd40d506f2e0ad23fc11ee8:refs/heads/nsc/random: :2021-02-10T16:29:06+00:00",
		"a02b85b63345a1406d7a19727f7a5472c976e053:refs/heads/sg/document-symbols: :2021-04-08T15:33:03-07:00",
		"234b0a484519129b251164ecb0674ec27d154d2f:refs/heads/symbols: :2021-01-01T22:51:55-08:00",
		"c165bfff52e9d4f87891bba497e3b70fea144d89:refs/tags/v0.10.0: :2020-08-04T08:23:30-05:00",
		"f73ee8ed601efea74f3b734eeb073307e1615606:refs/tags/v0.5.1: :2020-04-16T16:06:21-04:00",
		"6057f7ed8d331c82030c713b650fc8fd2c0c2347:refs/tags/v0.5.2: :2020-04-16T16:20:26-04:00",
		"7886287b8758d1baf19cf7b8253856128369a2a7:refs/tags/v0.5.3: :2020-04-16T16:55:58-04:00",
		"b69f89473bbcc04dc52cafaf6baa504e34791f5a:refs/tags/v0.6.0: :2020-04-20T12:10:49-04:00",
		"172b7fcf8b8c49b37b231693433586c2bfd1619e:refs/tags/v0.7.0: :2020-04-20T12:37:36-04:00",
		"5bc35c78fb5fb388891ca944cd12d85fd6dede95:refs/tags/v0.8.0: :2020-05-05T12:53:18-05:00",
		"14faa49ef098df9488536ca3c9b26d79e6bec4d6:refs/tags/v0.9.0: :2020-07-14T14:26:40-05:00",
		"0a82af8b6914d8c81326eee5f3a7e1d1106547f1:refs/tags/v1.0.0: :2020-08-19T19:33:39-05:00",
		"262defb72b96261a7d56b000d438c5c7ec6d0f3e:refs/tags/v1.1.0: :2020-08-21T14:15:44-05:00",
		"806b96eb544e7e632a617c26402eccee6d67faed:refs/tags/v1.1.1: :2020-08-21T16:02:35-05:00",
		"5d8865d6feacb4fce3313cade2c61dc29c6271e6:refs/tags/v1.1.2: :2020-08-22T13:45:26-05:00",
		"8c45a5635cf0a4968cc8c9dac2d61c388b53251e:refs/tags/v1.1.3: :2020-08-25T10:10:46-05:00",
		"fc212da31ce157ef0795e934381509c5a50654f6:refs/tags/v1.1.4: :2020-08-26T14:02:47-05:00",
		"4fd8b2c3522df32ffc8be983d42c3a504cc75fbc:refs/tags/v1.2.0: :2020-09-07T09:52:43-05:00",
		"9741f54aa0f14be1103b00c89406393ea4d8a08a:refs/tags/v1.3.0: :2021-02-10T23:21:31+00:00",
		"b358977103d2d66e2a3fc5f8081075c2834c4936:refs/tags/v1.3.1: :2021-02-24T20:16:45+00:00",
		"2882ad236da4b649b4c1259d815bf1a378e3b92f:refs/tags/v1.4.0: :2021-05-13T10:41:02-05:00",
		"340b84452286c18000afad9b140a32212a82840a:refs/tags/v1.5.0: :2021-05-20T18:41:41-05:00",
	})
	if err != nil {
		t.Fatalf("unexpected error parsing ref descriptions: %s", err)
	}

	mustParseDate := func(s string) time.Time {
		date, err := time.Parse(time.RFC3339, s)
		if err != nil {
			t.Fatalf("unexpected error parsing date string: %s", err)
		}

		return date
	}

	makeBranch := func(name, createdDate string, isDefaultBranch bool) gitdomain.RefDescription {
		return gitdomain.RefDescription{Name: name, Type: gitdomain.RefTypeBranch, IsDefaultBranch: isDefaultBranch, CreatedDate: mustParseDate(createdDate)}
	}

	makeTag := func(name, createdDate string) gitdomain.RefDescription {
		return gitdomain.RefDescription{Name: name, Type: gitdomain.RefTypeTag, IsDefaultBranch: false, CreatedDate: mustParseDate(createdDate)}
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

func testCommits(ctx context.Context, label string, repo api.RepoName, opt CommitsOptions, checker authz.SubRepoPermissionChecker, wantTotal uint, wantCommits []*gitdomain.Commit, t *testing.T) {
	t.Helper()
	commits, err := Commits(ctx, repo, opt, checker)
	if err != nil {
		t.Errorf("%s: Commits(): %s", label, err)
		return
	}

	total, err := CommitCount(ctx, repo, opt)
	if err != nil {
		t.Errorf("%s: CommitCount(): %s", label, err)
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
func getTestSubRepoPermsChecker() authz.SubRepoPermissionChecker {
	checker := authz.NewMockSubRepoPermissionChecker()
	checker.EnabledFunc.SetDefaultHook(func() bool {
		return true
	})
	checker.PermissionsFunc.SetDefaultHook(func(ctx context.Context, i int32, content authz.RepoContent) (authz.Perms, error) {
		return authz.Read, nil
	})
	return checker
}
