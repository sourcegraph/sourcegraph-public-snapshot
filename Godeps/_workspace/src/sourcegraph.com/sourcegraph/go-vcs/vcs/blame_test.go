package vcs_test

import (
	"reflect"
	"testing"
	"time"

	"sourcegraph.com/sourcegraph/go-vcs/vcs"
)

func TestRepository_BlameFile(t *testing.T) {
	t.Parallel()

	cmds := []string{
		"echo line1 > f",
		"git add f",
		"GIT_COMMITTER_NAME=a GIT_COMMITTER_EMAIL=a@a.com GIT_COMMITTER_DATE=2006-01-02T15:04:05Z git commit -m foo --author='a <a@a.com>' --date 2006-01-02T15:04:05Z",
		"echo line2 >> f",
		"git add f",
		"GIT_COMMITTER_NAME=a GIT_COMMITTER_EMAIL=a@a.com GIT_COMMITTER_DATE=2006-01-02T15:04:05Z git commit -m foo --author='a <a@a.com>' --date 2006-01-02T15:04:05Z",
	}
	hgCommands := []string{
		"echo line1 > f",
		"touch --date=2006-01-02T15:04:05Z f || touch -t " + times[0] + " f",
		"hg add f",
		"hg commit -m foo --date '2006-12-06 13:18:29 UTC' --user 'a <a@a.com>'",
		"echo line2 >> f",
		"touch --date=2006-01-02T15:04:05Z f || touch -t " + times[0] + " f",
		"hg add f",
		"hg commit -m foo --date '2006-12-06 13:18:29 UTC' --user 'a <a@a.com>'",
	}
	tests := map[string]struct {
		repo interface {
			vcs.Blamer
			ResolveRevision(spec string) (vcs.CommitID, error)
		}
		path string
		opt  *vcs.BlameOptions

		wantHunks []*vcs.Hunk
	}{
		"git libgit2": {
			repo: makeGitRepositoryLibGit2(t, cmds...),
			path: "f",
			opt: &vcs.BlameOptions{
				NewestCommit: "master",
			},
			wantHunks: []*vcs.Hunk{
				{
					StartLine: 1, EndLine: 2, StartByte: 0, EndByte: 6, CommitID: "e6093374dcf5725d8517db0dccbbf69df65dbde0",
					Author: vcs.Signature{Name: "a", Email: "a@a.com", Date: mustParseTime(time.RFC3339, "2006-01-02T15:04:05Z")},
				},
				{
					StartLine: 2, EndLine: 3, StartByte: 6, EndByte: 12, CommitID: "fad406f4fe02c358a09df0d03ec7a36c2c8a20f1",
					Author: vcs.Signature{Name: "a", Email: "a@a.com", Date: mustParseTime(time.RFC3339, "2006-01-02T15:04:05Z")},
				},
			},
		},
		"git cmd": {
			repo: makeGitRepositoryCmd(t, cmds...),
			path: "f",
			opt: &vcs.BlameOptions{
				NewestCommit: "master",
			},
			wantHunks: []*vcs.Hunk{
				{
					StartLine: 1, EndLine: 2, StartByte: 0, EndByte: 6, CommitID: "e6093374dcf5725d8517db0dccbbf69df65dbde0",
					Author: vcs.Signature{Name: "a", Email: "a@a.com", Date: mustParseTime(time.RFC3339, "2006-01-02T15:04:05Z")},
				},
				{
					StartLine: 2, EndLine: 3, StartByte: 6, EndByte: 12, CommitID: "fad406f4fe02c358a09df0d03ec7a36c2c8a20f1",
					Author: vcs.Signature{Name: "a", Email: "a@a.com", Date: mustParseTime(time.RFC3339, "2006-01-02T15:04:05Z")},
				},
			},
		},
		"hg cmd": {
			repo: makeHgRepositoryCmd(t, hgCommands...),
			path: "f",
			opt: &vcs.BlameOptions{
				NewestCommit: "tip",
			},
			wantHunks: []*vcs.Hunk{
				{
					StartLine: 1, EndLine: 2, StartByte: 0, EndByte: 6, CommitID: "f1f126ec4cf9398d85e8dac873afc3f9b174b1d6",
					Author: vcs.Signature{Name: "a", Email: "a@a.com", Date: mustParseTime(time.RFC3339, "2006-12-06T13:18:29Z")},
				},
				{
					StartLine: 2, EndLine: 3, StartByte: 6, EndByte: 12, CommitID: "63e47acf80095270f4e2b81e8cc01a89416c0cf3",
					Author: vcs.Signature{Name: "a", Email: "a@a.com", Date: mustParseTime(time.RFC3339, "2006-12-06T13:18:29Z")},
				},
			},
		},
	}

	for label, test := range tests {
		newestCommitID, err := test.repo.ResolveRevision(string(test.opt.NewestCommit))
		if err != nil {
			t.Errorf("%s: ResolveRevision(%q) on base: %s", label, test.opt.NewestCommit, err)
			continue
		}

		test.opt.NewestCommit = newestCommitID
		hunks, err := test.repo.BlameFile(test.path, test.opt)
		if err != nil {
			t.Errorf("%s: BlameFile(%s, %+v): %s", label, test.path, test.opt, err)
			continue
		}

		if !reflect.DeepEqual(hunks, test.wantHunks) {
			t.Errorf("%s: hunks != wantHunks\n\nhunks ==========\n%s\n\nwantHunks ==========\n%s", label, asJSON(hunks), asJSON(test.wantHunks))
		}
	}
}
