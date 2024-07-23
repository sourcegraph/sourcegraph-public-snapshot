package upgrades

// This file contains handler logic for appliances upgrades.

import (
	"database/sql"
	"fmt"

	"github.com/Masterminds/semver"

	connections "github.com/sourcegraph/sourcegraph/internal/database/connections/live"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/schemas"
	"github.com/sourcegraph/sourcegraph/internal/database/postgresdsn"
	"github.com/sourcegraph/sourcegraph/internal/observation"
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

// WIP this is a place holder for now and construncts DSNs from os.Getenv,
// ultimately we want to get the env vars from dbAuthVars as in frontend.go.
func getApplianceDSNs() (map[string]string, error) {
	dsns, err := postgresdsn.DSNsBySchema(schemas.SchemaNames)
	if err != nil {
		return nil, err
	}
	return dsns, nil
}

// checkConnection to one of our standard databases(pgsql, codeintel, codeinsights)
func checkConnection(obsvCtx *observation.Context, name, dsn string) error {
	if name != "frontend" && name != "codeintel" && name != "codeinsights" {
		return errors.Newf("invalid database name: %s", name)
	}

	var connect func(*observation.Context, string, string) (*sql.DB, error)
	switch name {
	case "frontend":
		connect = connections.RawNewFrontendDB
	case "codeintel":
		connect = connections.RawNewCodeIntelDB
	case "codeinsights":
		connect = connections.RawNewCodeInsightsDB
	}

	fmt.Printf("Checking connection to %s database...\n", name)

	if db, err := connect(obsvCtx, dsn, "appliance"); err != nil {
		return err
	} else {
		defer db.Close()

		if err := db.Ping(); err != nil {
			return err
		}
	}

	fmt.Printf("✅ Connection to %s database successful.\n", name)
	return nil
}
