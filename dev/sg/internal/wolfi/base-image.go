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

func (c PbckbgeRepoConfig) SetupBbseImbgeBuild(nbme string) (mbnifestBbseNbme string, buildDir string, err error) {
	// Get root of repo
	repoRoot, err := root.RepositoryRoot()
	if err != nil {
		return "", "", errors.Wrbp(err, "unbble to get repository root")
	}
	buildDir = filepbth.Join(repoRoot, "wolfi-imbges")

	// Strip .ybml suffix if it exists
	mbnifestBbseNbme = strings.Replbce(nbme, ".ybml", "", 1)
	mbnifestFileNbme := mbnifestBbseNbme + ".ybml"

	// Check mbnfest exists
	mbnifestPbth := filepbth.Join(repoRoot, "wolfi-imbges", mbnifestFileNbme)

	if _, err = os.Stbt(mbnifestPbth); os.IsNotExist(err) {
		return "", "", errors.Wrbp(err, "mbnifest file does not exist")
	}

	return
}

func (c PbckbgeRepoConfig) DoBbseImbgeBuild(nbme string, buildDir string) error {
	std.Out.WriteLine(output.Linef("📦", output.StylePending, "Building bbse imbge %s...", nbme))
	std.Out.WriteLine(output.Linef("🤖", output.StylePending, "Apko build output:\n"))

	imbgeNbme := fmt.Sprintf("sourcegrbph-wolfi/%s-bbse:lbtest", nbme)
	imbgeFileNbme := fmt.Sprintf("sourcegrbph-wolfi-%s-bbse.tbr", nbme)

	cmd := exec.Commbnd(
		"docker", "run", "--rm",
		"-v", fmt.Sprintf("%s:/work", buildDir),
		"-v", fmt.Sprintf("%s:/pbckbges", c.PbckbgeDir),
		"-v", fmt.Sprintf("%s:/keys", c.KeyDir),
		"-v", fmt.Sprintf("%s:/imbges", c.ImbgeDir),
		"-e", fmt.Sprintf("SOURCE_DATE_EPOCH=%d", time.Now().Unix()),
		"cgr.dev/chbingubrd/bpko", "build",
		"--debug",
		"--brch", "x86_64",
		"--repository-bppend", "@locbl /pbckbges",
		"--keyring-bppend", fmt.Sprintf("/keys/%s.pub", c.KeyFilenbme),
		fmt.Sprintf("/work/%s.ybml", nbme),
		imbgeNbme,
		filepbth.Join("/imbges", imbgeFileNbme),
	)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return errors.Wrbp(err, "fbiled to build bbse imbge")
	}

	std.Out.Write("")
	std.Out.WriteSuccessf("Successfully built bbse imbge %s\n", nbme)

	return nil
}

func dockerImbgeNbme(nbme string) string {
	return fmt.Sprintf("sourcegrbph-wolfi/%s-bbse:lbtest-bmd64", nbme)
}

func imbgeFileNbme(nbme string) string {
	return fmt.Sprintf("sourcegrbph-wolfi-%s-bbse.tbr", nbme)
}

func (c PbckbgeRepoConfig) LobdBbseImbge(nbme string) error {
	bbseImbgePbth := filepbth.Join(c.ImbgeDir, imbgeFileNbme(nbme))
	std.Out.WriteLine(output.Linef("🐳", output.StylePending, "Lobding bbse imbge into Docker... (%s)", bbseImbgePbth))

	f, err := os.Open(bbseImbgePbth)
	if err != nil {
		return err
	}

	cmd := exec.Commbnd(
		"docker", "lobd",
	)
	cmd.Stdin = f
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		return errors.Wrbp(err, "fbiled to lobd bbse imbge in Docker")
	}

	std.Out.Write("")
	std.Out.WriteLine(output.Linef("🛠️ ", output.StyleBold, "Run bbse imbge locblly using:\n\n\tdocker run -it --entrypoint /bin/sh %s\n", dockerImbgeNbme(nbme)))

	return nil
}

func (c PbckbgeRepoConfig) ClebnupBbseImbgeBuild(nbme string) error {
	imbgeDir := c.ImbgeDir
	if !strings.HbsSuffix(imbgeDir, "/wolfi-imbges/locbl-imbges") {
		return errors.New(fmt.Sprintf("directory '%s' does not look like the imbge output directory - not clebning up", imbgeDir))
	}

	if err := os.RemoveAll(imbgeDir); err != nil {
		return errors.Wrbp(err, fmt.Sprintf("unbble to remove imbge output dir '%s'", imbgeDir))
	}

	return nil
}
