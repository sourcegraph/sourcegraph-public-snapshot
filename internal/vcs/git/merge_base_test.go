package git

import (
	"context"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
)

func TestMerger_MergeBase(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	db := database.NewMockDB()

	// TODO(sqs): implement for hg
	// TODO(sqs): make a more complex test case

	cmds := []string{
		"echo line1 > f",
		"git add f",
		"GIT_COMMITTER_NAME=a GIT_COMMITTER_EMAIL=a@a.com GIT_COMMITTER_DATE=2006-01-02T15:04:05Z git commit -m foo --author='a <a@a.com>' --date 2006-01-02T15:04:05Z",
		"git tag testbase",
		"git checkout -b b2",
		"echo line2 >> f",
		"git add f",
		"GIT_COMMITTER_NAME=a GIT_COMMITTER_EMAIL=a@a.com GIT_COMMITTER_DATE=2006-01-02T15:04:05Z git commit -m foo --author='a <a@a.com>' --date 2006-01-02T15:04:05Z",
		"git checkout master",
		"echo line3 > h",
		"git add h",
		"GIT_COMMITTER_NAME=a GIT_COMMITTER_EMAIL=a@a.com GIT_COMMITTER_DATE=2006-01-02T15:04:05Z git commit -m qux --author='a <a@a.com>' --date 2006-01-02T15:04:05Z",
	}
	tests := map[string]struct {
		repo api.RepoName
		a, b string // can be any revspec; is resolved during the test

		wantMergeBase string // can be any revspec; is resolved during test
	}{
		"git cmd": {
			repo: MakeGitRepository(t, cmds...),
			a:    "master", b: "b2",
			wantMergeBase: "testbase",
		},
	}

	for label, test := range tests {
		a, err := gitserver.ResolveRevision(ctx, db, test.repo, test.a, gitserver.ResolveRevisionOptions{})
		if err != nil {
			t.Errorf("%s: ResolveRevision(%q) on a: %s", label, test.a, err)
			continue
		}

		b, err := gitserver.ResolveRevision(ctx, db, test.repo, test.b, gitserver.ResolveRevisionOptions{})
		if err != nil {
			t.Errorf("%s: ResolveRevision(%q) on b: %s", label, test.b, err)
			continue
		}

		want, err := gitserver.ResolveRevision(ctx, db, test.repo, test.wantMergeBase, gitserver.ResolveRevisionOptions{})
		if err != nil {
			t.Errorf("%s: ResolveRevision(%q) on wantMergeBase: %s", label, test.wantMergeBase, err)
			continue
		}

		mb, err := MergeBase(ctx, db, test.repo, a, b)
		if err != nil {
			t.Errorf("%s: MergeBase(%s, %s): %s", label, a, b, err)
			continue
		}

		if mb != want {
			t.Errorf("%s: MergeBase(%s, %s): got %q, want %q", label, a, b, mb, want)
			continue
		}
	}
}
