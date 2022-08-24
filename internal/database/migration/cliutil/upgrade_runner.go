package cliutil

import (
	"context"
	"fmt"

	"github.com/sourcegraph/sourcegraph/internal/database/migration/runner"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/schemas"
	"github.com/sourcegraph/sourcegraph/internal/oobmigration"
	"github.com/sourcegraph/sourcegraph/internal/oobmigration/migrations"
	"github.com/sourcegraph/sourcegraph/internal/version/upgradestore"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

// runUpgrade initializes a schema and out-of-band migration runner and performs the given upgrade plan.
func runUpgrade(
	ctx context.Context,
	runnerFactory RunnerFactoryWithSchemas,
	plan upgradePlan,
	skipVersionCheck bool,
	dryRun bool,
	registerMigratorsWithStore func(storeFactory migrations.StoreFactory) oobmigration.RegisterMigratorsFunc,
	out *output.Output,
) error {
	if len(plan.steps) == 0 {
		return errors.New("upgrade plan contains no steps")
	}

	var runnerSchemas []*schemas.Schema
	for _, schemaName := range schemas.SchemaNames {
		runnerSchemas = append(runnerSchemas, &schemas.Schema{
			Name:                schemaName,
			MigrationsTableName: schemas.MigrationsTableName(schemaName),
			Definitions:         plan.stitchedDefinitionsBySchemaName[schemaName],
		})
	}

	r, err := runnerFactory(ctx, schemas.SchemaNames, runnerSchemas)
	if err != nil {
		return err
	}
	db, err := extractDatabase(ctx, r)
	if err != nil {
		return err
	}
	registerMigrators := registerMigratorsWithStore(basestoreExtractor{r})

	if !skipVersionCheck {
		if err := checkUpgradeVersion(ctx, r, plan); err != nil {
			return errors.Newf("%s. Re-invoke with --skip-version-check to ignore this check", err)
		}
	}

	for i, step := range plan.steps {
		out.WriteLine(output.Linef(
			output.EmojiFingerPointRight,
			output.StyleReset,
			"Migrating to v%s (step %d of %d)",
			step.instanceVersion,
			i+1,
			len(plan.steps),
		))

		operations := make([]runner.MigrationOperation, 0, len(step.schemaMigrationLeafIDsBySchemaName))
		for schemaName, leafMigrationIDs := range step.schemaMigrationLeafIDsBySchemaName {
			operations = append(operations, runner.MigrationOperation{
				SchemaName:     schemaName,
				Type:           runner.MigrationOperationTypeTargetedUp,
				TargetVersions: leafMigrationIDs,
			})
		}

		out.WriteLine(output.Line(output.EmojiFingerPointRight, output.StyleReset, "Running schema migrations"))

		if !dryRun {
			if err := r.Run(ctx, runner.Options{
				Operations:           operations,
				PrivilegedMode:       runner.ApplyPrivilegedMigrations,
				PrivilegedHash:       "",
				IgnoreSingleDirtyLog: false,
			}); err != nil {
				return err
			}

			out.WriteLine(output.Line(output.EmojiSuccess, output.StyleSuccess, "Schema migrations complete"))
		}

		if len(step.outOfBandMigrationIDs) > 0 {
			if err := runOutOfBandMigrations(
				ctx,
				db,
				dryRun,
				registerMigrators,
				out,
				step.outOfBandMigrationIDs,
			); err != nil {
				return err
			}
		}
	}

	// After successful upgrade, set the new instance version. The frontend still checks on
	// startup that the previously running instance version was only one minor version away.
	// If we run the upload without updating that value, the new instance will refuse to
	// start without manual modification of the database.
	//
	// Note that we don't want to get rid of that check entirely from the frontend, as we do
	// still want to catch the cases where site-admins "jump forward" several versions while
	// using the zero-downtime upgrade path (not this upgrade utility).

	if err := updateVersion(ctx, r, plan.to); err != nil {
		return err
	}

	return nil
}

func checkUpgradeVersion(ctx context.Context, r Runner, plan upgradePlan) error {
	db, err := extractDatabase(ctx, r)
	if err != nil {
		return err
	}

	versionStr, ok, err := upgradestore.New(db).GetServiceVersion(ctx, "frontend")
	if err != nil {
		return err
	}
	if ok {
		version, ok := oobmigration.NewVersionFromString(versionStr)
		if !ok {
			return errors.Newf("cannot parse version: %q - expected [v]X.Y[.Z]", versionStr)
		}
		if oobmigration.CompareVersions(version, plan.from) == oobmigration.VersionOrderEqual {
			return nil
		}

		return errors.Newf("version assertion failed: %q != %q", version, plan.from)
	}

	return errors.Newf("version assertion failed: unknown version != %q", plan.from)
}

func updateVersion(ctx context.Context, r Runner, version oobmigration.Version) error {
	db, err := extractDatabase(ctx, r)
	if err != nil {
		return err
	}

	return upgradestore.New(db).SetServiceVersion(
		ctx,
		"frontend",
		fmt.Sprintf("%d.%d.0", version.Major, version.Minor),
	)
}
