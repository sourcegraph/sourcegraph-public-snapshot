package worker_test

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"sourcegraph.com/sourcegraph/go-vcs/vcs"
	"sourcegraph.com/sourcegraph/go-vcs/vcs/gitcmd"
	"src.sourcegraph.com/sourcegraph/sgx"
	"src.sourcegraph.com/sourcegraph/util/httputil"
	"src.sourcegraph.com/sourcegraph/util/testutil"
	"src.sourcegraph.com/sourcegraph/worker"
)

func init() {
	gitcmd.InsecureSkipCheckVerifySSH = true
	sgx.SetVerbose(true)
}

func TestPrepBuildDir_sshGitRepo_md(t *testing.T) {
	t.Parallel()

	tmpDir, err := ioutil.TempDir("", "sgx-prep-build-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)
	commitID, err := testutil.InitTrivialGitRepo(tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	s, remoteOpts, err := testutil.NewGitSSHServer(filepath.Dir(tmpDir))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	gitURL := s.GitURL + "/" + filepath.Base(tmpDir)

	cloneParentDir, err := ioutil.TempDir("", "sgx-prep-build-test-clone")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(cloneParentDir)
	cloneDir := filepath.Join(cloneParentDir, "clonedir")

	if err := worker.PrepBuildDir("git", gitURL, "", "", cloneDir, commitID, *remoteOpts); err != nil {
		t.Fatal(err)
	}
	if err := worker.PrepBuildDir("git", gitURL, "", "", cloneDir, commitID, *remoteOpts); err != nil {
		t.Fatal(err)
	}
}

func TestPrepBuildDir_httpGitRepo_lg(t *testing.T) {
	t.Parallel()

	s := httptest.NewServer(testutil.TrivialGitRepoHandler)
	defer s.Close()

	gitURL := s.URL
	commitID := testutil.TrivialGitRepoHandlerHeadCommitID

	cloneParentDir, err := ioutil.TempDir("", "sgx-prep-build-test-clone")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(cloneParentDir)
	cloneDir := filepath.Join(cloneParentDir, "clonedir")

	if err := worker.PrepBuildDir("git", gitURL, "", "", cloneDir, commitID, vcs.RemoteOpts{}); err != nil {
		t.Fatal(err)
	}
}

func TestPrepBuildDir_httpBasicAuthGitRepo_lg(t *testing.T) {
	t.Parallel()

	// Use StatusUnauthorized because we test invalid auth, valid auth and then invalid auth again.
	// Using StatusForbidden would cause the successful auth cases to fail after invalid auth was provided.
	// I'm not seeing issues with curl prompting for login (the original reason StatusForbidden was introduced) on my setup;
	// if that happens again, it should be fixed in another way.
	s := httptest.NewServer(httputil.BasicAuth("u", "p", http.StatusUnauthorized, testutil.TrivialGitRepoHandler))
	defer s.Close()

	cloneParentDir, err := ioutil.TempDir("", "sgx-prep-build-test-clone")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(cloneParentDir)
	cloneDir := filepath.Join(cloneParentDir, "clonedir")

	gitURL := s.URL
	commitID := testutil.TrivialGitRepoHandlerHeadCommitID

	// Need to wipe clone dir or else PrepBuildDir will complain that
	// it already exists.
	wipeCloneDir := func() {
		if err := os.RemoveAll(cloneParentDir); err != nil {
			t.Fatal(err)
		}
	}

	// No credentials should fail due to lack of auth.
	if err := os.Setenv("GIT_ASKPASS", "/bin/echo"); err != nil {
		t.Fatal(err)
	}
	if err := worker.PrepBuildDir("git", gitURL, "", "", cloneDir, commitID, vcs.RemoteOpts{}); err == nil {
		t.Fatal("succeeded without auth")
	}
	if err := os.Unsetenv("GIT_ASKPASS"); err != nil {
		t.Fatal(err)
	}

	// Bad credentials should fail.
	if err := worker.PrepBuildDir("git", gitURL, "baduser", "badpw", cloneDir, commitID, vcs.RemoteOpts{}); err == nil {
		t.Fatal("succeeded with bad auth")
	}

	// Succeeds with correct auth.
	if err := worker.PrepBuildDir("git", gitURL, "u", "p", cloneDir, commitID, vcs.RemoteOpts{}); err != nil {
		t.Fatal(err)
	}

	// No credentials should fail (after prepping with correct auth).
	wipeCloneDir()
	if err := os.Setenv("GIT_ASKPASS", "/bin/echo"); err != nil {
		t.Fatal(err)
	}
	if err := worker.PrepBuildDir("git", gitURL, "", "", cloneDir, commitID, vcs.RemoteOpts{}); err == nil {
		t.Fatal("succeeded without auth after prepping with correct auth")
	}
	if err := os.Unsetenv("GIT_ASKPASS"); err != nil {
		t.Fatal(err)
	}

	// Bad credentials should fail even after prepping with correct auth.
	if err := worker.PrepBuildDir("git", gitURL, "baduser", "badpw", cloneDir, commitID, vcs.RemoteOpts{}); err == nil {
		t.Fatal("succeeded with bad auth after prepping with correct auth")
	}
}

func TestPrepBuildDir_httpsHgRepo_lg(t *testing.T) {
	t.Skip("TODO(sqs) - make this a self-contained hg host that doesn't require an external server")

	t.Parallel()

	const (
		hgURL    = "https://bitbucket.org/sqs/go-vcs-hgtest"
		commitID = "bcc18e4692162e616cc6165589a24be4ea40e3d2"
	)

	cloneParentDir, err := ioutil.TempDir("", "sgx-prep-build-test-clone")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(cloneParentDir)
	cloneDir := filepath.Join(cloneParentDir, "clonedir")

	if err := worker.PrepBuildDir("hg", hgURL, "", "", cloneDir, commitID, vcs.RemoteOpts{}); err != nil {
		t.Fatal(err)
	}
}
