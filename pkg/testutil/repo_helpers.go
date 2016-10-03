package testutil

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"context"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/executil"
	"sourcegraph.com/sourcegraph/sourcegraph/services/backend"
)

// CreateRepo creates a new repo. Callers must call the returned
// done() func when done (if err is non-nil) to free up resources.
func CreateRepo(t *testing.T, ctx context.Context, repoURI string, mirror bool) (repo *sourcegraph.Repo, done func(), err error) {
	op := &sourcegraph.ReposCreateOp_NewRepo{
		URI: repoURI,
	}

	if mirror {
		s := httptest.NewServer(trivialGitRepoHandler)
		op.CloneURL, done = s.URL, s.Close
		op.Mirror = true
	}
	if done == nil {
		done = func() {} // no-op
	}

	repo, err = backend.Repos.Create(ctx, &sourcegraph.ReposCreateOp{Op: &sourcegraph.ReposCreateOp_New{New: op}})
	if err != nil {
		done()
		return nil, done, err
	}

	return repo, done, nil
}

// CreateEmptyMirrorRepo creates an empty mirror repo.
func CreateEmptyMirrorRepo(t *testing.T, ctx context.Context, repoURI string) error {
	op := &sourcegraph.ReposCreateOp_NewRepo{
		URI:      repoURI,
		CloneURL: "https://" + repoURI,
		Mirror:   true,
	}
	_, err := backend.Repos.Create(ctx, &sourcegraph.ReposCreateOp{Op: &sourcegraph.ReposCreateOp_New{New: op}})
	return err
}

// PushRepo pushes sample commits to the remote specified.
// If files is specified, it is treated as a map of filenames to file contents.
// If files is nil, a default set of some text files is used. All files are
// committed in the same commit.
// If deleteBranch is true it will push the commits to another branch and
// then atttempt to delete the branch.
func PushRepo(t *testing.T, ctx context.Context, pushURL, cloneURL string, files map[string]string, deleteBranch bool) error {
	if cloneURL == "" {
		return fmt.Errorf("PushRepo can't be called with `cloneURL` unset.")
	}

	// Clone the repository.
	tmpDir, err := ioutil.TempDir("", "")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir)

	dir := filepath.Join(tmpDir, "testrepo")
	cmd := exec.Command("git", "clone", cloneURL, dir)
	cmd.Dir = tmpDir
	prepGitCommand(cmd)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("exec %q failed: %s\n%s", cmd.Args, err, out)
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

	cmd = exec.Command("git", "commit", "-m", "hello", "--author", "a <a@a.com>", "--date", "2006-01-02T15:04:05Z")
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
	out, err := executil.CmdCombinedOutputWithTimeout(time.Second*10, cmd)
	logCmdOutut(t, cmd, out)
	if err != nil {
		return fmt.Errorf("exec %q failed: %s\n%s", cmd.Args, err, out)
	}

	if deleteBranch {
		env := cmd.Env
		// Create new branch.
		cmd = exec.Command("git", "push", pushURL, "master:tmpbranch")
		cmd.Env = env
		cmd.Dir = dir
		out, err := executil.CmdCombinedOutputWithTimeout(time.Second*10, cmd)
		logCmdOutut(t, cmd, out)
		if err != nil {
			return fmt.Errorf("exec %q failed: %s\n%s", cmd.Args, err, out)
		}

		// Delete the branch.
		cmd = exec.Command("git", "push", pushURL, ":tmpbranch")
		cmd.Env = env
		cmd.Dir = dir
		out, err = executil.CmdCombinedOutputWithTimeout(time.Second*10, cmd)
		logCmdOutut(t, cmd, out)
		if err != nil {
			return fmt.Errorf("exec %q failed: %s\n%s", cmd.Args, err, out)
		}
	}
	return nil
}

// CloneRepo tests cloning from the clone URL.
// If emptyFetch is true it performs a fetch right after a clone to test a fetch
// that does not go through the pack negotiation phase of the protocol.
func CloneRepo(t *testing.T, cloneURL, dir string, args []string, emptyFetch bool) (err error) {
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
	cmd.Args = append(cmd.Args, cloneURL, "testrepo")
	cmd.Env = append(cmd.Env, "GIT_ASKPASS=true") // disable password prompt
	cmd.Stdin = bytes.NewReader([]byte("\n"))
	cmd.Dir = dir
	prepGitCommand(cmd)
	out, err := executil.CmdCombinedOutputWithTimeout(time.Second*10, cmd)
	logCmdOutut(t, cmd, out)
	if err != nil {
		return fmt.Errorf("exec %q failed: %s\n%s", cmd.Args, err, out)
	}
	if emptyFetch {
		env := cmd.Env
		cmd := exec.Command("git", "fetch")
		cmd.Env = env
		cmd.Stdin = bytes.NewReader([]byte("\n"))
		cmd.Dir = filepath.Join(dir, "testrepo")
		out, err := executil.CmdCombinedOutputWithTimeout(time.Second*10, cmd)
		logCmdOutut(t, cmd, out)
		if err != nil {
			return fmt.Errorf("exec %q failed: %s\n%s", cmd.Args, err, out)
		}
	}
	return nil
}

// prepGitCommand adds environment variables for running a git command.
func prepGitCommand(cmd *exec.Cmd) *exec.Cmd {
	// Avoid using git's system/global configurations.
	cmd.Env = append(cmd.Env, "GIT_CONFIG_NOSYSTEM=1", "HOME=/doesnotexist", "XDG_CONFIG_HOME=/doesnotexist")
	// Debugging.
	cmd.Env = append(cmd.Env, "GIT_TRACE=1")
	cmd.Env = append(cmd.Env, "GIT_CURL_VERBOSE=1")
	cmd.Env = append(cmd.Env, "GIT_TRACE_PACKET=1")
	cmd.Env = append(cmd.Env, "GIT_TRACE_PACK_ACCESS=1")
	return cmd
}

func logCmdOutut(t *testing.T, cmd *exec.Cmd, out []byte) {
	t.Logf(">>> START - %s", strings.Join(cmd.Args, " "))
	t.Logf("=== ENV - %v", cmd.Env)
	t.Log(string(out))
	t.Logf(">>> END - %s", strings.Join(cmd.Args, " "))
}
