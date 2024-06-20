package main

// This file contains handler logic for appliances upgrades.

import (
	"context"
	"fmt"
	"os"

	"github.com/Masterminds/semver"

	"github.com/sourcegraph/sourcegraph/internal/database"
	connections "github.com/sourcegraph/sourcegraph/internal/database/connections/live"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/version/upgradestore"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// This is just for testing purposes.
func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: program <targetVersion>")
		os.Exit(1)
	}

	targetVersion := os.Args[1]
	DetermineUpgradePolicy(targetVersion)
}

// Get current Version
func GetCurrentVersion(ctx context.Context, obsvCtx *observation.Context) string {
	sqlDB, err := connections.RawNewFrontendDB(obsvCtx, "", "frontend")
	if err != nil {
		return errors.Errorf("failed to connect to frontend database: %s", err)
	}
	defer sqlDB.Close()

	db := database.NewDB(obsvCtx.Logger, sqlDB)
	upgradestore := upgradestore.New(db)

	currentVersion := upgradestore.GetServiceVersion(ctx)
}

// Takes a target Version and determines if the upgrade requires downtime or not
func DetermineUpgradePolicy(targetVersion string) {
	GetCurrentVersion(db)

	semverCurrent := semver.MustParse(currentVersion)
	semverTarget := semver.MustParse(targetVersion)

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
