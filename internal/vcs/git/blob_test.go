package git_test

import (
	"context"
	"io/ioutil"
	"os"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/vcs/git"
)

func TestRead(t *testing.T) {
	t.Parallel()

	const wantData = "abcd\n"
	repo := MakeGitRepository(t,
		"echo abcd > file1",
		"git add file1",
		"GIT_COMMITTER_NAME=a GIT_COMMITTER_EMAIL=a@a.com GIT_COMMITTER_DATE=2006-01-02T15:04:05Z git commit -m commit1 --author='a <a@a.com>' --date 2006-01-02T15:04:05Z",
	)
	const commitID = "3d689662de70f9e252d4f6f1d75284e23587d670"

	ctx := context.Background()

	tests := map[string]struct {
		file     string
		maxBytes int64
		checkFn  func(*testing.T, error, []byte)
	}{
		"all": {
			file: "file1",
			checkFn: func(t *testing.T, err error, data []byte) {
				if err != nil {
					t.Fatal(err)
				}
				if string(data) != wantData {
					t.Errorf("got %q, want %q", data, wantData)
				}
			},
		},

		"nonexistent": {
			file: "filexyz",
			checkFn: func(t *testing.T, err error, data []byte) {
				if err == nil {
					t.Fatal("err == nil")
				}
				if !os.IsNotExist(err) {
					t.Fatalf("got err %v, want os.IsNotExist", err)
				}
			},
		},
	}

	for name, test := range tests {
		t.Run(name+"-ReadFile", func(t *testing.T) {
			data, err := git.ReadFile(ctx, repo, commitID, test.file, test.maxBytes)
			test.checkFn(t, err, data)
		})
		t.Run(name+"-GetFileReader", func(t *testing.T) {
			rc, err := git.NewFileReader(ctx, repo, commitID, test.file)
			if err != nil {
				t.Fatal(err)
			}
			defer rc.Close()
			data, err := ioutil.ReadAll(rc)
			test.checkFn(t, err, data)
		})
	}

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
