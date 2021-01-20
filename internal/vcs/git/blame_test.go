package git

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/gitserver"
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
	}
	gitWantHunks := []*Hunk{
		{
			StartLine: 1, EndLine: 2, StartByte: 0, EndByte: 6, CommitID: "e6093374dcf5725d8517db0dccbbf69df65dbde0",
			Message: "foo", Author: Signature{Name: "a", Email: "a@a.com", Date: MustParseTime(time.RFC3339, "2006-01-02T15:04:05Z")},
		},
		{
			StartLine: 2, EndLine: 3, StartByte: 6, EndByte: 12, CommitID: "fad406f4fe02c358a09df0d03ec7a36c2c8a20f1",
			Message: "foo", Author: Signature{Name: "a", Email: "a@a.com", Date: MustParseTime(time.RFC3339, "2006-01-02T15:04:05Z")},
		},
	}
	tests := map[string]struct {
		repo gitserver.Repo
		path string
		opt  *BlameOptions

		wantHunks []*Hunk
	}{
		"git cmd": {
			repo: MakeGitRepository(t, gitCommands...),
			path: "f",
			opt: &BlameOptions{
				NewestCommit: "master",
			},
			wantHunks: gitWantHunks,
		},
	}

	for label, test := range tests {
		newestCommitID, err := ResolveRevision(ctx, test.repo, nil, string(test.opt.NewestCommit), ResolveRevisionOptions{})
		if err != nil {
			t.Errorf("%s: ResolveRevision(%q) on base: %s", label, test.opt.NewestCommit, err)
			continue
		}

		test.opt.NewestCommit = newestCommitID
		hunks, err := BlameFile(ctx, test.repo, test.path, test.opt)
		if err != nil {
			t.Errorf("%s: BlameFile(%s, %+v): %s", label, test.path, test.opt, err)
			continue
		}

		if !reflect.DeepEqual(hunks, test.wantHunks) {
			t.Errorf("%s: hunks != wantHunks\n\nhunks ==========\n%s\n\nwantHunks ==========\n%s", label, AsJSON(hunks), AsJSON(test.wantHunks))
		}
	}
}
