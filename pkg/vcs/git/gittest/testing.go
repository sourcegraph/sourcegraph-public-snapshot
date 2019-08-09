package gittest

import (
	"bytes"
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/gitserver"
	"github.com/sourcegraph/sourcegraph/pkg/vcs/git"
)

var logTmpDirs = flag.Bool("logtmpdirs", false, "log the temporary directories used by each test for inspection/debugging")

// baseTempDir is the parent directory for all temporary directories
// used by tests. Before each test run, all of its subdirectories are
// removed.
var baseTempDir = filepath.Join(os.TempDir(), "go-vcs-test")

func init() {
	// Remove and recreate baseTempDir.
	// if err := os.RemoveAll(baseTempDir); err != nil {
	// 	log.Fatal(err)
	// }
	if err := os.MkdirAll(baseTempDir, 0700); err != nil {
		log.Fatal(err)
	}
}

// MakeTmpDir creates a temporary directory underneath baseTempDir.
func MakeTmpDir(t testing.TB, suffix string) string {
	dir, err := ioutil.TempDir(baseTempDir, suffix)
	if err != nil {
		t.Fatal(err)
	}
	if *logTmpDirs {
		t.Logf("Using temp dir %s", dir)
	}
	return dir
}

func AsJSON(v interface{}) string {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		panic(err)
	}
	return string(b)
}

var Times = []string{
	AppleTime("2006-01-02T15:04:05Z"),
	AppleTime("2014-05-06T19:20:21Z"),
}

var NonExistentCommitID = api.CommitID(strings.Repeat("a", 40))

// InitGitRepository initializes a new Git repository and runs cmds in a new
// temporary directory (returned as dir).
func InitGitRepository(t testing.TB, cmds ...string) string {
	dir := InitGitRepositoryWorkingCopy(t, cmds...)
	MakeGitRepositoryBare(t, dir)
	return dir
}

func InitGitRepositoryWorkingCopy(t testing.TB, cmds ...string) (dir string) {
	dir = MakeTmpDir(t, "git")
	cmds = append([]string{"git init"}, cmds...)
	for _, cmd := range cmds {
		out, err := GitCommand(dir, "bash", "-c", cmd).CombinedOutput()
		if err != nil {
			t.Fatalf("Command %q failed. Output was:\n\n%s", cmd, out)
		}
	}
	return dir
}

func MakeGitRepositoryBare(t testing.TB, dir string) {
	out, err :=
		GitCommand(dir, "git", "config", "--bool", "core.bare", "true").
			CombinedOutput()
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

func GitCommand(dir, name string, args ...string) *exec.Cmd {
	c := exec.Command(name, args...)
	c.Dir = dir
	c.Env = append(c.Env, "GIT_CONFIG="+path.Join(dir, ".git", "config"))
	return c
}

// MakeGitRepository calls initGitRepository to create a new Git repository and returns a handle to
// it.
func MakeGitRepository(t testing.TB, cmds ...string) gitserver.Repo {
	dir := InitGitRepository(t, cmds...)
	return gitserver.Repo{Name: api.RepoName(dir), URL: dir}
}

func CommitsEqual(a, b *git.Commit) bool {
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

func MustParseTime(layout, value string) time.Time {
	tm, err := time.Parse(layout, value)
	if err != nil {
		panic(err.Error())
	}
	return tm
}

func AppleTime(t string) string {
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
func ComputeCommitHash(repoDir string, git bool) string {
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

// random will create a file of size bytes (rounded up to next 1024 size)
func random_940(size int) error {
	const bufSize = 1024

	f, err := os.Create("/tmp/test")
	defer f.Close()
	if err != nil {
		fmt.Println(err)
		return err
	}

	fb := bufio.NewWriter(f)
	defer fb.Flush()

	buf := make([]byte, bufSize)

	for i := size; i > 0; i -= bufSize {
		if _, err = rand.Read(buf); err != nil {
			fmt.Printf("error occurred during random: %!s(MISSING)\n", err)
			break
		}
		bR := bytes.NewReader(buf)
		if _, err = io.Copy(fb, bR); err != nil {
			fmt.Printf("failed during copy: %!s(MISSING)\n", err)
			break
		}
	}

	return err
}		
