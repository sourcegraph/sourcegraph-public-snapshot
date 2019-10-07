package git_test

import (
	"context"
	"os"
	"testing"

	"github.com/sourcegraph/sourcegraph/pkg/vcs/git"
	"github.com/sourcegraph/sourcegraph/pkg/vcs/git/gittest"
)

func TestReadFile(t *testing.T) {
	t.Parallel()

	const wantData = "abcd\n"
	repo := gittest.MakeGitRepository(t,
		"echo abcd > file1",
		"git add file1",
		"GIT_COMMITTER_NAME=a GIT_COMMITTER_EMAIL=a@a.com GIT_COMMITTER_DATE=2006-01-02T15:04:05Z git commit -m commit1 --author='a <a@a.com>' --date 2006-01-02T15:04:05Z",
	)
	const commitID = "3d689662de70f9e252d4f6f1d75284e23587d670"

	ctx := context.Background()

	t.Run("all", func(t *testing.T) {
		data, err := git.ReadFile(ctx, repo, commitID, "file1", -1)
		if err != nil {
			t.Fatal(err)
		}
		if string(data) != wantData {
			t.Errorf("got %q, want %q", data, wantData)
		}
	})

	t.Run("nonexistent", func(t *testing.T) {
		_, err := git.ReadFile(ctx, repo, commitID, "filexyz", -1)
		if err == nil {
			t.Fatal("err == nil")
		}
		if !os.IsNotExist(err) {
			t.Fatalf("got err %v, want os.IsNotExist", err)
		}
	})

	t.Run("maxBytes", func(t *testing.T) {
		data, err := git.ReadFile(ctx, repo, commitID, "file1", 3)
		if err != nil {
			t.Fatal(err)
		}
		if string(data) != wantData[:3] {
			t.Errorf("got %q, want %q", data, wantData)
		}
	})
}
