package main

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/langp"
)

func TestResolveFile(t *testing.T) {
	workspace, err := ioutil.TempDir("", "TestResolveFile")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(workspace)
	err = os.MkdirAll(filepath.Join(workspace, "gopath", "src"), os.FileMode(0700))
	if err != nil {
		t.Fatal(err)
	}
	// Resolve symbolic links, since git does
	workspace, err = filepath.EvalSymlinks(workspace)
	if err != nil {
		t.Fatal(err)
	}

	cases := []workspacePath{
		mkRepoPath(t, workspace, "github.com/foo/bar", "src/main.go"),
		mkRepoPath(t, workspace, "github.com/baz/bam", "main.go"),
	}
	// Test that dirs also can be resolved
	cases = append(cases, workspacePath{
		File: langp.File{
			Repo:   "github.com/foo/bar",
			Commit: cases[0].File.Commit,
			Path:   ".",
		},
		URI: "file:///gopath/src/github.com/foo/bar/.",
	})
	cases = append(cases, workspacePath{
		File: langp.File{
			Repo:   "github.com/foo/bar",
			Commit: cases[0].File.Commit,
			Path:   "src",
		},
		URI: "file:///gopath/src/github.com/foo/bar/src",
	})
	// Test that stdlib works
	cases = append(cases, workspacePath{
		File: langp.File{
			Repo:   "github.com/golang/go",
			Commit: "deadbeef",
			Path:   "src/strings/strings.go",
		},
		URI: "stdlib://deadbeef/src/strings/strings.go",
	})
	for _, c := range cases {
		got, err := resolveFile(workspace, cases[0].Repo, cases[0].Commit, c.URI)
		if err != nil {
			t.Fatalf("%+#v: %s", c, err)
		}
		want := &c.File
		if !reflect.DeepEqual(got, want) {
			t.Errorf("%s: got\n%+#v, want\n%+#v", c.URI, got, want)
		}
	}
}

type workspacePath struct {
	langp.File

	URI string
}

func mkRepoPath(t *testing.T, workspace, repo, path string) workspacePath {
	run := func(dir, cmd string, args ...string) string {
		c := exec.Command(cmd, args...)
		c.Dir = filepath.Join(workspace, "gopath", "src", dir)
		b, err := c.Output()
		if err != nil {
			t.Fatal(dir, cmd, strings.Join(args, " "), err)
		}
		return strings.TrimSpace(string(b))
	}
	run("", "mkdir", "-p", filepath.Dir(repo))
	run("", "git", "init", repo)
	run("", "mkdir", "-p", filepath.Dir(filepath.Join(repo, path)))
	run(repo, "touch", path)
	run(repo, "git", "add", path)
	run(repo, "git", "commit", "-m", "foo")
	commit := run(repo, "git", "rev-parse", "HEAD")
	return workspacePath{
		File: langp.File{
			Repo:   repo,
			Commit: commit,
			Path:   path,
		},
		URI: "file:///" + filepath.Join("gopath", "src", repo, path),
	}
}
