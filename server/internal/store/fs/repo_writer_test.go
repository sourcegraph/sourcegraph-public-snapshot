package fs

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"sourcegraph.com/sourcegraph/go-vcs/vcs"
	"sourcegraph.com/sqs/pbtypes"

	"golang.org/x/tools/godoc/vfs"
)

type testEntry struct {
	Filename string
	Contents []byte
}

type testCommit struct {
	Author    vcs.Signature
	Committer *vcs.Signature
	Message   string
}

var (
	wantCommit = testCommit{
		Author: vcs.Signature{
			Name:  "John",
			Email: "john@john.com",
		},
		Committer: &vcs.Signature{
			Name:  "Jane",
			Email: "jane@jane.co",
		},
		Message: "commit message",
	}
	wantCommit2 = testCommit{
		Author: vcs.Signature{
			Name:  "Johnny",
			Email: "john98@john.com",
		},
		Committer: &vcs.Signature{
			Name:  "Janey",
			Email: "jane21@jane.co",
		},
		Message: "other commit message",
	}
	singleEntry   = testEntry{"single", []byte(`contents`)}
	multipleEntry = []testEntry{
		{"file1", []byte(`Hello world!`)},
		{"file2", []byte(`Hello!`)},
		{"file3", []byte(`Goodbye world!`)},
	}
)

func setup(t *testing.T) string {
	tmpDir, err := ioutil.TempDir("", "repo-stage-test")
	if err != nil {
		t.Fatal(err)
	}

	cmd := exec.Command("git", "init")
	cmd.Dir = tmpDir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("could not initialize test repository (%s): %s", err, out)
	}
	return tmpDir
}

func teardown(t *testing.T, dir string) {
	if err := os.RemoveAll(dir); err != nil {
		t.Fatal(err)
	}
}

// gitIndexFiles returns a list of filenames in the git index.
func gitIndexFiles(dir string) ([]string, error) {
	cmd := exec.Command("git", "status", "--porcelain", "-z")
	cmd.Dir = dir
	cmd.Stderr = os.Stderr
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("exec %v failed: %s (stderr might have more)", cmd.Args, err)
	}
	if len(out) == 0 {
		return nil, nil
	}

	var files []string
	for _, line := range bytes.Split(out, []byte{'\x00'}) {
		workTreeStatus := string(line[1:2])
		if workTreeStatus == " " { // ensure updated in index
			file := string(line[3:])
			files = append(files, file)
		}
	}
	return files, nil
}

// TestRepoStage_New tests that a new repo stage for an inexistent ref is correctly
// created.
func TestRepoStage_New(t *testing.T) {
	repoDir := setup(t)
	defer teardown(t, repoDir)

	rs, err := NewRepoStage(repoDir, "refs/tests/a")
	if err != nil {
		t.Fatal(err)
	}
	defer rs.Free()
	if !strings.Contains(rs.repoDir, "/repo-stage-test") {
		t.Errorf("invalid repo path opened: %s", rs.repoDir)
	}
	files, err := gitIndexFiles(rs.stagingDir)
	if err != nil {
		t.Fatal(err)
	}
	if c := len(files); c != 0 {
		t.Errorf("expected no index entries but got %d in repo", c)
	}
}

// TestRepoStage_Stage tests that staging a file in a new ref correctly updates
// the repo's index and odb.
func TestRepoStage_Add(t *testing.T) {
	repoDir := setup(t)
	defer teardown(t, repoDir)

	rs, err := NewRepoStage(repoDir, "refs/tests/b")
	if err != nil {
		t.Fatal(err)
	}
	defer rs.Free()

	if err := rs.Add(singleEntry.Filename, singleEntry.Contents); err != nil {
		t.Fatal(err)
	}

	files, err := gitIndexFiles(rs.stagingDir)
	if err != nil {
		t.Fatal(err)
	}
	if c := len(files); c != 1 {
		t.Errorf("expected 1 file in index, got %d", c)
	}
	assertIndexAndOdbEntry(t, rs.stagingDir, files[0], singleEntry)
}

// TestRepoStage_Add_Multiple tests that multiple calls to Add will generate
// the correct index and odb entries.
func TestRepoStage_Add_Multiple(t *testing.T) {
	repoDir := setup(t)
	defer teardown(t, repoDir)

	rs, err := NewRepoStage(repoDir, "refs/test/c")
	if err != nil {
		t.Fatal(err)
	}
	defer rs.Free()
	for _, e := range multipleEntry {
		if err := rs.Add(e.Filename, e.Contents); err != nil {
			t.Fatal(err)
		}
	}
	files, err := gitIndexFiles(rs.stagingDir)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != len(multipleEntry) {
		t.Errorf("expected %d files in index, got %d", len(multipleEntry), len(files))
	}
	for _, entry := range multipleEntry {
		assertIndexAndOdbEntry(t, rs.stagingDir, entry.Filename, entry)
	}
}

