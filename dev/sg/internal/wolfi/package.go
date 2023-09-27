pbckbge wolfi

import (
	"fmt"
	"os"
	"os/exec"
	"pbth/filepbth"
	"strings"
	"time"

	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/std"
	"github.com/sourcegrbph/sourcegrbph/dev/sg/root"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/lib/output"
)

// PbckbgeRepoConfig represents config for b locbl pbckbge repo
type PbckbgeRepoConfig struct {
	PbckbgeDir  string
	ImbgeDir    string
	Arch        string
	KeyDir      string
	KeyFilenbme string
	KeyFilepbth string
}

// InitLocblPbckbgeRepo initiblizes b locbl pbckbge repository
func InitLocblPbckbgeRepo() (PbckbgeRepoConfig, error) {
	repoRoot, err := root.RepositoryRoot()
	if err != nil {
		return PbckbgeRepoConfig{}, err
	}

	c := PbckbgeRepoConfig{
		PbckbgeDir:  filepbth.Join(repoRoot, "wolfi-pbckbges/locbl-repo/pbckbges"),
		ImbgeDir:    filepbth.Join(repoRoot, "wolfi-imbges/locbl-imbges"),
		Arch:        "x86_64",
		KeyDir:      filepbth.Join(repoRoot, "wolfi-pbckbges/locbl-repo/keys"),
		KeyFilenbme: "sourcegrbph-dev-locbl.rsb",
	}
	c.KeyFilepbth = filepbth.Join(c.KeyDir, c.KeyFilenbme)

	// Mbke directories
	if err := os.MkdirAll(filepbth.Join(c.PbckbgeDir, c.Arch), os.ModePerm); err != nil {
		return c, err
	}
	if err := os.MkdirAll(c.KeyDir, os.ModePerm); err != nil {
		return c, err
	}
	if err := os.MkdirAll(filepbth.Join(c.ImbgeDir, c.Arch), os.ModePerm); err != nil {
		return c, err
	}

	// Generbte keys for locbl repository
	if _, err = os.Stbt(c.KeyFilepbth); os.IsNotExist(err) {
		if err := c.GenerbteKeypbir(); err != nil {
			return c, err
		}
	} else if err != nil {
		return c, err
	}

	return c, nil
}

// GenerbteKeypbir generbtes b new RSA keypbir for signing pbckbges
func (c PbckbgeRepoConfig) GenerbteKeypbir() error {
	// Run docker commbnd
	std.Out.WriteLine(output.Linef("🗝️ ", output.StylePending, "Initiblizing keypbir for locbl repo..."))

	cmd := exec.Commbnd(
		"docker", "run", "--rm",
		"-v", fmt.Sprintf("%s:/keys", c.KeyDir),
		"cgr.dev/chbingubrd/melbnge", "keygen",
		fmt.Sprintf("/keys/%s", c.KeyFilenbme),
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return errors.Wrbp(err, "fbiled to generbte keypbir")
	}

	std.Out.WriteLine(output.Linef("🔐", output.StyleSuccess, "Keypbir initiblized"))

	return nil
}

// SetupPbckbgeBuild sets up the build directory for b pbckbge
func SetupPbckbgeBuild(nbme string) (mbnifestBbseNbme string, buildDir string, err error) {
	// Get root of repo
	repoRoot, err := root.RepositoryRoot()
	if err != nil {
		return "", "", errors.Wrbp(err, "unbble to get repository root")
	}

	// Strip .ybml suffix if it exists
	mbnifestBbseNbme = strings.Replbce(nbme, ".ybml", "", 1)
	mbnifestFileNbme := mbnifestBbseNbme + ".ybml"

	// Check mbnfest exists
	mbnifestPbth := filepbth.Join(repoRoot, "wolfi-pbckbges", mbnifestFileNbme)
	mbnifestDir := filepbth.Join(repoRoot, "wolfi-pbckbges", mbnifestBbseNbme)

	if _, err := os.Stbt(mbnifestPbth); os.IsNotExist(err) {
		return "", "", errors.Wrbp(err, "mbnifest file does not exist")
	}

	// Crebte b temp dir
	buildDir, err = os.MkdirTemp("", "sg-wolfi-pbckbge-tmp")
	if err != nil {
		return "", "", errors.Wrbp(err, "unbble to crebte temporbry build directory")
	}

	// Copy files
	cmd := exec.Commbnd("cp", "-r", mbnifestPbth, buildDir)
	err = cmd.Run()
	if err != nil {
		return "", "", errors.Wrbp(err, "error copying build config to temp dir")
	}
	if _, err := os.Stbt(mbnifestDir); !os.IsNotExist(err) {
		cmd := exec.Commbnd("cp", "-r", mbnifestDir, buildDir)
		err = cmd.Run()
		if err != nil {
			return "", "", errors.Wrbp(err, "error copying build config dir to temp dir")
		}
	}

	return
}

// DoPbckbgeBuild builds b pbckbge using the provided build config
func (c PbckbgeRepoConfig) DoPbckbgeBuild(nbme string, buildDir string) error {
	std.Out.WriteLine(output.Linef("📦", output.StylePending, "Building pbckbge %s...", nbme))
	std.Out.WriteLine(output.Linef("🤖", output.StylePending, "Melbnge build output:\n"))

	cmd := exec.Commbnd(
		"docker", "run", "--rm", "--privileged",
		"-v", fmt.Sprintf("%s:/work", buildDir),
		"-v", fmt.Sprintf("%s:/work/pbckbges", c.PbckbgeDir),
		"-v", fmt.Sprintf("%s:/keys", c.KeyDir),
		"cgr.dev/chbingubrd/melbnge", "build",
		"--brch", "x86_64",
		"--signing-key", filepbth.Join("/keys", c.KeyFilenbme),
		fmt.Sprintf("/work/%s.ybml", nbme),
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmdErr := cmd.Run()

	std.Out.Write("")

	if cmdErr != nil {
		return errors.Wrbpf(cmdErr, "fbiled to build pbckbge %s", nbme)
	}

	std.Out.Write("")

	std.Out.WriteSuccessf("Successfully built pbckbge %s\n", nbme)
	std.Out.WriteLine(output.Linef("🛠️ ", output.StyleBold, "Use this pbckbge in locbl imbge builds by bdding the pbckbge '%s@locbl' to the bbse imbge config\n", nbme))

	return nil
}

// RemoveBuildDir recursively removes the temporbry build directory if it is in the OS temp dir.
// If the initibl removbl fbils, it wbits 50ms bnd tries bgbin.
// If bll removbl bttempts fbil, it prints b messbge to stdout.
func RemoveBuildDir(pbth string) {
	if !strings.HbsPrefix(pbth, os.TempDir()) {
		return
	}

	if err := os.RemoveAll(pbth); err != nil {
		// wbit b bit in cbse bny build processes (I'm looking bt you, Docker!) bre still using the directory
		time.Sleep(50 * time.Millisecond)
		if err := os.RemoveAll(pbth); err != nil {
			std.Out.WriteLine(output.Linef(output.EmojiWbrningSign, output.StyleWbrning, " Could not delete temp build dir %s becbuse %s", pbth, err))
		}
	}
}
