package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/log"
)

func TestCombine(t *testing.T) {
	tmp := t.TempDir()

	origin := filepath.Join(tmp, "origin")
	dir := filepath.Join(tmp, "combined-repo.git")

	// If we are running inside of the sourcegraph repo, use that as the
	// origin rather than using a small synthetic repo. In practice this means
	// when using go test we use the sg repo, in bazel we use a tiny synthetic
	// repo.
	if out, err := exec.Command("git", "rev-parse", "--show-toplevel").Output(); err == nil {
		origin = strings.TrimSpace(string(out))
		t.Log("using local repository instead of synthetic repo", origin)
	}

	setupPath := filepath.Join(tmp, "setup.sh")
	if err := os.WriteFile(setupPath, []byte(`#!/usr/bin/env bash

set -ex

## Setup origin repo if it doesn't exist.

if [ ! -d "$ORIGIN" ]; then
  mkdir -p "$ORIGIN"
  cd "$ORIGIN"
  git init

  git config user.email test@sourcegraph.com
  echo "foobar" > README.md
  git add README.md
  git commit -m "initial commit"
  echo "foobar" >> README.md
  git add README.md
  git commit -m "second commit"
fi

## Setup git-combine repo

mkdir -p "$DIR"
cd "$DIR"
git init --bare .

git config user.email test@sourcegraph.com
git remote add --no-tags sourcegraph "$ORIGIN"
git config --replace-all remote.origin.fetch '+HEAD:refs/remotes/sourcegraph/master'

git fetch --depth 100 sourcegraph
`), 0700); err != nil {
		t.Fatal(err)
	}

	cmd := exec.Command("sh", setupPath)
	cmd.Dir = tmp
	cmd.Env = append(
		os.Environ(),
		"DIR="+dir,
		"ORIGIN="+origin,
	)
	if testing.Verbose() {
		cmd.Stderr = os.Stderr
	}
	if err := cmd.Run(); err != nil {
		t.Fatal(err)
	}

	opt := Options{
		Logger: log.NoOp(),
	}
	if testing.Verbose() {
		opt.Logger = log.Scoped("test-git-combine")
	}

	if err := Combine(dir, opt); err != nil {
		t.Fatal(err)
	}

	// We only test that we have commits now. If we don't, git show will fail.
	cmd = exec.Command("git", "show")
	cmd.Dir = dir
	if err := cmd.Run(); err != nil {
		t.Fatal(err)
	}
}

func TestSanitizeMessage(t *testing.T) {
	cases := []struct {
		message   string
		wantTitle string
	}{{
		message:   "",
		wantTitle: "",
	}, {
		message:   "unchanged title\n\nbody",
		wantTitle: "unchanged title",
	}, {
		message:   "foo\n\nbar\nbaz",
		wantTitle: "foo",
	}, {
		message:   "foo\nbar\nbaz",
		wantTitle: "foo",
	}, {
		message:   "use decoration for active inlay hints link, support cusor decoration fyi @hediet, https://github.com/microsoft/vscode/issues/129528",
		wantTitle: "use decoration for active inlay hints link, support cusor decoration fyi",
	}, {
		message:   "naively @strip at the first @",
		wantTitle: "naively",
	}, {
		message:   "naively https://foo.com strip at the url",
		wantTitle: "naively",
	}}

	for _, tc := range cases {
		dir := "test"
		commit := &object.Commit{
			Message: tc.message,
			Hash:    plumbing.ZeroHash,
		}
		want := fmt.Sprintf("%s: %s\n\nCommit: %s\n", dir, tc.wantTitle, commit.Hash)
		got := sanitizeMessage(dir, commit)
		if d := cmp.Diff(want, got); d != "" {
			t.Errorf("unexpected for %q:\n%s", tc.message, d)
		}
	}
}