// TestRepoStage_Commit tests that staged files get committed correctly onto the
// repository.
func TestRepoStage_Commit(t *testing.T) {
	repoDir := setup(t)
	defer teardown(t, repoDir)

	rs, err := NewRepoStage(repoDir, "refs/tests/d")
	if err != nil {
		t.Fatal(err)
	}
	defer rs.Free()

	if err := rs.Add(singleEntry.Filename, singleEntry.Contents); err != nil {
		t.Fatal(err)
	}
	wantCommit.Author.Date = pbtypes.NewTimestamp(time.Now())
	wantCommit.Committer.Date = pbtypes.NewTimestamp(time.Now())
	if err := rs.Commit(wantCommit.Author, *wantCommit.Committer, wantCommit.Message); err != nil {
		t.Fatal(err)
	}
	assertTreeAtTip(t, rs.repoDir, "refs/tests/d", wantCommit, []testEntry{singleEntry}, 0)
}

// TestRepoStage_Commit_Multiple tests that multiple consequential stages and
// commits create the expected commit tree and file structures.
func TestRepoStage_Commit_Mulitple(t *testing.T) {
	repoDir := setup(t)
	defer teardown(t, repoDir)

	rs, err := NewRepoStage(repoDir, "refs/test/e")
	if err != nil {
		t.Fatal(err)
	}
	defer rs.Free()

	// First stage
	for _, e := range multipleEntry {
		if err := rs.Add(e.Filename, e.Contents); err != nil {
			t.Fatal(err)
		}
	}
	// First commit
	wantCommit.Author.Date, wantCommit.Committer.Date = pbtypes.NewTimestamp(time.Now()), pbtypes.NewTimestamp(time.Now())
	if err := rs.Commit(wantCommit.Author, *wantCommit.Committer, wantCommit.Message); err != nil {
		t.Fatal(err)
	}
	commit1 := assertTreeAtTip(t, rs.repoDir, "refs/test/e", wantCommit, multipleEntry, 0)

	// Second stage
	if err := rs.Add(singleEntry.Filename, singleEntry.Contents); err != nil {
		t.Fatal(err)
	}
	// Second commit
	wantCommit2.Author.Date, wantCommit2.Committer.Date = pbtypes.NewTimestamp(time.Now()), pbtypes.NewTimestamp(time.Now())
	if err := rs.Commit(wantCommit2.Author, *wantCommit2.Committer, wantCommit2.Message); err != nil {
		t.Fatal(err)
	}
	commit2 := assertTreeAtTip(t, rs.repoDir, "refs/test/e", wantCommit2, append(multipleEntry, singleEntry), 1)
	if !reflect.DeepEqual(commit2.Parents[0], commit1.ID) {
		t.Error("incorrect commit tree, child not linked to parent")
	}
}

// TestRepoStage_Multiple_Instances tests that multiple instances with
// commits create the expected commit tree and file structures.
// This test is the same as TestRepoStage_Commit_Multiple except it will
// use two separate instances for each commit, asuring that the index
// is restored correctly.
func TestRepoStage_Mulitple_Instances(t *testing.T) {
	repoDir := setup(t)
	defer teardown(t, repoDir)

	rs, err := NewRepoStage(repoDir, "refs/test/e")
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range multipleEntry {
		if err := rs.Add(e.Filename, e.Contents); err != nil {
			t.Fatal(err)
		}
	}
	wantCommit.Author.Date, wantCommit.Committer.Date = pbtypes.NewTimestamp(time.Now()), pbtypes.NewTimestamp(time.Now())
	if err := rs.Commit(wantCommit.Author, *wantCommit.Committer, wantCommit.Message); err != nil {
		t.Fatal(err)
	}
	commit1 := assertTreeAtTip(t, rs.repoDir, "refs/test/e", wantCommit, multipleEntry, 0)
	rs.Free()

	// Reinstantiate
	rs, err = NewRepoStage(repoDir, "refs/test/e")
	if err != nil {
		t.Fatal(err)
	}
	defer rs.Free()
	if err := rs.Add(singleEntry.Filename, singleEntry.Contents); err != nil {
		t.Fatal(err)
	}
	wantCommit2.Author.Date, wantCommit2.Committer.Date = pbtypes.NewTimestamp(time.Now()), pbtypes.NewTimestamp(time.Now())
	if err := rs.Commit(wantCommit2.Author, *wantCommit2.Committer, wantCommit2.Message); err != nil {
		t.Fatal(err)
	}
	commit2 := assertTreeAtTip(t, rs.repoDir, "refs/test/e", wantCommit2, append(multipleEntry, singleEntry), 1)
	if !reflect.DeepEqual(commit2.Parents[0], commit1.ID) {
		t.Error("incorrect commit tree, child not linked to parent")
	}
}

