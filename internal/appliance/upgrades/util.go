package upgrades

// This file contains handler logic for appliances upgrades.

import (
	"fmt"

	"github.com/Masterminds/semver"

	"github.com/sourcegraph/sourcegraph/internal/version"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// Takes a target Version and determines if the upgrade requires downtime or not.
// Returns true if downtime is required and false if not, or on error.
// Return err on invalid target version.
func DetermineUpgradePolicy(currentVersion, targetVersion string) (downtime bool, err error) {
	current := semver.MustParse(currentVersion)
	target := semver.MustParse(targetVersion)

	// Rule out downgrades
	if target.Major() < current.Major() {
		fmt.Println("❌ Downgrade is not supported.")
		return false, errors.New("downgrade is not supported")
	} else if target.Major() == current.Major() && target.Minor() < current.Minor() {
		fmt.Println("❌ Downgrade is not supported.")
		return false, errors.New("downgrade is not supported")
	}

	// If there is a diff between major versions, the policy is MVU
	if target.Major() != current.Major() {
		// Check if the current version is the last minor version in the major release
		lastMinorInMajor, ok := version.LastMinorVersionInMajorRelease[int(current.Major())]
		if ok && int(current.Minor()) == lastMinorInMajor && target.Major() == current.Major()+1 && target.Minor() == 0 {
			fmt.Println("✅ Standard upgrade policy selected.")
			return false, nil
		}
		fmt.Println("✅ MVU upgrade policy selected.")
		return true, nil
	}
	// If there is a diff of greater than one between minor versions, the policy is MVU
	if target.Major() == current.Major() && target.Minor()-current.Minor() > 1 {
		fmt.Println("✅ MVU upgrade policy selected.")
		return true, nil
	}

	fmt.Println("✅ Standard upgrade policy selected.")
	return false, nil
}
