package main

import (
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
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
		Limit:  100,
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
