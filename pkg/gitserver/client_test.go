package gitserver_test

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/gitserver/server"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/gitserver"
	"github.com/sourcegraph/sourcegraph/pkg/vcs/git"
)

func TestClient_Archive(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer((&server.Server{}).Handler())
	defer srv.Close()

	cli := gitserver.NewClient(&http.Client{})
	cli.Addrs = func(context.Context) []string {
		u, _ := url.Parse(srv.URL)
		return []string{u.Host}
	}

	repoWithDotGitDir := git.MakeTmpDir(t, "repo-with-dot-git-dir")
	if err := createRepoWithDotGitDir(repoWithDotGitDir); err != nil {
		t.Fatal(err)
	}

	gitCommands := []string{
		"mkdir dir1",
		"echo -n infile1 > dir1/file1",
		"touch --date=2006-01-02T15:04:05Z dir1 dir1/file1 || touch -t " + git.Times[0] + " dir1 dir1/file1",
		"git add dir1/file1",
		"GIT_COMMITTER_NAME=a GIT_COMMITTER_EMAIL=a@a.com GIT_COMMITTER_DATE=2006-01-02T15:04:05Z git commit -m commit1 --author='a <a@a.com>' --date 2006-01-02T15:04:05Z",
		"echo -n infile2 > 'file 2'",
		"touch --date=2014-05-06T19:20:21Z 'file 2' || touch -t " + git.Times[1] + " 'file 2'",
		"git add 'file 2'",
		"GIT_COMMITTER_NAME=a GIT_COMMITTER_EMAIL=a@a.com GIT_COMMITTER_DATE=2014-05-06T19:20:21Z git commit -m commit2 --author='a <a@a.com>' --date 2014-05-06T19:20:21Z",
	}
	tests := map[string]struct {
		repo gitserver.Repo
		want map[string]string
		err  error
	}{
		"git cmd": {
			repo: git.MakeGitRepository(t, gitCommands...),
			want: map[string]string{
				"dir1/":      "",
				"dir1/file1": "infile1",
				"file 2":     "infile2",
			},
		},
		"repo with .git dir": {
			repo: gitserver.Repo{Name: api.RepoName(repoWithDotGitDir), URL: repoWithDotGitDir},
			want: map[string]string{"file1": "hello\n", ".git/mydir/file2": "milton\n", ".git/mydir/": "", ".git/": ""},
		},
		"repo not found": {
			repo: gitserver.Repo{Name: api.RepoName("not-found")},
			err:  errors.New("repository does not exist: not-found"),
		},
	}

	ctx := context.Background()
	for label, test := range tests {
		rc, err := cli.Archive(ctx, test.repo, gitserver.ArchiveOptions{Treeish: "HEAD", Format: "zip"})
		if have, want := fmt.Sprint(err), fmt.Sprint(test.err); have != want {
			t.Errorf("%s: Archive: have err %v, want %v", label, have, want)
		}

		if rc == nil {
			continue
		}

		defer rc.Close()
		data, err := ioutil.ReadAll(rc)
		if err != nil {
			t.Errorf("%s: ReadAll: %s", label, err)
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

func createRepoWithDotGitDir(dir string) error {
	b64 := func(s string) string {
		b, err := base64.StdEncoding.DecodeString(s)
		if err != nil {
			panic(err)
		}
		return string(b)
	}

	// This repo was synthesized by hand to contain a file whose path is `.git/mydir/file2` (the Git
	// CLI will not let you create a file with a `.git` path component).
	//
	// The synthesized bad commit is:
	//
	// commit aa600fc517ea6546f31ae8198beb1932f13b0e4c (HEAD -> master)
	// Author: Quinn Slack <qslack@qslack.com>
	// 	Date:   Tue Jun 5 16:17:20 2018 -0700
	//
	// wip
	//
	// diff --git a/.git/mydir/file2 b/.git/mydir/file2
	// new file mode 100644
	// index 0000000..82b919c
	// --- /dev/null
	// +++ b/.git/mydir/file2
	// @@ -0,0 +1 @@
	// +milton
	files := map[string]string{
		"config": `
[core]
repositoryformatversion=0
filemode=true
`,
		"HEAD":              `ref: refs/heads/master`,
		"refs/heads/master": `aa600fc517ea6546f31ae8198beb1932f13b0e4c`,
		"objects/e7/9c5e8f964493290a409888d5413a737e8e5dd5": b64("eAFLyslPUrBgyMzLLMlMzOECACgtBOw="),
		"objects/ce/013625030ba8dba906f756967f9e9ca394464a": b64("eAFLyslPUjBjyEjNycnnAgAdxQQU"),
		"objects/82/b919c9c565d162c564286d9d6a2497931be47e": b64("eAFLyslPUjBnyM3MKcnP4wIAIw8ElA=="),
		"objects/e5/231c1d547df839dce09809e43608fe6c537682": b64("eAErKUpNVTAzYTAxAAIFvfTMEgbb8lmsKdJ+zz7ukeMOulcqZqOllmloYGBmYqKQlpmTashwjtFMlZl7xe2VbN/DptXPm7N4ipsXACOoGDo="),
		"objects/da/5ecc846359eaf23e8abe907b3125fdd7abdbc0": b64("eAErKUpNVTA2ZjA0MDAzMVFIy8xJNWJo2il58mjqxaSjKRq5c7NUpk+WflIHABZRD2I="),
		"objects/d0/01d287018593691c36042e1c8089fde7415296": b64("eAErKUpNVTA2ZjA0MDAzMVFIy8xJNWQ4x2imysy94vZKtu9h0+rnzVk8xc0LAP2TDiQ="),
		"objects/b4/009ecbf1eba01c5279f25840e2afc0d15f5005": b64("eAGdjdsJAjEQRf1OFdOAMpPN5gEitiBWEJIRBzcJu2b7N2IHfh24nMtJrRTpQA4PfWOGjEhZe4fk5zDZQGmyaDRT8ujDI7MzNOtgVdz7s21w26VWuC8xveC8vr+8/nBKrVxgyF4bJBfgiA5RjXUEO/9xVVKlS1zUB/JxNbA="),
		"objects/3d/779a05641b4ee6f1bc1e0b52de75163c2a2669": b64("eAErKUpNVTA2YjAxAAKF3MqUzCKGW3FnWpIjX32y69o3odpQ9e/11bcPAAAipRGQ"),
		"objects/aa/600fc517ea6546f31ae8198beb1932f13b0e4c": b64("eAGdjlkKAjEQBf3OKfoCSmfpLCDiFcQTZDodHHQWxwxe3xFv4FfBKx4UT8PQNzDa7doiAkLGataFXCg12lRYMEVM4qzHWMUz2eCjUXNeZGzQOdwkd1VLl1EzmZCqoehQTK6MRVMlRFJ5bbdpgcvajyNcH5nvcHy+vjz/cOBpOIEmE41D7xD2GBDVtm6BTf64qnc/qw9c4UKS"),
		"objects/e6/9de29bb2d1d6434b8b29ae775ad8c2e48c5391": b64("eAFLyslPUjBgAAAJsAHw"),
	}
	for name, data := range files {
		name = filepath.Join(dir, name)
		if err := os.MkdirAll(filepath.Dir(name), 0700); err != nil {
			return err
		}
		if err := ioutil.WriteFile(name, []byte(data), 0600); err != nil {
			return err
		}
	}
	return nil
}
