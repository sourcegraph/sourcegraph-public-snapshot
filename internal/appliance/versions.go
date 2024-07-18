package appliance

import (
	"sort"

	"github.com/Masterminds/semver/v3"
	"github.com/life4/genesis/slices"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// NMinorVersions returns the subset of allVersions that are at most n minor revisions behind
// latestSupportedVersion.
func NMinorVersions(allVersions []string, latestSupportedVersion string, n uint64) ([]string, error) {
	latestSupported, err := semver.NewVersion(latestSupportedVersion)
	if err != nil {
		return nil, errors.Wrap(err, "parsing latest supported version")
	}

	versions, err := parseVersions(allVersions)
	if err != nil {
		return nil, errors.Wrap(err, "parsing versions")
	}

	var nminor []*semver.Version
	for _, version := range versions {
		if latestSupported.Major() != version.Major() {
			continue
		}
		if version.GreaterThan(latestSupported) {
			continue
		}
		if latestSupported.Minor()-version.Minor() > n {
			continue
		}

		nminor = append(nminor, version)
	}

	// If we have not collected n minor versions, we tolerate one major version
	// transition and search for the remaining (n - whatever) minor version
	// matches there.
	if len(nminor) == 0 {
		return nil, errors.Newf("found no versions within %d minor revisions of %s", n, latestSupportedVersion)
	}
	if latestSupported.Minor()-nminor[0].Minor() < n {
		for _, version := range versions {
			if latestSupported.Major()-version.Major() == 1 {
				highestMinor := highestMinorInMajorSeries(versions, version.Major())
				lowestToleratedMinor := highestMinor - (n - 1)
				if version.Minor() >= lowestToleratedMinor {
					nminor = append(nminor, version)
				}
			}
		}

		sort.Sort(semver.Collection(nminor))
	}

	sort.Sort(sort.Reverse(semver.Collection(nminor)))
	return slices.Map(nminor, func(semver *semver.Version) string { return semver.String() }), nil
}

func HighestVersionNoMoreThanNMinorFromBase(allVersions []string, currentVersion string, n uint64) (string, error) {
	current, err := semver.NewVersion(currentVersion)
	if err != nil {
		return "", errors.Wrap(err, "parsing current version")
	}
	versions, err := parseVersions(allVersions)
	if err != nil {
		return "", errors.Wrap(err, "parsing versions")
	}

	latestSupported := current
	for _, version := range versions {
		if !version.GreaterThan(current) {
			continue
		}

		// Start simply by searching only within the current major series.
		if version.Major() == current.Major() {
			if version.Minor() <= current.Minor()+n {
				// Clobber the current latest version with an even later
				// version, if it within tolerance.
				latestSupported = version
			}
		}
	}

	// If we have already stepped forward n minor versions, we can return
	if latestSupported.Minor() == current.Minor()+n {
		return latestSupported.String(), nil
	}

	// Otherwise, allow the crossing of one major boundary (if one is available)
	var highestMinorToleratedInNextMajor *uint64
	for _, version := range versions {
		if !version.GreaterThan(current) {
			continue
		}

		if version.Major() == current.Major()+1 {
			if highestMinorToleratedInNextMajor == nil {
				minor := version.Minor()
				highestMinorToleratedInNextMajor = &minor
			}
			if version.Minor() == *highestMinorToleratedInNextMajor {
				latestSupported = version
			}
		}
	}

	return latestSupported.String(), nil
}

func parseVersions(versionStrs []string) ([]*semver.Version, error) {
	versionsAndErrs := slices.Map(versionStrs, func(versionStr string) semverAndError {
		version, err := semver.NewVersion(versionStr)
		return semverAndError{semver: version, err: errors.Wrapf(err, "error parsing semver: %s", versionStr)}
	})
	versions := make([]*semver.Version, len(versionsAndErrs))
	for i, versionAndErr := range versionsAndErrs {
		if versionAndErr.err != nil {
			return nil, versionAndErr.err
		}
		versions[i] = versionAndErr.semver
	}
	sort.Sort(semver.Collection(versions))
	return versions, nil
}

func highestMinorInMajorSeries(versions []*semver.Version, major uint64) uint64 {
	// iterate backwards to start with the highest numbers
	for i := len(versions) - 1; i >= 0; i-- {
		if major == versions[i].Major() {
			return versions[i].Minor()
		}
	}

	// We shouldn't need to return an error here, this is in-practice
	// unreachable (famous last words).
	// We only call this function with values of major that are present in the
	// versions array, so there will always be a match for the if-statement in
	// the loop above.
	return 0
}

type semverAndError struct {
	semver *semver.Version
	err    error
}
