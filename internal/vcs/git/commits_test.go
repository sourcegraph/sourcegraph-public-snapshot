package git

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
)

func TestRepository_GetCommit(t *testing.T) {
	ctx := context.Background()

	gitCommands := []string{
		"GIT_COMMITTER_NAME=a GIT_COMMITTER_EMAIL=a@a.com GIT_COMMITTER_DATE=2006-01-02T15:04:05Z git commit --allow-empty -m foo --author='a <a@a.com>' --date 2006-01-02T15:04:05Z",
		"GIT_COMMITTER_NAME=c GIT_COMMITTER_EMAIL=c@c.com GIT_COMMITTER_DATE=2006-01-02T15:04:07Z git commit --allow-empty -m bar --author='a <a@a.com>' --date 2006-01-02T15:04:06Z",
	}
	wantGitCommit := &Commit{
		ID:        "b266c7e3ca00b1a17ad0b1449825d0854225c007",
		Author:    Signature{Name: "a", Email: "a@a.com", Date: MustParseTime(time.RFC3339, "2006-01-02T15:04:06Z")},
		Committer: &Signature{Name: "c", Email: "c@c.com", Date: MustParseTime(time.RFC3339, "2006-01-02T15:04:07Z")},
		Message:   "bar",
		Parents:   []api.CommitID{"ea167fe3d76b1e5fd3ed8ca44cbd2fe3897684f8"},
	}
	tests := map[string]struct {
		repo             gitserver.Repo
		id               api.CommitID
		wantCommit       *Commit
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
		runCommitLog = func(ctx context.Context, cmd *gitserver.Cmd, opt CommitsOptions) ([]*Commit, error) {
			// Track the value of NoEnsureRevision we pass to gitserver
			noEnsureRevision = opt.NoEnsureRevision
			return oldRunCommitLog(ctx, cmd, opt)
		}

		resolveRevisionOptions := ResolveRevisionOptions{
			NoEnsureRevision: test.noEnsureRevision,
		}
		commit, err := GetCommit(ctx, test.repo, nil, test.id, resolveRevisionOptions)
		if err != nil {
			t.Errorf("%s: GetCommit: %s", label, err)
			continue
		}

		if !CommitsEqual(commit, test.wantCommit) {
			t.Errorf("%s: got commit == %+v, want %+v", label, commit, test.wantCommit)
		}

		// Test that trying to get a nonexistent commit returns RevisionNotFoundError.
		if _, err := GetCommit(ctx, test.repo, nil, NonExistentCommitID, resolveRevisionOptions); !gitserver.IsRevisionNotFound(err) {
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

func TestRepository_Commits(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// TODO(sqs): test CommitsOptions.Base

	gitCommands := []string{
		"GIT_COMMITTER_NAME=a GIT_COMMITTER_EMAIL=a@a.com GIT_COMMITTER_DATE=2006-01-02T15:04:05Z git commit --allow-empty -m foo --author='a <a@a.com>' --date 2006-01-02T15:04:05Z",
		"GIT_COMMITTER_NAME=c GIT_COMMITTER_EMAIL=c@c.com GIT_COMMITTER_DATE=2006-01-02T15:04:07Z git commit --allow-empty -m bar --author='a <a@a.com>' --date 2006-01-02T15:04:06Z",
	}
	wantGitCommits := []*Commit{
		{
			ID:        "b266c7e3ca00b1a17ad0b1449825d0854225c007",
			Author:    Signature{Name: "a", Email: "a@a.com", Date: MustParseTime(time.RFC3339, "2006-01-02T15:04:06Z")},
			Committer: &Signature{Name: "c", Email: "c@c.com", Date: MustParseTime(time.RFC3339, "2006-01-02T15:04:07Z")},
			Message:   "bar",
			Parents:   []api.CommitID{"ea167fe3d76b1e5fd3ed8ca44cbd2fe3897684f8"},
		},
		{
			ID:        "ea167fe3d76b1e5fd3ed8ca44cbd2fe3897684f8",
			Author:    Signature{Name: "a", Email: "a@a.com", Date: MustParseTime(time.RFC3339, "2006-01-02T15:04:05Z")},
			Committer: &Signature{Name: "a", Email: "a@a.com", Date: MustParseTime(time.RFC3339, "2006-01-02T15:04:05Z")},
			Message:   "foo",
			Parents:   nil,
		},
	}
	tests := map[string]struct {
		repo        gitserver.Repo
		id          api.CommitID
		wantCommits []*Commit
		wantTotal   uint
	}{
		"git cmd": {
			repo:        MakeGitRepository(t, gitCommands...),
			id:          "b266c7e3ca00b1a17ad0b1449825d0854225c007",
			wantCommits: wantGitCommits,
			wantTotal:   2,
		},
	}

	for label, test := range tests {
		commits, err := Commits(ctx, test.repo, CommitsOptions{Range: string(test.id)})
		if err != nil {
			t.Errorf("%s: Commits: %s", label, err)
			continue
		}

		total, err := CommitCount(ctx, test.repo, CommitsOptions{Range: string(test.id)})
		if err != nil {
			t.Errorf("%s: CommitCount: %s", label, err)
			continue
		}

		if total != test.wantTotal {
			t.Errorf("%s: got %d total commits, want %d", label, total, test.wantTotal)
		}

		if len(commits) != len(test.wantCommits) {
			t.Errorf("%s: got %d commits, want %d", label, len(commits), len(test.wantCommits))
		}

		for i := 0; i < len(commits) || i < len(test.wantCommits); i++ {
			var gotC, wantC *Commit
			if i < len(commits) {
				gotC = commits[i]
			}
			if i < len(test.wantCommits) {
				wantC = test.wantCommits[i]
			}
			if !CommitsEqual(gotC, wantC) {
				t.Errorf("%s: got commit %d == %+v, want %+v", label, i, gotC, wantC)
			}
		}

		// Test that trying to get a nonexistent commit returns RevisionNotFoundError.
		if _, err := Commits(ctx, test.repo, CommitsOptions{Range: string(NonExistentCommitID)}); !gitserver.IsRevisionNotFound(err) {
			t.Errorf("%s: for nonexistent commit: got err %v, want RevisionNotFoundError", label, err)
		}
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
	wantGitCommits := []*Commit{
		{
			ID:        "b266c7e3ca00b1a17ad0b1449825d0854225c007",
			Author:    Signature{Name: "a", Email: "a@a.com", Date: MustParseTime(time.RFC3339, "2006-01-02T15:04:06Z")},
			Committer: &Signature{Name: "c", Email: "c@c.com", Date: MustParseTime(time.RFC3339, "2006-01-02T15:04:07Z")},
			Message:   "bar",
			Parents:   []api.CommitID{"ea167fe3d76b1e5fd3ed8ca44cbd2fe3897684f8"},
		},
	}
	wantGitCommits2 := []*Commit{
		{
			ID:        "ade564eba4cf904492fb56dcd287ac633e6e082c",
			Author:    Signature{Name: "a", Email: "a@a.com", Date: MustParseTime(time.RFC3339, "2006-01-02T15:04:08Z")},
			Committer: &Signature{Name: "c", Email: "c@c.com", Date: MustParseTime(time.RFC3339, "2006-01-02T15:04:08Z")},
			Message:   "qux",
			Parents:   []api.CommitID{"b266c7e3ca00b1a17ad0b1449825d0854225c007"},
		},
	}
	tests := map[string]struct {
		repo        gitserver.Repo
		opt         CommitsOptions
		wantCommits []*Commit
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
	}

	for label, test := range tests {
		commits, err := Commits(ctx, test.repo, test.opt)
		if err != nil {
			t.Errorf("%s: Commits(): %s", label, err)
			continue
		}

		total, err := CommitCount(ctx, test.repo, test.opt)
		if err != nil {
			t.Errorf("%s: CommitCount(): %s", label, err)
			continue
		}

		if total != test.wantTotal {
			t.Errorf("%s: got %d total commits, want %d", label, total, test.wantTotal)
		}

		if len(commits) != len(test.wantCommits) {
			t.Errorf("%s: got %d commits, want %d", label, len(commits), len(test.wantCommits))
		}

		for i := 0; i < len(commits) || i < len(test.wantCommits); i++ {
			var gotC, wantC *Commit
			if i < len(commits) {
				gotC = commits[i]
			}
			if i < len(test.wantCommits) {
				wantC = test.wantCommits[i]
			}
			if !CommitsEqual(gotC, wantC) {
				t.Errorf("%s: got commit %d == %+v, want %+v", label, i, gotC, wantC)
			}
		}
	}
}

func TestRepository_Commits_options_path(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	gitCommands := []string{
		"GIT_COMMITTER_NAME=a GIT_COMMITTER_EMAIL=a@a.com GIT_COMMITTER_DATE=2006-01-02T15:04:05Z git commit --allow-empty -m commit1 --author='a <a@a.com>' --date 2006-01-02T15:04:05Z",
		"touch file1",
		"touch --date=2006-01-02T15:04:05Z file1 || touch -t " + Times[0] + " file1",
		"git add file1",
		"GIT_COMMITTER_NAME=a GIT_COMMITTER_EMAIL=a@a.com GIT_COMMITTER_DATE=2006-01-02T15:04:05Z git commit -m commit2 --author='a <a@a.com>' --date 2006-01-02T15:04:05Z",
		"GIT_COMMITTER_NAME=c GIT_COMMITTER_EMAIL=c@c.com GIT_COMMITTER_DATE=2006-01-02T15:04:07Z git commit --allow-empty -m commit3 --author='a <a@a.com>' --date 2006-01-02T15:04:06Z",
	}
	wantGitCommits := []*Commit{
		{
			ID:        "546a3ef26e581624ef997cb8c0ba01ee475fc1dc",
			Author:    Signature{Name: "a", Email: "a@a.com", Date: MustParseTime(time.RFC3339, "2006-01-02T15:04:05Z")},
			Committer: &Signature{Name: "a", Email: "a@a.com", Date: MustParseTime(time.RFC3339, "2006-01-02T15:04:05Z")},
			Message:   "commit2",
			Parents:   []api.CommitID{"a04652fa1998a0a7d2f2f77ecb7021de943d3aab"},
		},
	}
	tests := map[string]struct {
		repo        gitserver.Repo
		opt         CommitsOptions
		wantCommits []*Commit
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

	for label, test := range tests {
		commits, err := Commits(ctx, test.repo, test.opt)
		if err != nil {
			t.Errorf("%s: Commits(): %s", label, err)
			continue
		}

		total, err := CommitCount(ctx, test.repo, test.opt)
		if err != nil {
			t.Errorf("%s: CommitCount: %s", label, err)
			continue
		}

		if total != test.wantTotal {
			t.Errorf("%s: got %d total commits, want %d", label, total, test.wantTotal)
		}

		if len(commits) != len(test.wantCommits) {
			t.Errorf("%s: got %d commits, want %d", label, len(commits), len(test.wantCommits))
		}

		for i := 0; i < len(commits) || i < len(test.wantCommits); i++ {
			var gotC, wantC *Commit
			if i < len(commits) {
				gotC = commits[i]
			}
			if i < len(test.wantCommits) {
				wantC = test.wantCommits[i]
			}
			if !CommitsEqual(gotC, wantC) {
				t.Errorf("%s: got commit %d == %+v, want %+v", label, i, gotC, wantC)
			}
		}
	}
}
