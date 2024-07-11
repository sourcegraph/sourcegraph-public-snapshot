package appliance_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/appliance"
)

var allVersions = []string{
	"1.2.3",
	"1.2.4",
	"2.0.0",
	"2.2.0",
	"2.5.0",
	"2.7.0",
	"2.7.1",
	"3.0.0",
	"3.0.1",
	"3.0.2",
	"3.1.0",
	"3.1.1",
	"3.1.2",
	"3.2.0",
	"3.3.0",

	// We can't guarantee that the releaseregistry returns in sorted semver
	// order, as opposed to sorted by time.
	"3.2.1",

	"4.0.0",
	"4.0.1",
	"4.1.0",
	"4.1.1",
}

func TestNMinorVersions(t *testing.T) {
	for _, tc := range []struct {
		name                   string
		latestSupportedVersion string
		expectedOutput         []string
	}{
		{
			name:                   "returns all patch versions within 2 minor points of a given release",
			latestSupportedVersion: "3.2.0",
			expectedOutput: []string{
				"3.2.0",
				"3.1.2",
				"3.1.1",
				"3.1.0",
				"3.0.2",
				"3.0.1",
				"3.0.0",
			},
		},
		{
			name:                   "returns all patch versions within <n minor points of a given release when there are not enough releases",
			latestSupportedVersion: "1.2.4",
			expectedOutput: []string{
				"1.2.4",
				"1.2.3",
			},
		},
		{
			name:                   "returns all patch versions within 2 minor revisions, where crossing a single major boundary counts as 1 minor revision",
			latestSupportedVersion: "4.0.1",
			expectedOutput: []string{
				"4.0.1",
				"4.0.0",
				"3.3.0",
				"3.2.1",
				"3.2.0",
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			versions, err := appliance.NMinorVersions(allVersions, tc.latestSupportedVersion, 2)
			require.NoError(t, err)
			require.Equal(t, tc.expectedOutput, versions)
		})
	}
}

func TestHighestVersionNoMoreThanNMinorFromBase(t *testing.T) {
	for _, tc := range []struct {
		name            string
		currentVersion  string
		expectedVersion string
	}{
		{
			name:            "returns highest patch version 2 minor points above specified release",
			currentVersion:  "3.0.2",
			expectedVersion: "3.2.1",
		},
		{
			name:            "returns latest patch version when there are no newer minor versions",
			currentVersion:  "4.1.0",
			expectedVersion: "4.1.1",
		},
		{
			name:            "returns latest version available when the current version is less than 2 minor points behind",
			currentVersion:  "4.0.0",
			expectedVersion: "4.1.1",
		},
		{
			name:            "returns current version when it is the latest version",
			currentVersion:  "4.1.1",
			expectedVersion: "4.1.1",
		},
		{
			name:            "tolerates crossing a major version boundary, but then no further minor versions",
			currentVersion:  "3.2.1",
			expectedVersion: "4.0.1",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			latestSupportedVersion, err := appliance.HighestVersionNoMoreThanNMinorFromBase(allVersions, tc.currentVersion, 2)
			require.NoError(t, err)
			require.Equal(t, tc.expectedVersion, latestSupportedVersion)
		})
	}
}
