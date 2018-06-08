package git_test

import (
	"archive/zip"
	"bytes"
	"io/ioutil"
	"reflect"
	"testing"

	"github.com/sourcegraph/sourcegraph/pkg/gitserver"
	"github.com/sourcegraph/sourcegraph/pkg/vcs/git"
)

func TestRepository_Archive(t *testing.T) {
	t.Parallel()

	gitCommands := []string{
		"mkdir dir1",
		"echo -n infile1 > dir1/file1",
		"touch --date=2006-01-02T15:04:05Z dir1 dir1/file1 || touch -t " + times[0] + " dir1 dir1/file1",
		"git add dir1/file1",
		"GIT_COMMITTER_NAME=a GIT_COMMITTER_EMAIL=a@a.com GIT_COMMITTER_DATE=2006-01-02T15:04:05Z git commit -m commit1 --author='a <a@a.com>' --date 2006-01-02T15:04:05Z",
		"echo -n infile2 > 'file 2'",
		"touch --date=2014-05-06T19:20:21Z 'file 2' || touch -t " + times[1] + " 'file 2'",
		"git add 'file 2'",
		"GIT_COMMITTER_NAME=a GIT_COMMITTER_EMAIL=a@a.com GIT_COMMITTER_DATE=2014-05-06T19:20:21Z git commit -m commit2 --author='a <a@a.com>' --date 2014-05-06T19:20:21Z",
	}
	tests := map[string]struct {
		repo gitserver.Repo
		want map[string]string
	}{
		"git cmd": {
			repo: makeGitRepository(t, gitCommands...),
			want: map[string]string{
				"dir1/":      "",
				"dir1/file1": "infile1",
				"file 2":     "infile2",
			},
		},
	}

	for label, test := range tests {
		data, err := git.Archive(ctx, test.repo, "HEAD")
		if err != nil {
			t.Errorf("%s: Archive: %s", label, err)
			continue
		}
		zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
		if err != nil {
			t.Errorf("%s: zip.NewReader: %s", label, err)
			continue
		}

		got := map[string]string{}
		for _, f := range zr.File {
			r, err := f.Open()
			if err != nil {
				t.Errorf("%s: failed to open %q because %s", label, f.Name, err)
				continue
			}
			contents, err := ioutil.ReadAll(r)
			r.Close()
			if err != nil {
				t.Errorf("%s: Read(%q): %s", label, f.Name, err)
				continue
			}
			got[f.Name] = string(contents)
		}

		if !reflect.DeepEqual(got, test.want) {
			t.Errorf("%s: got %v, want %v", label, got, test.want)
		}
	}
}
