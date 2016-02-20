package testutil

import (
	"bytes"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"golang.org/x/crypto/ssh"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"src.sourcegraph.com/sourcegraph/auth/authutil"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/pkg/vcs"
	"src.sourcegraph.com/sourcegraph/util/executil"
)

func EnsureRepoExists(t *testing.T, ctx context.Context, repoURI string) {
	cl, _ := sourcegraph.NewClientFromContext(ctx)

	repo, err := cl.Repos.Get(ctx, &sourcegraph.RepoSpec{URI: repoURI})
	if err != nil {
		t.Fatalf("repo %s does not exist: %s", repoURI, err)
	}

	// Make sure the repo has been cloned to vcsstore.
	repoRevSpec := sourcegraph.RepoRevSpec{RepoSpec: sourcegraph.RepoSpec{URI: repoURI}, Rev: repo.DefaultBranch}
	getCommitWithRefreshAndRetry(t, ctx, repoRevSpec)
}

// getCommitWithRefreshAndRetry tries to get a repository commit. If
// it doesn't exist, it triggers a refresh of the repo's VCS data and
// then retries (until maxGetCommitVCSRefreshWait has elapsed).
func getCommitWithRefreshAndRetry(t *testing.T, ctx context.Context, repoRevSpec sourcegraph.RepoRevSpec) *vcs.Commit {
	cl, _ := sourcegraph.NewClientFromContext(ctx)

	wait := time.Second * 9 * ciFactor

	timeout := time.After(wait)
	done := make(chan struct{})
	var commit *vcs.Commit
	var err error
	go func() {
		refreshTriggered := false
		for {
			commit, err = cl.Repos.GetCommit(ctx, &repoRevSpec)

			// Keep retrying if it's a NotFound, but stop trying if we succeeded, or if it's some other
			// error.
			if err == nil || grpc.Code(err) != codes.NotFound {
				break
			}

			if !refreshTriggered {
				if _, err = cl.MirrorRepos.RefreshVCS(ctx, &sourcegraph.MirrorReposRefreshVCSOp{Repo: repoRevSpec.RepoSpec}); err != nil {
					err = fmt.Errorf("failed to trigger VCS refresh for repo %s: %s", repoRevSpec.URI, err)
					break
				}
				t.Logf("repo %s revision %s not on remote; triggered refresh of VCS data, waiting %s", repoRevSpec.URI, repoRevSpec.Rev, wait)
				refreshTriggered = true
			}
			time.Sleep(time.Second)
		}
		done <- struct{}{}
	}()
	select {
	case <-done:
		if err != nil {
			t.Fatal(err)
		}
		return commit
	case <-timeout:
		t.Fatalf("repo %s revision %s not found on remote, even after triggering a VCS refresh and waiting %s (vcsstore should not have taken so long)", repoRevSpec.URI, repoRevSpec.Rev, wait)
		panic("unreachable")
	}
}

// CreateRepo creates a new repo. Callers must call the returned
// done() func when done (if err is non-nil) to free up resources.
func CreateRepo(t *testing.T, ctx context.Context, repoURI string, mirror bool) (repo *sourcegraph.Repo, done func(), err error) {
	cl, _ := sourcegraph.NewClientFromContext(ctx)

	op := &sourcegraph.ReposCreateOp{
		URI: repoURI,
		VCS: "git",
	}

	if mirror {
		s := httptest.NewServer(trivialGitRepoHandler)
		op.CloneURL, done = s.URL, s.Close
		op.Mirror = true
	}
	if done == nil {
		done = func() {} // no-op
	}

	repo, err = cl.Repos.Create(ctx, op)
	if err != nil {
		done()
		return nil, done, err
	}

	return repo, done, nil
}

// CreateAndPushRepo is short-handed for:
//
//  CreateAndPushRepoFiles(t, ctx, repoURI, nil)
//
func CreateAndPushRepo(t *testing.T, ctx context.Context, repoURI string) (repo *sourcegraph.Repo, commitID string, done func(), err error) {
	return CreateAndPushRepoFiles(t, ctx, repoURI, nil)
}

// CreateAndPushRepoFiles creates and pushes sample commits to a repo. Callers
// must call the returned done() func when done (if err is non-nil) to free up
// resources.
func CreateAndPushRepoFiles(t *testing.T, ctx context.Context, repoURI string, files map[string]string) (repo *sourcegraph.Repo, commitID string, done func(), err error) {
	repo, done, err = CreateRepo(t, ctx, repoURI, false)
	if err != nil {
		return nil, "", nil, err
	}

	authedCloneURL, err := authutil.AddSystemAuthToURL(ctx, "internal:write", repo.HTTPCloneURL)
	if err != nil {
		return nil, "", nil, err
	}

	if err := PushRepo(ctx, authedCloneURL, authedCloneURL, nil, files); err != nil {
		return nil, "", nil, err
	}

	cl, _ := sourcegraph.NewClientFromContext(ctx)
	commit, err := cl.Repos.GetCommit(ctx, &sourcegraph.RepoRevSpec{
		RepoSpec: sourcegraph.RepoSpec{URI: repo.URI},
		Rev:      "master",
	})
	if err != nil {
		return nil, "", nil, err
	}
	return repo, string(commit.ID), done, nil
}

