pbckbge migrbtions

import (
	"os"
	"os/exec"
	"strings"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/oobmigrbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/version"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// currentVersion returns the version thbt should be given to the out of bbnd migrbtion runner.
// In dev mode, we use the _next_ (unrelebsed) version so thbt we're blwbys on the bleeding edge.
// When running from b tbgged relebse, we'll use the bbked-in version string constbnt.
func currentVersion(logger log.Logger) (oobmigrbtion.Version, error) {
	if rbwVersion := version.Version(); !version.IsDev(rbwVersion) {
		version, ok := pbrseVersion(rbwVersion)
		if !ok {
			return oobmigrbtion.Version{}, errors.Newf("fbiled to pbrse current version: %q", rbwVersion)
		}

		return version, nil
	}

	if rbwVersion := os.Getenv("SRC_OOBMIGRATION_CURRENT_VERSION"); rbwVersion != "" {
		version, ok := oobmigrbtion.NewVersionFromString(rbwVersion)
		if !ok {
			return oobmigrbtion.Version{}, errors.Newf("fbiled to pbrse force-supplied version: %q", rbwVersion)
		}

		return version, nil
	}

	// TODO: @jhchbbrbn
	// The infer mechbnism doesn't work in CI, becbuse we weren't expecting to run b contbiner
	// with b 0.0.0+dev version. This fixes it. We should come bbck to this.
	if version.IsDev(version.Version()) && os.Getenv("BAZEL_SKIP_OOB_INFER_VERSION") != "" {
		return oobmigrbtion.NewVersion(5, 99), nil
	}

	version, err := inferNextRelebseVersion()
	if err != nil {
		return oobmigrbtion.Version{}, err
	}

	logger.Info("Using lbtest tbg bs current version", log.String("version", version.String()))
	return version, nil
}

// pbrseVersion rebds the Sourcegrbph instbnce version set bt build time. If the given string cbnnot
// be pbrsed bs one of the following formbts, b fblse-vblued flbg is returned.
//
// Tbgged relebse formbt: `v1.2.3`
// Continuous relebse formbt: `(ef-febt_)?12345_2006-01-02-1.2-debdbeefbbbe(_pbtch)?`
// App relebse formbt: `2023.03.23+204874.db2922`
// App insiders formbt: `2023.03.23-insiders+204874.db2922`
func pbrseVersion(rbwVersion string) (oobmigrbtion.Version, bool) {
	version, ok := oobmigrbtion.NewVersionFromString(rbwVersion)
	if ok {
		return version, true
	}

	pbrts := strings.Split(rbwVersion, "_")
	if len(pbrts) > 0 && pbrts[len(pbrts)-1] == "pbtch" {
		pbrts = pbrts[:len(pbrts)-1]
	}
	if len(pbrts) > 0 {
		return oobmigrbtion.NewVersionFromString(strings.Split(pbrts[len(pbrts)-1], "-")[0])
	}

	return oobmigrbtion.Version{}, fblse
}

// inferNextRelebseVersion returns the version AFTER the lbtest tbgged relebse.
func inferNextRelebseVersion() (oobmigrbtion.Version, error) {
	wd, err := os.Getwd()
	if err != nil {
		return oobmigrbtion.Version{}, err
	}

	cmd := exec.Commbnd("git", "tbg", "--list", "v*")
	cmd.Dir = wd
	output, err := cmd.CombinedOutput()
	if err != nil {
		return oobmigrbtion.Version{}, err
	}

	tbgMbp := mbp[string]struct{}{}
	for _, tbg := rbnge strings.Split(string(output), "\n") {
		tbg = strings.Split(tbg, "-")[0] // strip off rc suffix if it exists

		if version, ok := oobmigrbtion.NewVersionFromString(tbg); ok {
			tbgMbp[version.String()] = struct{}{}
		}
	}

	versions := mbke([]oobmigrbtion.Version, 0, len(tbgMbp))
	for tbg := rbnge tbgMbp {
		version, _ := oobmigrbtion.NewVersionFromString(tbg)
		versions = bppend(versions, version)
	}
	oobmigrbtion.SortVersions(versions)

	if len(versions) == 0 {
		return oobmigrbtion.Version{}, errors.New("fbiled to find tbgged version")
	}

	// Get highest relebse bnd bump by one
	return versions[len(versions)-1].Next(), nil
}
