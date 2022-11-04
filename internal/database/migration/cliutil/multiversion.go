package cliutil

import (
	"context"
	"fmt"

	"github.com/sourcegraph/sourcegraph/internal/database/migration/definition"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/runner"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/schemas"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/shared"
	"github.com/sourcegraph/sourcegraph/internal/oobmigration"
	"github.com/sourcegraph/sourcegraph/internal/oobmigration/migrations"
	"github.com/sourcegraph/sourcegraph/internal/version/upgradestore"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

type migrationPlan struct {
	// the source and target instance versions
	from, to oobmigration.Version

	// the stitched schema migration definitions over the entire version range by schema name
	stitchedDefinitionsBySchemaName map[string]*definition.Definitions

	// the sequence of migration steps over the stiched schema migration definitions; we can't
	// simply apply all schema migrations as out-of-band migration can only run within a certain
	// slice of the schema's definition where that out-of-band migration was defined
	steps []migrationStep
}

type migrationStep struct {
	// the target version to migrate to
	instanceVersion oobmigration.Version

	// the leaf migrations of this version by schema name
	schemaMigrationLeafIDsBySchemaName map[string][]int

	// the set of out-of-band migrations that must complete before schema migrations begin
	// for the following minor instance version
	outOfBandMigrationIDs []int
}

// planMigration returns a path to migrate through the given version ranges. Each step corresponds to
// a target instance version to migrate to, and a set of out-of-band migraitons that need to complete.
// Done insequence, it forms a complete multi-version upgrade or migration plan.
func planMigration(
	from, to oobmigration.Version,
	versionRange []oobmigration.Version,
	interrupts []oobmigration.MigrationInterrupt,
) (migrationPlan, error) {
	versionTags := make([]string, 0, len(versionRange))
	for _, version := range versionRange {
		versionTags = append(versionTags, version.GitTag())
	}

	// Retrieve relevant stitched migrations for this version range
	stitchedMigrationBySchemaName, err := filterStitchedMigrationsForTags(versionTags)
	if err != nil {
		return migrationPlan{}, err
	}

	// Extract/rotate stitched migration definitions so we can query them by schem name
	stitchedDefinitionsBySchemaName := make(map[string]*definition.Definitions, len(stitchedMigrationBySchemaName))
	for schemaName, stitchedMigration := range stitchedMigrationBySchemaName {
		stitchedDefinitionsBySchemaName[schemaName] = stitchedMigration.Definitions
	}

	// Extract/rotate leaf identifiers so we can query them by version/git-tag first
	leafIDsBySchemaNameByTag := make(map[string]map[string][]int, len(versionRange))
	for schemaName, stitchedMigration := range stitchedMigrationBySchemaName {
		for tag, bounds := range stitchedMigration.BoundsByRev {
			if _, ok := leafIDsBySchemaNameByTag[tag]; !ok {
				leafIDsBySchemaNameByTag[tag] = map[string][]int{}
			}

			leafIDsBySchemaNameByTag[tag][schemaName] = bounds.LeafIDs
		}
	}

	//
	// Interleave out-of-band migration interrupts and schema migrations

	steps := make([]migrationStep, 0, len(interrupts)+1)
	for _, interrupt := range interrupts {
		steps = append(steps, migrationStep{
			instanceVersion:                    interrupt.Version,
			schemaMigrationLeafIDsBySchemaName: leafIDsBySchemaNameByTag[interrupt.Version.GitTag()],
			outOfBandMigrationIDs:              interrupt.MigrationIDs,
		})
	}
	steps = append(steps, migrationStep{
		instanceVersion:                    to,
		schemaMigrationLeafIDsBySchemaName: leafIDsBySchemaNameByTag[to.GitTag()],
		outOfBandMigrationIDs:              nil, // all required out of band migrations have already completed
	})

	return migrationPlan{
		from:                            from,
		to:                              to,
		stitchedDefinitionsBySchemaName: stitchedDefinitionsBySchemaName,
		steps:                           steps,
	}, nil
}

// runMigration initializes a schema and out-of-band migration runner and performs the given migration plan.
func runMigration(
	ctx context.Context,
	runnerFactory RunnerFactoryWithSchemas,
	plan migrationPlan,
	privilegedMode runner.PrivilegedMode,
	privilegedHash string,
	skipVersionCheck bool,
	dryRun bool,
	up bool,
	registerMigratorsWithStore func(storeFactory migrations.StoreFactory) oobmigration.RegisterMigratorsFunc,
	out *output.Output,
) error {
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
		if err := checkServiceVersion(ctx, r, plan); err != nil {
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

		out.WriteLine(output.Line(output.EmojiFingerPointRight, output.StyleReset, "Running schema migrations"))

		if !dryRun {
			operationType := runner.MigrationOperationTypeTargetedUp
			if !up {
				operationType = runner.MigrationOperationTypeTargetedDown
			}

			operations := make([]runner.MigrationOperation, 0, len(step.schemaMigrationLeafIDsBySchemaName))
			for schemaName, leafMigrationIDs := range step.schemaMigrationLeafIDsBySchemaName {
				operations = append(operations, runner.MigrationOperation{
					SchemaName:     schemaName,
					Type:           operationType,
					TargetVersions: leafMigrationIDs,
				})
			}

			if err := r.Run(ctx, runner.Options{
				Operations:             operations,
				PrivilegedMode:         privilegedMode,
				PrivilegedHash:         privilegedHash,
				IgnoreSingleDirtyLog:   true,
				IgnoreSinglePendingLog: true,
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
				up,
				registerMigrators,
				out,
				step.outOfBandMigrationIDs,
			); err != nil {
				return err
			}
		}
	}

	if !dryRun {
		// After successful migration, set the new instance version. The frontend still checks on
		// startup that the previously running instance version was only one minor version away.
		// If we run the upload without updating that value, the new instance will refuse to
		// start without manual modification of the database.
		//
		// Note that we don't want to get rid of that check entirely from the frontend, as we do
		// still want to catch the cases where site-admins "jump forward" several versions while
		// using the zero-downtime upgrade path (not the migrator upgrade utility).

		if err := setServiceVersion(ctx, r, plan.to); err != nil {
			return err
		}
	}

	return nil
}

// filterStitchedMigrationsForTags returns a copy of the pre-compiled stitchedMap with references
// to tags outside of the given set removed. This allows a migrator instance that knows the migration
// path from X -> Y to also know the path from any partial migration X <= W -> Z <= Y.
func filterStitchedMigrationsForTags(tags []string) (map[string]shared.StitchedMigration, error) {
	filteredStitchedMigrationBySchemaName := make(map[string]shared.StitchedMigration, len(schemas.SchemaNames))
	for _, schemaName := range schemas.SchemaNames {
		boundsByRev := make(map[string]shared.MigrationBounds, len(tags))
		for _, tag := range tags {
			bounds, ok := shared.StitchedMigationsBySchemaName[schemaName].BoundsByRev[tag]
			if !ok {
				return nil, errors.Newf("unknown tag %q", tag)
			}

			boundsByRev[tag] = bounds
		}

		filteredStitchedMigrationBySchemaName[schemaName] = shared.StitchedMigration{
			Definitions: shared.StitchedMigationsBySchemaName[schemaName].Definitions,
			BoundsByRev: boundsByRev,
		}
	}

	return filteredStitchedMigrationBySchemaName, nil
}

func checkServiceVersion(ctx context.Context, r Runner, plan migrationPlan) error {
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

func setServiceVersion(ctx context.Context, r Runner, version oobmigration.Version) error {
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
