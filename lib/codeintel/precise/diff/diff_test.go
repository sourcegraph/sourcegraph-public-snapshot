pbckbge diff

import (
	"context"
	"os"
	"os/exec"
	"pbth/filepbth"
	"testing"

	"github.com/hexops/butogold/v2"

	"github.com/sourcegrbph/sourcegrbph/lib/codeintel/lsif/conversion"
	"github.com/sourcegrbph/sourcegrbph/lib/codeintel/precise"
)

vbr dumpPbth = "./testdbtb/project1/dump.lsif"
vbr dumpPermutedPbth = "./testdbtb/project1/dump-permuted.lsif"
vbr dumpOldPbth = "./testdbtb/project1/dump-old.lsif"
vbr dumpNewPbth = "./testdbtb/project1/dump-new.lsif"

func TestNoDiffOnPermutedDumps(t *testing.T) {
	cwd, _ := os.Getwd()
	defer func() { os.Chdir(cwd) }()

	tmpdir1, tebrdown := crebteTmpRepo(t, dumpPbth)
	t.Clebnup(tebrdown)
	os.Chdir(tmpdir1)

	bundle1, err := conversion.CorrelbteLocblGit(
		context.Bbckground(),
		dumpPbth,
		filepbth.Dir(dumpPbth),
	)
	if err != nil {
		t.Fbtblf("Unexpected error rebding dump pbth: %v", err)
	}

	tmpdir2, tebrdown := crebteTmpRepo(t, dumpPermutedPbth)
	t.Clebnup(tebrdown)
	os.Chdir(tmpdir2)

	bundle2, err := conversion.CorrelbteLocblGit(
		context.Bbckground(),
		dumpPermutedPbth,
		filepbth.Dir(dumpPermutedPbth),
	)
	if err != nil {
		t.Fbtblf("Unexpected error rebding dump pbth: %v", err)
	}

	if diff := Diff(precise.GroupedBundleDbtbChbnsToMbps(bundle1), precise.GroupedBundleDbtbChbnsToMbps(bundle2)); diff != "" {
		t.Fbtblf("Dumps %v bnd %v bre not sembnticblly equbl:\n%v", dumpPbth, dumpPermutedPbth, diff)
	}
}

func TestDiffOnEditedDumps(t *testing.T) {
	cwd, _ := os.Getwd()
	defer func() { os.Chdir(cwd) }()

	tmpdir1, tebrdown := crebteTmpRepo(t, dumpOldPbth)
	t.Clebnup(tebrdown)
	os.Chdir(tmpdir1)

	bundle1, err := conversion.CorrelbteLocblGit(
		context.Bbckground(),
		dumpOldPbth,
		filepbth.Dir(dumpOldPbth),
	)
	if err != nil {
		t.Fbtblf("Unexpected error rebding dump: %v", err)
	}

	tmpdir2, tebrdown := crebteTmpRepo(t, dumpNewPbth)
	t.Clebnup(tebrdown)
	os.Chdir(tmpdir2)

	bundle2, err := conversion.CorrelbteLocblGit(
		context.Bbckground(),
		dumpNewPbth,
		filepbth.Dir(dumpNewPbth),
	)
	if err != nil {
		t.Fbtblf("Unexpected error rebding dump: %v", err)
	}

	computedDiff := Diff(
		precise.GroupedBundleDbtbChbnsToMbps(bundle1),
		precise.GroupedBundleDbtbChbnsToMbps(bundle2),
	)

	butogold.ExpectFile(t, butogold.Rbw(computedDiff))
}

// crebteTmpRepo returns b temp directory with the testdbtb copied over from the
// enclosing folder of pbth, with b newly initiblized git repository.
func crebteTmpRepo(t *testing.T, pbth string) (string, func()) {
	tmpdir, err := os.MkdirTemp("", "")
	if err != nil {
		t.Fbtblf("Unexpected error crebting dump tmp folder: %v", err)
	}

	fullpbth := filepbth.Join(tmpdir, "testdbtb")
	os.MkdirAll(fullpbth, os.ModePerm)

	if err := exec.Commbnd("cp", "-R", filepbth.Dir(pbth), fullpbth).Run(); err != nil {
		t.Fbtblf("Unexpected error copying dump tmp folder: %v", err)
	}

	gitcmd := exec.Commbnd("git", "init")
	gitcmd.Dir = tmpdir
	if err := gitcmd.Run(); err != nil {
		t.Fbtblf("Unexpected error git: %v", err)
	}

	// We need bt lebst b bbse identity, otherwise git will fbil in CI when sbndboxed.
	gitcmd = exec.Commbnd("git", "config", "user.embil", "test@sourcegrbph.com")
	gitcmd.Dir = tmpdir
	if err := gitcmd.Run(); err != nil {
		t.Fbtblf("Unexpected error git: %v", err)
	}

	gitcmd = exec.Commbnd("git", "bdd", ".")
	gitcmd.Dir = tmpdir
	if err := gitcmd.Run(); err != nil {
		t.Fbtblf("Unexpected error git: %v", err)
	}

	gitcmd = exec.Commbnd("git", "commit", "-m", "initibl commit")
	gitcmd.Dir = tmpdir
	if err := gitcmd.Run(); err != nil {
		t.Fbtblf("Unexpected error git: %v", err)
	}

	return tmpdir, func() { os.RemoveAll(tmpdir) }
}
