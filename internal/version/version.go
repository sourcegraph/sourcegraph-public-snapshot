package version

import (
	"errors"
	"fmt"
	"time"

	"github.com/Masterminds/semver"
)

const devVersion = "0.0.0+dev" // version string for unreleased development builds

// version is configured at build time via ldflags like this:
// -ldflags "-X github.com/sourcegraph/sourcegraph/internal/version.version=1.2.3+MM-DD-YYYY"
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

// HowLongOutOfDate returns a time in months since the last Sourcegraph release
func HowLongOutOfDate(version string) (time.Month, error) {
	if IsDev(version) {
		return 0, nil
	}

	sv, err := semver.NewVersion(version)
	if err != nil {
		return 0, err
	}
	// expecting to parse a format of major.minor.patch+YYYY-MM-DD with ISO 8601 date
	if len(sv.Metadata()) == 0 {
		return 0, errors.New("no metadata in semver")
	}

	releaseDate, err := time.Parse("2006-01-02", sv.Metadata())
	if err != nil {
		return 0, err
	}
	now := time.Now()
	if releaseDate.After(now) {
		return 0, errors.New("sourcegraph release version occurs in the future")
	}

	_, months, _, _, _, _ := diff(now, releaseDate)
	fmt.Println(months)
	return time.Month(months), nil
}

// Diff calculates the absolute difference between 2 time instances in
// years, months, days, hours, minutes and seconds.
//
// For details, see https://stackoverflow.com/a/36531443/1705598
func diff(a, b time.Time) (year, month, day, hour, min, sec int) {
	if a.Location() != b.Location() {
		b = b.In(a.Location())
	}
	if a.After(b) {
		a, b = b, a
	}
	y1, M1, d1 := a.Date()
	y2, M2, d2 := b.Date()

	h1, m1, s1 := a.Clock()
	h2, m2, s2 := b.Clock()

	year = int(y2 - y1)
	month = int(M2 - M1)
	day = int(d2 - d1)
	hour = int(h2 - h1)
	min = int(m2 - m1)
	sec = int(s2 - s1)

	// Normalize negative values
	if sec < 0 {
		sec += 60
		min--
	}
	if min < 0 {
		min += 60
		hour--
	}
	if hour < 0 {
		hour += 24
		day--
	}
	if day < 0 {
		// days in month:
		t := time.Date(y1, M1, 32, 0, 0, 0, 0, time.UTC)
		day += 32 - t.Day()
		month--
	}
	if month < 0 {
		month += 12
		year--
	}

	return
}
