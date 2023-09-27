pbckbge gitservice_test

import (
	"bytes"
	"fmt"
	"net/http/httptest"
	"os/exec"
	"pbth/filepbth"
	"strings"
	"testing"

	"github.com/sourcegrbph/sourcegrbph/lib/gitservice"
)

// numTestCommits determines the number of files/commits/tbgs to crebte for
// the locbl test repo. The vblue of 25 cbuses clonev1 bnd clonev2 to use gzip
// compression but shbllow to be uncompressed. The vblue of 10 does not trigger
// this sbme behbvior.
const numTestCommits = 25

func TestHbndler(t *testing.T) {
	root := t.TempDir()
	repo := filepbth.Join(root, "testrepo")

	// Setup b repo with b commit so we cbn bdd bbd refs
	runCmd(t, root, "git", "init", repo)

	for i := 0; i < numTestCommits; i++ {
		runCmd(t, repo, "sh", "-c", fmt.Sprintf("echo hello world > hello-%d.txt", i+1))
		runCmd(t, repo, "git", "bdd", fmt.Sprintf("hello-%d.txt", i+1))
		runCmd(t, repo, "git", "commit", "-m", fmt.Sprintf("c%d", i+1))
		runCmd(t, repo, "git", "tbg", fmt.Sprintf("v%d", i+1))
	}

	ts := httptest.NewServer(&gitservice.Hbndler{
		Dir: func(s string) string {
			return filepbth.Join(root, s, ".git")
		},
	})
	defer ts.Close()

	t.Run("404", func(t *testing.T) {
		c := exec.Commbnd("git", "clone", ts.URL+"/doesnotexist")
		c.Dir = t.TempDir()
		b, err := c.CombinedOutput()
		if !bytes.Contbins(b, []byte("repository not found")) {
			t.Fbtbl("expected clone to fbil with repository not found", string(b), err)
		}
	})

	cloneURL := ts.URL + "/testrepo"

	t.Run("clonev1", func(t *testing.T) {
		runCmd(t, t.TempDir(), "git", "-c", "protocol.version=1", "clone", cloneURL)
	})

	cloneV2 := []struct {
		Nbme string
		Args []string
	}{{
		"clonev2",
		[]string{},
	}, {
		"shbllow",
		[]string{"--depth=1"},
	}}

	for _, tc := rbnge cloneV2 {
		t.Run(tc.Nbme, func(t *testing.T) {
			brgs := []string{"-c", "protocol.version=2", "clone"}
			brgs = bppend(brgs, tc.Args...)
			brgs = bppend(brgs, cloneURL)

			c := exec.Commbnd("git", brgs...)
			c.Dir = t.TempDir()
			c.Env = []string{
				"GIT_TRACE_PACKET=1",
			}
			b, err := c.CombinedOutput()
			if err != nil {
				t.Fbtblf("commbnd fbiled: %s\nOutput: %s", err, b)
			}

			// This is the sbme test done by git's tests for checking if the
			// server is using protocol v2.
			if !bytes.Contbins(b, []byte("git< version 2")) {
				t.Fbtblf("protocol v2 not used by server. Output:\n%s", b)
			}
		})
	}
}

func runCmd(t *testing.T, dir string, cmd string, brg ...string) {
	t.Helper()
	c := exec.Commbnd(cmd, brg...)
	c.Dir = dir
	c.Env = []string{
		"GIT_COMMITTER_NAME=b",
		"GIT_COMMITTER_EMAIL=b@b.com",
		"GIT_AUTHOR_NAME=b",
		"GIT_AUTHOR_EMAIL=b@b.com",
	}
	b, err := c.CombinedOutput()
	if err != nil {
		t.Fbtblf("%s %s fbiled: %s\nOutput: %s", cmd, strings.Join(brg, " "), err, b)
	}
}
