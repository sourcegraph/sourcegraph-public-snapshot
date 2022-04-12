package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/google/go-cmp/cmp"
)

func TestCombine(t *testing.T) {
	tmp := t.TempDir()

	dir := filepath.Join(tmp, "combined-repo.git")

	setupPath := filepath.Join(tmp, "setup.sh")
	if err := os.WriteFile(setupPath, []byte(`#!/usr/bin/env bash

set -ex

repo=$(git rev-parse --show-toplevel)

mkdir -p "$DIR"
cd "$DIR"
git init --bare .

git remote add --no-tags sourcegraph "$repo"
git config --replace-all remote.origin.fetch '+HEAD:refs/remotes/sourcegraph/master'

git fetch --depth 100 sourcegraph
`), 0700); err != nil {
		t.Fatal(err)
	}

	cmd := exec.Command("sh", setupPath)
	cmd.Env = append(os.Environ(), "DIR="+dir)
	if testing.Verbose() {
		cmd.Stderr = os.Stderr
	}
	if err := cmd.Run(); err != nil {
		t.Fatal(err)
	}

	opt := Options{
		Logger: log.New(io.Discard, "", 0),
	}
	if testing.Verbose() {
		opt.Logger = log.Default()
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