// PushRepo pushes sample commits to the remote specified.
// If files is specified, it is treated as a map of filenames to file contents.
// If files is nil, a default set of some text files is used. All files are
// committed in the same commit.
func PushRepo(ctx context.Context, pushURL, cloneURL string, key *rsa.PrivateKey, files map[string]string) error {
	if cloneURL == "" {
		return fmt.Errorf("PushRepo can't be called with both `cloneURL` unset.")
	}

	// Clone the repository.
	tmpDir, err := ioutil.TempDir("", "")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir)

	dir := filepath.Join(tmpDir, "testrepo")
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	if err := CloneRepo(cloneURL, dir, key, nil); err != nil {
		return err
	}

	// Add files and make a commit.
	if files == nil {
		files = map[string]string{"myfile.txt": "a"}
	}
	for path, data := range files {
		if err := os.MkdirAll(filepath.Dir(filepath.Join(dir, path)), 0700); err != nil {
			return err
		}
		if err := ioutil.WriteFile(filepath.Join(dir, path), []byte(data), 0700); err != nil {
			return err
		}
		cmd := exec.Command("git", "add", path)
		cmd.Dir = dir
		prepGitCommand(cmd)
		if out, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("exec %q failed: %s\n%s", cmd.Args, err, out)
		}
	}

	cmd := exec.Command("git", "commit", "-m", "hello", "--author", "a <a@a.com>", "--date", "2006-01-02T15:04:05Z")
	cmd.Env = append(cmd.Env, "GIT_COMMITTER_NAME=a", "GIT_COMMITTER_EMAIL=a@a.com", "GIT_COMMITTER_DATE=2006-01-02T15:04:05Z")
	cmd.Dir = dir
	prepGitCommand(cmd)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("exec %q failed: %s\n%s", cmd.Args, err, out)
	}

	// Push the commits.
	if pushURL == "" {
		pushURL = "origin"
	}
	cmd = exec.Command("git", "push", pushURL, "master")
	cmd.Env = append(cmd.Env, "GIT_ASKPASS=true") // disable password prompt
	cmd.Dir = dir
	prepGitCommand(cmd)
	if key != nil {
		// Attempting to push over SSH.
		if _, err := prepGitSSHCommand(cmd, dir, key); err != nil {
			return err
		}
	}
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("exec %q failed: %s\n%s", cmd.Args, err, out)
	}
	return nil
}

func CloneRepo(cloneURL, dir string, key *rsa.PrivateKey, args []string) (err error) {
	if dir == "" {
		var err error
		dir, err = ioutil.TempDir("", "")
		if err != nil {
			return err
		}
		defer os.RemoveAll(dir)
	}
	cmd := exec.Command("git", "clone")
	cmd.Args = append(cmd.Args, args...)
	cmd.Args = append(cmd.Args, cloneURL, dir)
	cmd.Dir = dir
	cmd.Env = append(cmd.Env, "GIT_ASKPASS=true") // disable password prompt
	prepGitCommand(cmd)
	if key != nil {
		// Attempting to clone over SSH.
		if _, err := prepGitSSHCommand(cmd, dir, key); err != nil {
			return err
		}
	}

	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf

	defer func() {
		err = wrapExecErr(err, cmd, buf.String())
	}()

	// Bypass password prompts by sending in a password usually.
	in, err := cmd.StdinPipe()
	if err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	if _, err := in.Write([]byte("\n")); err != nil {
		return err
	}
	if err := in.Close(); err != nil {
		return err
	}

	if err := executil.CmdWaitWithTimeout(5*time.Second, cmd); err != nil {
		return err
	}
	return nil
}

// prepGitSSHCommand performs the necessary configurations to execute an git
// command using the provided RSA key.
func prepGitSSHCommand(cmd *exec.Cmd, dir string, key *rsa.PrivateKey) (*exec.Cmd, error) {
	sshDir := filepath.Join(dir, ".ssh")
	if err := os.Mkdir(sshDir, 0700); err != nil {
		return cmd, err
	}

	idFile := filepath.Join(sshDir, "sshkey")

	// Write public key.
	sshPublicKey, err := ssh.NewPublicKey(&key.PublicKey)
	if err != nil {
		return cmd, err
	}
	publicKey := ssh.MarshalAuthorizedKey(sshPublicKey)
	if err := ioutil.WriteFile(idFile+".pub", publicKey, 0600); err != nil {
		return cmd, err
	}

	// Write private key.
	keyPrivatePEM := pem.EncodeToMemory(&pem.Block{
		Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	if err := ioutil.WriteFile(idFile, keyPrivatePEM, 0600); err != nil {
		return cmd, err
	}

	// Generate the necessary SSH command.
	// NOTE: GIT_SSH_COMMAND requires git version 2.3+, so we must
	// use GIT_SSH.
	gitSSH := filepath.Join(sshDir, "gitssh")
	if err := ioutil.WriteFile(gitSSH, []byte(fmt.Sprintf("#!/bin/sh\nssh -i %s -vvvv -F /dev/null -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null \"$@\"\n", idFile)), 0500); err != nil {
		return cmd, err
	}
	cmd.Env = append(cmd.Env, fmt.Sprintf("GIT_SSH=%s", gitSSH))
	return cmd, err
}

// prepGitCommand adds environment variables for running a git command.
func prepGitCommand(cmd *exec.Cmd) *exec.Cmd {
	// Avoid using git's system/global configurations.
	cmd.Env = append(cmd.Env, "GIT_CONFIG_NOSYSTEM=1", "HOME=/doesnotexist", "XDG_CONFIG_HOME=/doesnotexist")
	// Debugging.
	//cmd.Env = append(cmd.Env, "GIT_TRACE=1")
	//cmd.Env = append(cmd.Env, "GIT_CURL_VERBOSE=1")
	return cmd
}
