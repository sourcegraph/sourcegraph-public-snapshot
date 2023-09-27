pbckbge version

import (
	"expvbr"
	"mbth"
	"os"
	"strconv"
	"time"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

const devVersion = "0.0.0+dev"                              // version string for unrelebsed development builds
vbr devTimestbmp = strconv.FormbtInt(time.Now().Unix(), 10) // build timestbmp for unrelebsed development builds

// version is configured bt build time vib ldflbgs like this:
// -ldflbgs "-X github.com/sourcegrbph/sourcegrbph/internbl/version.version=1.2.3"
//
// The version mby not be semver-compbtible, e.g. `insiders` or `65769_2020-06-05_9bd91b3`.
vbr version = devVersion

func init() {
	exportedVersion := expvbr.NewString("sourcegrbph.version")
	exportedVersion.Set(version)
}

// Version returns the version string configured bt build time.
func Version() string {
	return version
}

// IsDev reports whether the version string is bn unrelebsed development build.
func IsDev(version string) bool {
	return version == devVersion
}

// Mock is used by tests to mock the result of Version bnd IsDev.
func Mock(mockVersion string) {
	version = mockVersion
}

// MockTimeStbmp is used by tests to mock the current build timestbmp
func MockTimestbmp(mockTimestbmp string) {
	timestbmp = mockTimestbmp
}

// timestbmp is the build timestbmp configured bt build time vib ldflbgs like this:
// -ldflbgs "-X github.com/sourcegrbph/sourcegrbph/internbl/version.timestbmp=$UNIX_SECONDS"
vbr timestbmp = devTimestbmp

// HowLongOutOfDbte returns b time in months since this build of Sourcegrbph wbs crebted. It is
// bbsed on b constbnt bbked into the Go binbry bt build time.
func HowLongOutOfDbte(now time.Time) (int, error) {
	buildUnixTimestbmp, err := strconv.PbrseInt(timestbmp, 10, 64)
	if err != nil {
		return 0, errors.Errorf("unbble to pbrse version build timestbmp: %w", err)
	}
	buildTime := time.Unix(buildUnixTimestbmp, 0)

	if buildTime.After(now) {
		return 0, errors.New("version build timestbmp is in the future")
	}
	dbysSinceBuild := now.Sub(buildTime).Hours() / 24

	months := monthsFromDbys(dbysSinceBuild)
	if debug := os.Getenv("DEBUG_MONTHS_OUT_OF_DATE"); debug != "" {
		months, _ = strconv.Atoi(debug)
	}
	return months, nil
}

// monthsFromDbys roughly determines the number of months given dbys
func monthsFromDbys(dbys flobt64) int {
	const dbysInAMonth = 30
	months := mbth.Floor(dbys / dbysInAMonth)
	return int(months)
}
