package main

import (
	"context"
	"fmt"

	"github.com/Masterminds/semver"

	"github.com/sourcegraph/run"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// standardUpgradeTest initializes Sourcegraph's dbs and runs a standard upgrade
// i.e. an upgrade test between some last minor version and the current release candidate
func standardUpgradeTest(ctx context.Context, initVersion, targetVersion, latestStableVersion *semver.Version) Test {
	//start test env
	test, networkName, dbs, cleanup, err := setupTestEnv(ctx, "standard", initVersion)
	if err != nil {
		test.AddError(errors.Newf("🚨 failed to setup env: %w", err))
		cleanup()
		return test
	}
	defer cleanup()

	// ensure env correctly initialized
	if err := validateDBs(ctx, &test, initVersion.String(), fmt.Sprintf("sourcegraph/migrator:%s", latestStableVersion.String()), networkName, dbs, false); err != nil {
		test.AddError(errors.Newf("🚨 Upgrade failed: %w", err))
		return test
	}

	test.AddLog("-- ⚙️  performing standard upgrade")

	// Run standard upgrade via migrators "up" command
	out, err := run.Cmd(ctx, dockerMigratorBaseString(test, "up", "migrator:candidate", networkName, dbs)...).Run().String()
	if err != nil {
		test.AddError(errors.Newf("🚨 failed to upgrade: %w", err))
		cleanup()
		return test
	}
	test.AddLog(out)

	fctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Start frontend with candidate
	var cleanFrontend func()
	cleanFrontend, err = startFrontend(fctx, test, "frontend", "candidate", networkName, false, dbs)
	if err != nil {
		test.AddError(errors.Newf("🚨 failed to start candidate frontend: %w", err))
		cleanFrontend()
		return test
	}
	defer cleanFrontend()

	test.AddLog("-- ⚙️  post upgrade validation")
	// Validate the upgrade
	if err := validateDBs(ctx, &test, targetVersion.String(), "migrator:candidate", networkName, dbs, true); err != nil {
		test.AddError(errors.Newf("🚨 Upgrade failed: %w", err))
		return test
	}

	return test
}

// multiversionUpgradeTest tests the migrator upgrade command,
// initializing the three main dbs and conducting an upgrade to the release candidate version
func multiversionUpgradeTest(ctx context.Context, initVersion, targetVersion, latestStableVersion *semver.Version) Test {
	test, networkName, dbs, cleanup, err := setupTestEnv(ctx, "multiversion", initVersion)
	if err != nil {
		fmt.Println("🚨 failed to setup env: ", err)
		cleanup()
		return test
	}
	defer cleanup()

	// ensure env correctly initialized, always use latest migrator for drift check,
	// this allows us to avoid issues from changes in migrators invocation
	if err := validateDBs(ctx, &test, initVersion.String(), fmt.Sprintf("sourcegraph/migrator:%s", latestStableVersion.String()), networkName, dbs, false); err != nil {
		test.AddError(errors.Newf("🚨 Initializing env in multiversion test failed: %w", err))
		return test
	}

	// Run multiversion upgrade using candidate image
	//
	// If the build version in the test has been stamped this is the "to" input for the upgrade command. If the builds arent stamped we use the latest stable release verstion,
	// as the target for `migrator upgrade`. This is safe because we assume the target version of the test is always the latest release version. Therefore `migrator up` should always be a single minor version away.
	var toVersion string
	if targetVersion.String() != "0.0.0+dev" { // if version is stamped
		toVersion = targetVersion.String()
	} else {
		toVersion = latestStableVersion.String()
	}
	test.AddLog(fmt.Sprintf("-- ⚙️  performing multiversion upgrade (--from %s --to %s)", initVersion.String(), toVersion))
	out, err := run.Cmd(ctx,
		dockerMigratorBaseString(test, fmt.Sprintf("upgrade --from %s --to %s --ignore-migrator-update", initVersion.String(), toVersion), "migrator:candidate", networkName, dbs)...).
		Run().String()
	if err != nil {
		test.AddError(errors.Newf("🚨 failed to upgrade: %w", err))
		cleanup()
		return test
	}
	test.AddLog(out)

	// Run migrator up with migrator candidate to apply any patch migrations defined on the candidate version
	out, err = run.Cmd(ctx,
		dockerMigratorBaseString(test, "up", "migrator:candidate", networkName, dbs)...).
		Run().String()
	if err != nil {
		test.AddError(errors.Newf("🚨 failed to upgrade: %w", err))
		cleanup()
		return test
	}
	test.AddLog(out)

	// Start frontend with candidate
	var cleanFrontend func()
	cleanFrontend, err = startFrontend(ctx, test, "frontend", "candidate", networkName, false, dbs)
	if err != nil {
		test.AddError(errors.Newf("🚨 failed to start candidate frontend: %w", err))
		cleanFrontend()
		return test
	}
	defer cleanFrontend()

	test.AddLog("-- ⚙️  post upgrade validation")
	// Validate the upgrade
	if err := validateDBs(ctx, &test, targetVersion.String(), "migrator:candidate", networkName, dbs, true); err != nil {
		test.AddError(errors.Newf("🚨 Upgrade failed: %w", err))
		return test
	}

	return test
}

// Logic in tryAutoUpgrade is not compatible with dev builds. The autoUpgrade test must be run with a stamp version
//
// Without this in place autoupgrade fails and exits while trying to make an oobmigration comparison here: https://sourcegraph.com/github.com/sourcegraph/sourcegraph/-/blob/cmd/frontend/internal/cli/autoupgrade.go?L67-76
// {"SeverityText":"WARN","Timestamp":1706721478276103721,"InstrumentationScope":"frontend","Caller":"cli/autoupgrade.go:73","Function":"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/cli.tryAutoUpgrade","Body":"unexpected string for desired instance schema version, skipping auto-upgrade","Resource":{"service.name":"frontend","service.version":"devVersion","service.instance.id":"487754e1c54a"},"Attributes":{"version":"devVersion"}}
func autoUpgradeTest(ctx context.Context, initVersion, targetVersion, latestStableVersion *semver.Version) Test {
	//start test env
	test, networkName, dbs, cleanup, err := setupTestEnv(ctx, "auto", initVersion)
	if err != nil {
		test.AddError(errors.Newf("failed to setup env: %w", err))
		cleanup()
		return test
	}
	defer cleanup()

	// ensure env correctly initialized, always use latest migrator for drift check,
	// this allows us to avoid issues from changes in migrators invocation
	if err := validateDBs(ctx, &test, initVersion.String(), fmt.Sprintf("sourcegraph/migrator:%s", latestStableVersion.String()), networkName, dbs, false); err != nil {
		test.AddError(errors.Newf("🚨 Initializing env in autoupgrade test failed: %w", err))
		return test
	}

	// Set SRC_AUTOUPGRADE=true on Migrator and Frontend containers. Then start the frontend container.
	test.AddLog("-- ⚙️  performing auto upgrade")

	fctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Start frontend with candidate
	var cleanFrontend func()
	cleanFrontend, err = startFrontend(fctx, test, "frontend", "candidate", networkName, true, dbs)
	if err != nil {
		test.AddError(errors.Newf("🚨 failed to start candidate frontend: %w", err))
		cleanFrontend()
		return test
	}
	defer cleanFrontend()

	test.AddLog("-- ⚙️  post upgrade validation")
	// Validate the upgrade
	if err := validateDBs(ctx, &test, targetVersion.String(), "migrator:candidate", networkName, dbs, true); err != nil {
		test.AddError(errors.Newf("🚨 Upgrade failed: %w", err))
		return test
	}

	return test
}