// assertIndexAndOdbEntry verifies that inside the given repository's staging area,
// the given index entry is present.
func assertIndexAndOdbEntry(t *testing.T, stagingRepoDir, filename string, file testEntry) {
	entry, err := os.Stat(filepath.Join(stagingRepoDir, filename))
	if err != nil {
		t.Fatal(err)
	}
	if entry.Size() != int64(len(file.Contents)) {
		t.Errorf("bad file size in index, got %d, wanted %d", entry.Size(), len(file.Contents))
	}
	if entry.Name() != file.Filename {
		t.Errorf("bad file name in index, got '%s', wanted 'file1'", entry.Name())
	}
	if !entry.Mode().IsRegular() {
		t.Errorf("bad filemode in index, wanted blob (regular file), got %v", entry.Mode())
	}
	data, err := ioutil.ReadFile(filepath.Join(stagingRepoDir, file.Filename))
	if err != nil {
		t.Fatal(err)
	}
	if got, want := string(data), string(file.Contents); got != want {
		t.Errorf("file %s: expected contents '%s', got '%s'", file.Filename, want, got)
	}
}

// assertTreeTip verifies that inside the given repo and refName, want is the
// tip commit containing the given files in the tree and having parentCount
// number of parents. It also returns the tip commit.
func assertTreeAtTip(t *testing.T, repoDir string, refName string, want testCommit, files []testEntry, parentCount uint) *vcs.Commit {
	repo, err := vcs.Open("git", repoDir)
	if err != nil {
		t.Fatal(err)
	}

	ref, err := repo.(refResolver).ResolveRef(refName)
	if err != nil {
		t.Fatal(err)
	}
	commit, err := repo.GetCommit(ref)
	if err != nil {
		t.Fatal(err)
	}
	if commit.Message != want.Message {
		t.Errorf("bad commit message, want '%s', got '%s'", want.Message, commit.Message)
	}
	if !signatureEqual(commit.Author, want.Author) {
		t.Errorf("bad commit author, want '%v', got '%v'", want.Author, commit.Author)
	}
	if !signatureEqual(*commit.Committer, *want.Committer) {
		t.Errorf("bad commit committer, want '%v', got '%v'", want.Committer, commit.Committer)
	}
	tree, err := repo.FileSystem(commit.ID)
	if err != nil {
		t.Fatal(err)
	}
	fis, err := tree.ReadDir(".")
	if err != nil {
		t.Fatal(err)
	}
	if count := len(fis); count != len(files) {
		t.Errorf("expected %d file in tree, got %d", len(files), count)
	}
	for _, f := range files {
		entry, err := tree.Stat(f.Filename)
		if err != nil {
			t.Fatal(err)
		}
		if entry.Name() != f.Filename {
			t.Errorf("expected file '%s' in tree, got '%s'", f.Filename, entry.Name)
		}
		if !entry.Mode().IsRegular() {
			t.Errorf("expected regular filemode, got %v", entry.Mode())
		}
		data, err := vfs.ReadFile(tree, f.Filename)
		if err != nil {
			t.Fatal(err)
		}
		if got, want := string(data), string(f.Contents); got != want {
			t.Errorf("file %s: expected contents '%s', got '%s'", f.Filename, want, got)
		}
	}
	if len(commit.Parents) != int(parentCount) {
		t.Errorf("expected %d parents, got %d", parentCount, len(commit.Parents))
	}
	return commit
}

// signatureEqual verifies that two signatures are equal by formatting their time
// signature so that it becomes comparable.
func signatureEqual(a, b vcs.Signature) bool {
	if a.Name != b.Name {
		return false
	}
	if a.Email != b.Email {
		return false
	}
	timeFmt := "Mon Jan 2 15:04:05 2006"
	t1, t2 := a.Date.Time().Format(timeFmt), b.Date.Time().Format(timeFmt)
	if t1 != t2 {
		return false
	}
	return true
}
