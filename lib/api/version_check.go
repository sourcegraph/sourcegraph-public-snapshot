package api

import (
	"github.com/grafana/regexp"

	"github.com/Masterminds/semver"
)

// NOTE: A version with a prerelease suffix (e.g. the "-rc.3" of "3.35.1-rc.3") is not
// considered by semver to satisfy a constraint without a prerelease suffix, regardless of
// whether or not the major/minor/patch version is greater than or equal to that of the
// constraint.
//
// For example, the version "3.35.1-rc.3" is not considered to satisfy the constraint ">=
// 3.23.0". This is likely not the expected outcome. However, the same version IS
// considered to satisfy the constraint "3.23.0-0". Thus, it is recommended to pass a
// constraint with a minimum prerelease version suffix attached if comparisons to
// prerelease versions are ever expected. See
// https://github.com/Masterminds/semver#working-with-prerelease-versions for more.
func CheckSourcegraphVersion(version, constraint, minDate string) (bool, error) {
	if version == "dev" || version == "0.0.0+dev" {
		return true, nil
	}

	// Since we don't actually care about the abbreviated commit hash at the end of the
	// version string, we match on 7 or more characters. Currently, the Sourcegraph version
	// is expected to return 12:
	// https://sourcegraph.com/github.com/sourcegraph/sourcegraph/-/blob/enterprise/dev/ci/internal/ci/config.go?L96.
	buildDate := regexp.MustCompile(`^\d+_(\d{4}-\d{2}-\d{2})_[a-z0-9]{7,}$`)
	matches := buildDate.FindStringSubmatch(version)
	if len(matches) > 1 {
		return matches[1] >= minDate, nil
	}

	c, err := semver.NewConstraint(constraint)
	if err != nil {
		return false, nil
	}

	v, err := semver.NewVersion(version)
	if err != nil {
		return false, err
	}

	return c.Check(v), nil
}
