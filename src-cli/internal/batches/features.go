package batches

import (
	"github.com/sourcegraph/sourcegraph/lib/api"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// FeatureFlags represent features that are only available on certain
// Sourcegraph versions and we therefore have to detect at runtime.
type FeatureFlags struct {
	Sourcegraph40 bool
	BinaryDiffs   bool
}

func (ff *FeatureFlags) SetFromVersion(version string) error {
	for _, feature := range []struct {
		flag       *bool
		constraint string
		minDate    string
	}{
		// NOTE: It's necessary to include a "-0" prerelease suffix on each constraint so that
		// prereleases of future versions are still considered to satisfy the constraint.
		//
		// For example, the version "3.35.1-rc.3" is not considered to satisfy the constraint
		// ">= 3.23.0". However, the same version IS considered to satisfy the constraint
		// "3.23.0-0". See
		// https://github.com/Masterminds/semver#working-with-prerelease-versions for more.
		// Example usage:
		// {&ff.FlagName, ">= 3.23.0-0", "2020-11-24"},
		{&ff.Sourcegraph40, ">= 4.0.0-0", "2022-08-24"},
		{&ff.BinaryDiffs, ">= 4.3.0-0", "2022-11-29"},
	} {
		value, err := api.CheckSourcegraphVersion(version, feature.constraint, feature.minDate)
		if err != nil {
			return errors.Wrap(err, "failed to check version returned by Sourcegraph")
		}
		*feature.flag = value
	}

	return nil
}
