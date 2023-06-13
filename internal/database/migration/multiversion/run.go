package multiversion

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/output"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/definition"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/runner"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/schemas"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/store"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/oobmigration"
	"github.com/sourcegraph/sourcegraph/internal/oobmigration/migrations"
	"github.com/sourcegraph/sourcegraph/internal/version/upgradestore"
)

type Store interface {
	WithMigrationLog(ctx context.Context, definition definition.Definition, up bool, f func() error) error
	Describe(ctx context.Context) (map[string]schemas.SchemaDescription, error)
	Versions(ctx context.Context) (appliedVersions, pendingVersions, failedVersions []int, _ error)
}

func RunMigration(
	ctx context.Context,
	db database.DB,
	runnerFactory runner.RunnerFactoryWithSchemas,
	plan MigrationPlan,
	privilegedMode runner.PrivilegedMode,
	privilegedHashes []string,
	skipVersionCheck bool,
	skipDriftCheck bool,
	dryRun bool,
	up bool,
	animateProgress bool,
	registerMigratorsWithStore func(storeFactory migrations.StoreFactory) oobmigration.RegisterMigratorsFunc,
	expectedSchemaFactories []schemas.ExpectedSchemaFactory,
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

	r, err := runnerFactory(schemas.SchemaNames, runnerSchemas)
	if err != nil {
		return err
	}

	registerMigrators := registerMigratorsWithStore(store.BasestoreExtractor{Runner: r})

	// Note: Error is correctly checked here; we want to use the return value
	// `patch` below but only if we can best-effort fetch it. We want to allow
	// the user to skip erroring here if they are explicitly skipping this
	// version check.
	version, patch, ok, err := GetServiceVersion(ctx, db)
	if !skipVersionCheck {
		if err != nil {
			return err
		}
		if !ok {
			return errors.Newf("version assertion failed: unknown version != %q. Re-invoke with --skip-version-check to ignore this check", plan.from)
		}
		if oobmigration.CompareVersions(version, plan.from) != oobmigration.VersionOrderEqual {
			return errors.Newf("version assertion failed: %q != %q. Re-invoke with --skip-version-check to ignore this check", version, plan.from)
		}
	}

	if !skipDriftCheck {
		if err := CheckDrift(ctx, r, plan.from.GitTagWithPatch(patch), out, false, schemas.SchemaNames, expectedSchemaFactories); err != nil {
			return err
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
				Operations:     operations,
				PrivilegedMode: privilegedMode,
				MatchPrivilegedHash: func(hash string) bool {
					for _, candidate := range privilegedHashes {
						if hash == candidate {
							return true
						}
					}

					return false
				},
				IgnoreSingleDirtyLog:   true,
				IgnoreSinglePendingLog: true,
			}); err != nil {
				return err
			}

			out.WriteLine(output.Line(output.EmojiSuccess, output.StyleSuccess, "Schema migrations complete"))
		}

		if len(step.outOfBandMigrationIDs) > 0 {
			if err := RunOutOfBandMigrations(
				ctx,
				db,
				dryRun,
				up,
				animateProgress,
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
		// using the standard upgrade path (not a multi-version upgrade that handles these cases).

		if err := upgradestore.New(db).SetServiceVersion(ctx, fmt.Sprintf("%d.%d.0", plan.to.Major, plan.to.Minor)); err != nil {
			return err
		}
	}

	return nil
}

func RunOutOfBandMigrations(
	ctx context.Context,
	db database.DB,
	dryRun bool,
	up bool,
	animateProgress bool,
	registerMigrations oobmigration.RegisterMigratorsFunc,
	out *output.Output,
	ids []int,
) (err error) {
	if len(ids) != 0 {
		out.WriteLine(output.Linef(output.EmojiFingerPointRight, output.StyleReset, "Running out of band migrations %v", ids))
		if dryRun {
			return nil
		}
	}

	store := oobmigration.NewStoreWithDB(db)
	runner := oobmigration.NewRunnerWithDB(&observation.TestContext, db, time.Second)
	if err := runner.SynchronizeMetadata(ctx); err != nil {
		return err
	}
	if err := registerMigrations(ctx, db, runner); err != nil {
		return err
	}

	if len(ids) == 0 {
		migrations, err := store.List(ctx)
		if err != nil {
			return err
		}

		for _, migration := range migrations {
			ids = append(ids, migration.ID)
		}
	}
	sort.Ints(ids)

	if dryRun {
		return nil
	}

	if err := runner.UpdateDirection(ctx, ids, !up); err != nil {
		return err
	}

	go runner.StartPartial(ids)
	defer runner.Stop()
	defer func() {
		if err == nil {
			out.WriteLine(output.Line(output.EmojiSuccess, output.StyleSuccess, "Out of band migrations complete"))
		} else {
			out.WriteLine(output.Linef(output.EmojiFailure, output.StyleFailure, "Out of band migrations failed: %s", err))
		}
	}()

	updateMigrationProgress, cleanup := oobmigration.MakeProgressUpdater(out, ids, animateProgress)
	defer cleanup()

	ticker := time.NewTicker(time.Second).C
	for {
		migrations, err := store.GetByIDs(ctx, ids)
		if err != nil {
			return err
		}
		sort.Slice(migrations, func(i, j int) bool { return migrations[i].ID < migrations[j].ID })

		for i, m := range migrations {
			updateMigrationProgress(i, m)
		}

		complete := true
		for _, m := range migrations {
			if !m.Complete() {
				if m.ApplyReverse && m.NonDestructive {
					continue
				}

				complete = false
			}
		}
		if complete {
			return nil
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker:
		}
	}
}
