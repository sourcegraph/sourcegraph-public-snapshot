package git

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
)

func TestRepository_BlameFile(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	gitCommands := []string{
		"echo line1 > f",
		"git add f",
		"GIT_COMMITTER_NAME=a GIT_COMMITTER_EMAIL=a@a.com GIT_COMMITTER_DATE=2006-01-02T15:04:05Z git commit -m foo --author='a <a@a.com>' --date 2006-01-02T15:04:05Z",
		"echo line2 >> f",
		"git add f",
		"GIT_COMMITTER_NAME=a GIT_COMMITTER_EMAIL=a@a.com GIT_COMMITTER_DATE=2006-01-02T15:04:05Z git commit -m foo --author='a <a@a.com>' --date 2006-01-02T15:04:05Z",
		"git mv f f2",
		"echo line3 >> f2",
		"git add f2",
		"GIT_COMMITTER_NAME=a GIT_COMMITTER_EMAIL=a@a.com GIT_COMMITTER_DATE=2006-01-02T15:04:05Z git commit -m foo --author='a <a@a.com>' --date 2006-01-02T15:04:05Z",
	}
	gitWantHunks := []*Hunk{
		{
			StartLine: 1, EndLine: 2, StartByte: 0, EndByte: 6, CommitID: "e6093374dcf5725d8517db0dccbbf69df65dbde0",
			Message: "foo", Author: gitdomain.Signature{Name: "a", Email: "a@a.com", Date: MustParseTime(time.RFC3339, "2006-01-02T15:04:05Z")},
			Filename: "f",
		},
		{
			StartLine: 2, EndLine: 3, StartByte: 6, EndByte: 12, CommitID: "fad406f4fe02c358a09df0d03ec7a36c2c8a20f1",
			Message: "foo", Author: gitdomain.Signature{Name: "a", Email: "a@a.com", Date: MustParseTime(time.RFC3339, "2006-01-02T15:04:05Z")},
			Filename: "f",
		},
		{
			StartLine: 3, EndLine: 4, StartByte: 12, EndByte: 18, CommitID: "311d75a2b414a77f5158a0ed73ec476f5469b286",
			Message: "foo", Author: gitdomain.Signature{Name: "a", Email: "a@a.com", Date: MustParseTime(time.RFC3339, "2006-01-02T15:04:05Z")},
			Filename: "f2",
		},
	}
	tests := map[string]struct {
		repo api.RepoName
		path string
		opt  *BlameOptions

		wantHunks []*Hunk
	}{
		"git cmd": {
			repo: MakeGitRepository(t, gitCommands...),
			path: "f2",
			opt: &BlameOptions{
				NewestCommit: "master",
			},
			wantHunks: gitWantHunks,
		},
	}

	for label, test := range tests {
		newestCommitID, err := ResolveRevision(ctx, test.repo, string(test.opt.NewestCommit), ResolveRevisionOptions{})
		if err != nil {
			t.Errorf("%s: ResolveRevision(%q) on base: %s", label, test.opt.NewestCommit, err)
			continue
		}

		test.opt.NewestCommit = newestCommitID
		runBlameFileTest(ctx, t, test.repo, test.path, test.opt, nil, label, test.wantHunks)

		checker := authz.NewMockSubRepoPermissionChecker()
		ctx = actor.WithActor(ctx, &actor.Actor{
			UID: 1,
		})
		// Sub-repo permissions
		// Case: user has read access to file, doesn't filter anything
		checker.EnabledFunc.SetDefaultHook(func() bool {
			return true
		})
		checker.PermissionsFunc.SetDefaultHook(func(ctx context.Context, i int32, content authz.RepoContent) (authz.Perms, error) {
			if content.Path == "f2" {
				return authz.Read, nil
			}
			return authz.None, nil
		})
		runBlameFileTest(ctx, t, test.repo, test.path, test.opt, checker, label, test.wantHunks)

		// Sub-repo permissions
		// Case: user doesn't have access to the file, nothing returned.
		checker.PermissionsFunc.SetDefaultHook(func(ctx context.Context, i int32, content authz.RepoContent) (authz.Perms, error) {
			return authz.None, nil
		})
		runBlameFileTest(ctx, t, test.repo, test.path, test.opt, checker, label, nil)
	}
}

func runBlameFileTest(ctx context.Context, t *testing.T, repo api.RepoName, path string, opt *BlameOptions,
	checker authz.SubRepoPermissionChecker, label string, wantHunks []*Hunk) {
	t.Helper()
	hunks, err := BlameFile(ctx, repo, path, opt, checker)
	if err != nil {
		t.Errorf("%s: BlameFile(%s, %+v): %s", label, path, opt, err)
		return
	}
	if !reflect.DeepEqual(hunks, wantHunks) {
		t.Errorf("%s: hunks != wantHunks\n\nhunks ==========\n%s\n\nwantHunks ==========\n%s", label, AsJSON(hunks), AsJSON(wantHunks))
	}
}
