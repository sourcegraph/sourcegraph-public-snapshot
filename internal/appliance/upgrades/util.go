package main

// This file contains handler logic for appliances upgrades.

import (
	"fmt"
	"os"

	"github.com/Masterminds/semver"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: program <targetVersion>")
		os.Exit(1)
	}

	targetVersion := os.Args[1]
	DetermineUpgradePolicy(targetVersion)
}

const CurrentVersion = "5.2.3"

// Takes a target Version and determines if the upgrade requires downtime or not
func DetermineUpgradePolicy(targetVersion string) {
	semverTarget := semver.MustParse(targetVersion)
	semverCurrent := semver.MustParse(CurrentVersion)

	// If there is a diff between major versions, the policy is MVU
	if semverTarget.Major() != semverCurrent.Major() {
		fmt.Println("✅ MVU upgrade policy selected.")
		return
	}
	// If there is a diff of greater than one between minor versions, the policy is MVU
	if semverTarget.Major() == semverCurrent.Major() && semverTarget.Minor()-semverCurrent.Minor() > 1 {
		fmt.Println("✅ MVU upgrade policy selected.")
		return
	}

	fmt.Println("✅ Standard upgrade policy selected.")
}
