package vfsutil

import (
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestGitRepoVFS(t *testing.T) {
	if testing.Short() {
		t.Skip("skip network-intensive test")
	}

	fs := &GitRepoVFS{
		CloneURL: "git://github.com/gorilla/schema",
		Rev:      "0164a00ab4cd01d814d8cd5bf63fd9fcea30e23b",
	}
	want := map[string]string{
		"/LICENSE":         "...",
		"/README.md":       "schema...",
		"/cache.go":        "// Copyright...",
		"/converter.go":    "// Copyright...",
		"/decoder.go":      "// Copyright...",
		"/decoder_test.go": "// Copyright...",
		"/doc.go":          "// Copyright...",
		"/.travis.yml":     "...",
	}

	testVFS(t, fs, want)
}

func TestGitRepoVFS_subtree(t *testing.T) {
	if testing.Short() {
		t.Skip("skip network-intensive test")
	}

	// Any public repo will work.
	fs := &GitRepoVFS{
		CloneURL: "git://github.com/gorilla/rpc",
		Rev:      "e592e2e099465ae27afa66ec089d570904cd2d53",
		Subtree:  "protorpc",
	}
	want := map[string]string{
		"/doc.go":           "// Copyright 2...",
		"/protorpc_test.go": "// Copyright 2...",
		"/server.go":        "// Copyright 2...",
	}

	testVFS(t, fs, want)
}

func TestGitRepoVFS_cache(t *testing.T) {
	// We use a different gitArchiveBasePath to ensure it is empty
	{
		d, err := ioutil.TempDir("", "vfsutil_test")
		if err != nil {
			t.Fatal(err)
		}
		defer os.RemoveAll(d)
		defer func(orig string) { gitArchiveBasePath = orig }(gitArchiveBasePath)
		gitArchiveBasePath = d
	}

	cloneURL, err := ioutil.TempDir("", "vfsutil_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(cloneURL)

	runCmds := func(cmds ...string) string {
		var out []byte
		var err error
		cmds = append(cmds, "git rev-parse HEAD")
		for _, cmd := range cmds {
			c := exec.Command("bash", "-c", cmd)
			c.Dir = cloneURL
			c.Env = []string{
				"GIT_COMMITTER_NAME=a",
				"GIT_COMMITTER_EMAIL=a@a.com",
				"GIT_AUTHOR_NAME=a",
				"GIT_AUTHOR_EMAIL=a@a.com",
			}
			out, err = c.CombinedOutput()
			if err != nil {
				t.Fatalf("Command %q failed. Output was:\n\n%s", cmd, out)
			}
		}
		return strings.TrimSpace(string(out))
	}

	rev1 := runCmds(
		"git init",
		"echo -n text1 > file",
		"git add file",
		"git commit -m msg",
	)
	rev2 := runCmds(
		"echo -n text2 > file",
		"git add file",
		"git commit -m msg",
	)

	// On first attempt the cache should be empty, so this tests we can
	// clone
	fs := &GitRepoVFS{
		CloneURL: cloneURL,
		Rev:      rev1,
	}
	want := map[string]string{
		"/file": "text1",
	}
	testVFS(t, fs, want)

	// Our second attempt should have the commit already. So this tests we
	// just directly use it
	fs = &GitRepoVFS{
		CloneURL: cloneURL,
		Rev:      rev2,
	}
	want = map[string]string{
		"/file": "text2",
	}
	testVFS(t, fs, want)

	// Now we add a commit to test the update path
	rev3 := runCmds(
		"echo -n text3 > file",
		"git add file",
		"git commit -m msg",
	)
	fs = &GitRepoVFS{
		CloneURL: cloneURL,
		Rev:      rev3,
	}
	want = map[string]string{
		"/file": "text3",
	}
	testVFS(t, fs, want)
}
