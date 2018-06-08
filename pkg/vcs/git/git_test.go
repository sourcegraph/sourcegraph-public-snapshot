package git_test

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/gitserver"
	"github.com/sourcegraph/sourcegraph/pkg/vcs/git"
)

var times = []string{
	appleTime("2006-01-02T15:04:05Z"),
	appleTime("2014-05-06T19:20:21Z"),
}

var nonexistentCommitID = api.CommitID(strings.Repeat("a", 40))

var ctx = context.Background()

// initGitRepository initializes a new Git repository and runs cmds in a new
// temporary directory (returned as dir).
func initGitRepository(t testing.TB, cmds ...string) string {
	dir := initGitRepositoryWorkingCopy(t, cmds...)
	makeGitRepositoryBare(t, dir)
	return dir
}

func initGitRepositoryWorkingCopy(t testing.TB, cmds ...string) (dir string) {
	dir = makeTmpDir(t, "git")
	cmds = append([]string{"git init"}, cmds...)
	for _, cmd := range cmds {
		c := exec.Command("bash", "-c", cmd)
		c.Dir = dir
		out, err := c.CombinedOutput()
		if err != nil {
			t.Fatalf("Command %q failed. Output was:\n\n%s", cmd, out)
		}
	}
	return dir
}

func makeGitRepositoryBare(t testing.TB, dir string) {
	c := exec.Command("git", "config", "--bool", "core.bare", "true")
	c.Dir = dir
	out, err := c.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to convert to bare repo: %s\nOut: %s", err, out)
	}
	wc := dir + "-workingcopy"
	err = os.Rename(dir, wc)
	if err != nil {
		t.Fatalf("Failed to convert to bare repo: %s", err)
	}
	err = os.Rename(filepath.Join(wc, ".git"), dir)
	if err != nil {
		t.Fatalf("Failed to convert to bare repo: %s", err)
	}
}

// makeGitRepository calls initGitRepository to create a new Git repository and returns a handle to
// it.
func makeGitRepository(t testing.TB, cmds ...string) gitserver.Repo {
	dir := initGitRepository(t, cmds...)
	return gitserver.Repo{Name: api.RepoURI(dir), URL: dir}
}

func commitsEqual(a, b *git.Commit) bool {
	if (a == nil) != (b == nil) {
		return false
	}
	if a.Author.Date != b.Author.Date {
		return false
	}
	a.Author.Date = b.Author.Date
	if ac, bc := a.Committer, b.Committer; ac != nil && bc != nil {
		if ac.Date != bc.Date {
			return false
		}
		ac.Date = bc.Date
	} else if !(ac == nil && bc == nil) {
		return false
	}
	return reflect.DeepEqual(a, b)
}

func mustParseTime(layout, value string) time.Time {
	tm, err := time.Parse(layout, value)
	if err != nil {
		panic(err.Error())
	}
	return tm
}

func appleTime(t string) string {
	ti, _ := time.Parse(time.RFC3339, t)
	return ti.Local().Format("200601021504.05")
}

// Computes hash of last commit in a given repo dir
// On Windows, content of a "link file" differs based on the tool that produced it.
// For example:
// - Cygwin may create four different link types, see https://cygwin.com/cygwin-ug-net/using.html#pathnames-symlinks,
// - MSYS's ln copies target file
// Such behavior makes impossible precalculation of SHA hashes to be used in TestRepository_FileSystem_Symlinks
// because for example Git for Windows (http://git-scm.com) is not aware of symlinks and computes link file's SHA which
// may differ from original file content's SHA.
// As a temporary workaround, we calculating SHA hash by asking git/hg to compute it
func computeCommitHash(repoDir string, git bool) string {
	buf := &bytes.Buffer{}

	if git {
		// git cat-file tree "master^{commit}" | git hash-object -t commit --stdin
		cat := exec.Command("git", "cat-file", "commit", "master^{commit}")
		cat.Dir = repoDir
		hash := exec.Command("git", "hash-object", "-t", "commit", "--stdin")
		hash.Stdin, _ = cat.StdoutPipe()
		hash.Stdout = buf
		hash.Dir = repoDir
		_ = hash.Start()
		_ = cat.Run()
		_ = hash.Wait()
	} else {
		hash := exec.Command("hg", "--debug", "id", "-i")
		hash.Dir = repoDir
		hash.Stdout = buf
		_ = hash.Run()
	}
	return strings.TrimSpace(buf.String())
}
