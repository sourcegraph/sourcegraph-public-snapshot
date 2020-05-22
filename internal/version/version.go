package version

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/Masterminds/semver"
)

const devVersion = "0.0.0+dev" // version string for unreleased development builds

// version is configured at build time via ldflags like this:
// -ldflags "-X github.com/sourcegraph/sourcegraph/internal/version.version=1.2.3+UnixTimestamp"
var version = devVersion

// Version returns the version string configured at build time.
func Version() string {
	return version
}

// IsDev reports whether the version string is an unreleased development build.
func IsDev(version string) bool {
	return version == devVersion
}

// Mock is used by tests to mock the result of Version and IsDev.
func Mock(mockVersion string) {
	version = mockVersion
}

// HowLongOutOfDate returns a time in months since the last Sourcegraph release based on semantic versions  &
// the fact that Sourcegraph releases every month. It works without needing to call sourcegraph.com and works
// in airgap situations
func HowLongOutOfDate(currentVersion string) (int, error) {
	if IsDev(currentVersion) {
		return 0, nil
	}

	sv, err := semver.NewVersion(currentVersion)
	if err != nil {
		return 0, err
	}

	// expecting major.minor.patch+UnixTimestamp
	if len(sv.Metadata()) == 0 {
		return 0, errors.New("no metadata in semver")
	}
	buildUnixTimestamp, err := strconv.ParseInt(sv.Metadata(), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("unable to parse semver metadata, is it a unix timestamp? %w", err)
	}
	buildTime := time.Unix(buildUnixTimestamp, 0)

	now := time.Now()
	if buildTime.After(now) {
		return 0, errors.New("sourcegraph release version occurs in the future")
	}
	daysSinceBuild := now.Sub(buildTime).Hours() / 24

	months := monthsFromDays(daysSinceBuild)
	return months, nil
}

// monthsFromDays returns a maximum of 6
func monthsFromDays(days float64) int {
	const daysInAMonth = 30
	months := math.Floor(days / daysInAMonth)
	return int(months)
}
