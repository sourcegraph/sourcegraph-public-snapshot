package diff

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/hexops/autogold/v2"

	"github.com/sourcegraph/sourcegraph/lib/codeintel/lsif/conversion"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/precise"
)

var dumpPath = "./testdata/project1/dump.lsif"
var dumpPermutedPath = "./testdata/project1/dump-permuted.lsif"
var dumpOldPath = "./testdata/project1/dump-old.lsif"
var dumpNewPath = "./testdata/project1/dump-new.lsif"

func TestNoDiffOnPermutedDumps(t *testing.T) {
	cwd, _ := os.Getwd()
	defer func() { os.Chdir(cwd) }()

	tmpdir1, teardown := createTmpRepo(t, dumpPath)
	t.Cleanup(teardown)
	os.Chdir(tmpdir1)

	bundle1, err := conversion.CorrelateLocalGit(
		context.Background(),
		dumpPath,
		filepath.Dir(dumpPath),
	)
	if err != nil {
		t.Fatalf("Unexpected error reading dump path: %v", err)
	}

	tmpdir2, teardown := createTmpRepo(t, dumpPermutedPath)
	t.Cleanup(teardown)
	os.Chdir(tmpdir2)

	bundle2, err := conversion.CorrelateLocalGit(
		context.Background(),
		dumpPermutedPath,
		filepath.Dir(dumpPermutedPath),
	)
	if err != nil {
		t.Fatalf("Unexpected error reading dump path: %v", err)
	}

	if diff := Diff(precise.GroupedBundleDataChansToMaps(bundle1), precise.GroupedBundleDataChansToMaps(bundle2)); diff != "" {
		t.Fatalf("Dumps %v and %v are not semantically equal:\n%v", dumpPath, dumpPermutedPath, diff)
	}
}

func TestDiffOnEditedDumps(t *testing.T) {
	cwd, _ := os.Getwd()
	defer func() { os.Chdir(cwd) }()

	tmpdir1, teardown := createTmpRepo(t, dumpOldPath)
	t.Cleanup(teardown)
	os.Chdir(tmpdir1)

	bundle1, err := conversion.CorrelateLocalGit(
		context.Background(),
		dumpOldPath,
		filepath.Dir(dumpOldPath),
	)
	if err != nil {
		t.Fatalf("Unexpected error reading dump: %v", err)
	}

	tmpdir2, teardown := createTmpRepo(t, dumpNewPath)
	t.Cleanup(teardown)
	os.Chdir(tmpdir2)

	bundle2, err := conversion.CorrelateLocalGit(
		context.Background(),
		dumpNewPath,
		filepath.Dir(dumpNewPath),
	)
	if err != nil {
		t.Fatalf("Unexpected error reading dump: %v", err)
	}

	computedDiff := Diff(
		precise.GroupedBundleDataChansToMaps(bundle1),
		precise.GroupedBundleDataChansToMaps(bundle2),
	)

	autogold.ExpectFile(t, autogold.Raw(computedDiff))
}

// createTmpRepo returns a temp directory with the testdata copied over from the
// enclosing folder of path, with a newly initialized git repository.
func createTmpRepo(t *testing.T, path string) (string, func()) {
	tmpdir, err := os.MkdirTemp("", "")
	if err != nil {
		t.Fatalf("Unexpected error creating dump tmp folder: %v", err)
	}

	fullpath := filepath.Join(tmpdir, "testdata")
	os.MkdirAll(fullpath, os.ModePerm)

	if err := exec.Command("cp", "-R", filepath.Dir(path), fullpath).Run(); err != nil {
		t.Fatalf("Unexpected error copying dump tmp folder: %v", err)
	}

	gitcmd := exec.Command("git", "init")
	gitcmd.Dir = tmpdir
	if err := gitcmd.Run(); err != nil {
		t.Fatalf("Unexpected error git: %v", err)
	}

	// We need at least a base identity, otherwise git will fail in CI when sandboxed.
	gitcmd = exec.Command("git", "config", "user.email", "test@sourcegraph.com")
	gitcmd.Dir = tmpdir
	if err := gitcmd.Run(); err != nil {
		t.Fatalf("Unexpected error git: %v", err)
	}

	gitcmd = exec.Command("git", "add", ".")
	gitcmd.Dir = tmpdir
	if err := gitcmd.Run(); err != nil {
		t.Fatalf("Unexpected error git: %v", err)
	}

	gitcmd = exec.Command("git", "commit", "-m", "initial commit")
	gitcmd.Dir = tmpdir
	if err := gitcmd.Run(); err != nil {
		t.Fatalf("Unexpected error git: %v", err)
	}

	return tmpdir, func() { os.RemoveAll(tmpdir) }
}
