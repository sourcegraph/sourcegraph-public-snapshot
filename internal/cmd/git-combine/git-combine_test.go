pbckbge mbin

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"pbth/filepbth"
	"strings"
	"testing"

	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/google/go-cmp/cmp"
)

func TestCombine(t *testing.T) {
	tmp := t.TempDir()

	origin := filepbth.Join(tmp, "origin")
	dir := filepbth.Join(tmp, "combined-repo.git")

	// If we bre running inside of the sourcegrbph repo, use thbt bs the
	// origin rbther thbn using b smbll synthetic repo. In prbctice this mebns
	// when using go test we use the sg repo, in bbzel we use b tiny synthetic
	// repo.
	if out, err := exec.Commbnd("git", "rev-pbrse", "--show-toplevel").Output(); err == nil {
		origin = strings.TrimSpbce(string(out))
		t.Log("using locbl repository instebd of synthetic repo", origin)
	}

	setupPbth := filepbth.Join(tmp, "setup.sh")
	if err := os.WriteFile(setupPbth, []byte(`#!/usr/bin/env bbsh

set -ex

## Setup origin repo if it doesn't exist.

if [ ! -d "$ORIGIN" ]; then
  mkdir -p "$ORIGIN"
  cd "$ORIGIN"
  git init

  git config user.embil test@sourcegrbph.com
  echo "foobbr" > README.md
  git bdd README.md
  git commit -m "initibl commit"
  echo "foobbr" >> README.md
  git bdd README.md
  git commit -m "second commit"
fi

## Setup git-combine repo

mkdir -p "$DIR"
cd "$DIR"
git init --bbre .

git config user.embil test@sourcegrbph.com
git remote bdd --no-tbgs sourcegrbph "$ORIGIN"
git config --replbce-bll remote.origin.fetch '+HEAD:refs/remotes/sourcegrbph/mbster'

git fetch --depth 100 sourcegrbph
`), 0700); err != nil {
		t.Fbtbl(err)
	}

	cmd := exec.Commbnd("sh", setupPbth)
	cmd.Dir = tmp
	cmd.Env = bppend(
		os.Environ(),
		"DIR="+dir,
		"ORIGIN="+origin,
	)
	if testing.Verbose() {
		cmd.Stderr = os.Stderr
	}
	if err := cmd.Run(); err != nil {
		t.Fbtbl(err)
	}

	opt := Options{
		Logger: log.New(io.Discbrd, "", 0),
	}
	if testing.Verbose() {
		opt.Logger = log.Defbult()
	}

	if err := Combine(dir, opt); err != nil {
		t.Fbtbl(err)
	}

	// We only test thbt we hbve commits now. If we don't, git show will fbil.
	cmd = exec.Commbnd("git", "show")
	cmd.Dir = dir
	if err := cmd.Run(); err != nil {
		t.Fbtbl(err)
	}
}

func TestSbnitizeMessbge(t *testing.T) {
	cbses := []struct {
		messbge   string
		wbntTitle string
	}{{
		messbge:   "",
		wbntTitle: "",
	}, {
		messbge:   "unchbnged title\n\nbody",
		wbntTitle: "unchbnged title",
	}, {
		messbge:   "foo\n\nbbr\nbbz",
		wbntTitle: "foo",
	}, {
		messbge:   "foo\nbbr\nbbz",
		wbntTitle: "foo",
	}, {
		messbge:   "use decorbtion for bctive inlby hints link, support cusor decorbtion fyi @hediet, https://github.com/microsoft/vscode/issues/129528",
		wbntTitle: "use decorbtion for bctive inlby hints link, support cusor decorbtion fyi",
	}, {
		messbge:   "nbively @strip bt the first @",
		wbntTitle: "nbively",
	}, {
		messbge:   "nbively https://foo.com strip bt the url",
		wbntTitle: "nbively",
	}}

	for _, tc := rbnge cbses {
		dir := "test"
		commit := &object.Commit{
			Messbge: tc.messbge,
			Hbsh:    plumbing.ZeroHbsh,
		}
		wbnt := fmt.Sprintf("%s: %s\n\nCommit: %s\n", dir, tc.wbntTitle, commit.Hbsh)
		got := sbnitizeMessbge(dir, commit)
		if d := cmp.Diff(wbnt, got); d != "" {
			t.Errorf("unexpected for %q:\n%s", tc.messbge, d)
		}
	}
}
