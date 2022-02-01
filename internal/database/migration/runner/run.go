package runner

import (
	"context"
	"fmt"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/jackc/pgconn"

	"github.com/sourcegraph/sourcegraph/internal/database/migration/definition"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/storetypes"
)

func (r *Runner) Run(ctx context.Context, options Options) error {
	schemaNames := make([]string, 0, len(options.Operations))
	for _, operation := range options.Operations {
		schemaNames = append(schemaNames, operation.SchemaName)
	}

	operationMap := make(map[string]MigrationOperation, len(options.Operations))
	for _, operation := range options.Operations {
		operationMap[operation.SchemaName] = operation
	}
	if len(operationMap) != len(options.Operations) {
		return fmt.Errorf("multiple operations defined on the same schema")
	}

	numRoutines := 1
	if options.Parallel {
		numRoutines = len(schemaNames)
	}
	semaphore := make(chan struct{}, numRoutines)

	return r.forEachSchema(ctx, schemaNames, func(ctx context.Context, schemaContext schemaContext) error {
		schemaName := schemaContext.schema.Name

		// Block until we can write into this channel. This ensures that we only have at most
		// the same number of active goroutines as we have slots in the channel's buffer.
		semaphore <- struct{}{}
		defer func() { <-semaphore }()

		if err := r.runSchema(ctx, operationMap[schemaName], schemaContext); err != nil {
			return errors.Wrapf(err, "failed to run migration for schema %q", schemaName)
		}

		return nil
	})
}

// runSchema applies (or unapplies) the set of migrations required to fulfill the given operation. This
// method will attempt to coordinate with with other concurrently running instances and may block while
// attempting to acquire a lock. An error is returned only if user intervention is deemed a necessity,
// the "dirty database" condition, or on context cancellation.
func (r *Runner) runSchema(ctx context.Context, operation MigrationOperation, schemaContext schemaContext) error {
	// First, rewrite operations into a smaller set of operations we'll handle below. This call converts
	// upgrade and revert operations into targeted upto and downto operations.
	operation, err := desugarOperation(schemaContext, operation)
	if err != nil {
		return err
	}

	gatherDefinitions := schemaContext.schema.Definitions.Up
	if operation.Type != MigrationOperationTypeTargetedUp {
		gatherDefinitions = schemaContext.schema.Definitions.Down
	}

	// Get the set of migrations that need to be applied or unapplied, depending on the migration direction.
	definitions, err := gatherDefinitions(schemaContext.initialSchemaVersion.appliedVersions, operation.TargetVersions)
	if err != nil {
		return err
	}

	// Before we commit to performing an upgrade (which takes locks), determine if there is anything to do
	// and early out if not. We'll no-op if there are no definitions with pending or failed attempts, and
	// all migrations are applied (when migrating up) or unapplied (when migrating down).
	if byState := groupByState(schemaContext.initialSchemaVersion, definitions); len(byState.pending)+len(byState.failed) == 0 {
		if operation.Type == MigrationOperationTypeTargetedUp && len(byState.applied) == len(definitions) {
			return nil
		}

		if operation.Type == MigrationOperationTypeTargetedDown && len(byState.applied) == 0 {
			return nil
		}
	}

	for {
		// Attempt to apply as many migrations as possible. We do this iteratively in chunks as we are unable
		// to hold a consistent advisory lock in the presence of migrations utilizing concurrent index creation.
		// Therefore, some invocations of this method will return with a flag to request re-invocation under a
		// new lock.

		if retry, err := r.applyMigrations(ctx, operation, schemaContext, definitions); err != nil {
			return err
		} else if !retry {
			break
		}
	}

	return nil
}

// applyMigrations attempts to take an advisory lock, then re-checks the version of the database. If there are
// still migrations to apply from the given definitions, they are applied in-order. If not all definitions are
// applied, this method returns a true-valued flag indicating that it should be re-invoked to attempt to apply
// the remaining definitions.
func (r *Runner) applyMigrations(
	ctx context.Context,
	operation MigrationOperation,
	schemaContext schemaContext,
	definitions []definition.Definition,
) (retry bool, _ error) {
	var (
		droppedLock bool
		up          = operation.Type == MigrationOperationTypeTargetedUp
	)

	callback := func(schemaVersion schemaVersion, _ definitionsByState, earlyUnlock unlockFunc) error {
		// Filter the set of definitions we still need to apply given our new view of the schema
		definitions := filterAppliedDefinitions(schemaVersion, operation, definitions)
		if len(definitions) == 0 {
			// Stop retry loop
			return nil
		}

		logger.Info(
			"Applying migrations",
			"schema", schemaContext.schema.Name,
			"up", up,
			"count", len(definitions),
		)

		for _, definition := range definitions {
			if up && definition.IsCreateIndexConcurrently {
				// Handle execution of `CREATE INDEX CONCURRENTLY` specially
				if unlocked, err := r.createIndexConcurrently(ctx, schemaContext, definition, earlyUnlock); err != nil {
					return err
				} else if unlocked {
					// We've forfeited our lock, but want to continue applying the remaining migrations (if any).
					// Setting this value here sends us back to the caller to be re-invoked.
					droppedLock = true
					return nil
				}
			} else {
				// Apply all other types of migrations uniformly
				if err := r.applyMigration(ctx, schemaContext, operation, definition); err != nil {
					return err
				}
			}
		}

		// Stop retry loop
		return nil
	}

	retry, err := r.withLockedSchemaState(ctx, schemaContext, definitions, callback)
	return retry || droppedLock, err
}

// applyMigration applies the given migration in the direction indicated by the given operation.
func (r *Runner) applyMigration(
	ctx context.Context,
	schemaContext schemaContext,
	operation MigrationOperation,
	definition definition.Definition,
) error {
	up := operation.Type == MigrationOperationTypeTargetedUp

	logger.Info(
		"Applying migration",
		"schema", schemaContext.schema.Name,
		"migrationID", definition.ID,
		"up", up,
	)

	direction := schemaContext.store.Up
	if !up {
		direction = schemaContext.store.Down
	}

	applyMigration := func() error {
		return direction(ctx, definition)
	}
	if err := schemaContext.store.WithMigrationLog(ctx, definition, up, applyMigration); err != nil {
		return errors.Wrapf(err, "failed to apply migration %d", definition.ID)
	}

	return nil
}

const indexPollInterval = time.Second

// createIndexConcurrently deals with the special case of `CREATE INDEX CONCURRENTLY` migrations. We cannot
// hold an advisory lock during concurrent index creation without trivially deadlocking concurrent migrator
// instances (see `interrnal/database/migration/store/store_test.go:TestIndexStatus` for an example). Instead,
// we use Postgres system tables to determine the status of the index being created and re-issue the index
// creation command if it's missing or invalid.
//
// If the given `unlock` function is called then `unlocked` will be true on return. This allows the caller
// to maintain the lock in the case that the index already exists due to an out-of-band operation.
func (r *Runner) createIndexConcurrently(
	ctx context.Context,
	schemaContext schemaContext,
	definition definition.Definition,
	unlock func(err error) error,
) (unlocked bool, _ error) {
	tableName := definition.IndexMetadata.TableName
	indexName := definition.IndexMetadata.IndexName

pollIndexStatusLoop:
	for {
		// Query the current status of the target index
		status, exists, err := schemaContext.store.IndexStatus(ctx, tableName, indexName)
		if err != nil {
			return false, errors.Wrap(err, "failed to query state of index")
		}

		logger.Info(
			"Checked progress of index creation",
			append(
				[]interface{}{
					"tableName", tableName,
					"indexName", indexName,
					"exists", exists,
					"isValid", status.IsValid,
				},
				renderIndexStatus(status)...,
			)...,
		)

		if exists && status.IsValid {
			// Index exists and is valid; nothing to do
			return unlocked, nil
		}

		if exists && status.Phase == nil {
			// Index is invalid but no creation operation is in-progress. We can try to repair this
			// state automatically by dropping the index and re-create it as if it never existed.
			// Assuming that the down migration drops the index created in the up direction, we'll
			// just apply that. We open a (hopefully) short-lived transaction here to drop the
			// existing index and write the migration log entry in the same shot.

			tx, err := schemaContext.store.Transact(ctx)
			if err != nil {
				return false, err
			}

			dropIndex := func() error {
				return tx.Down(ctx, definition)
			}
			if err := tx.WithMigrationLog(ctx, definition, false, dropIndex); err != nil {
				// Ensure we don't leak txn on error here
				return false, tx.Done(err)
			}

			if err := tx.Done(nil); err != nil {
				return false, err
			}
		}

		// Release advisory lock before attempting to create index or wait on the the index creation
		// operation. Concurrent index creation works in several distinct phases. One of those phases
		// requires that all existant transactions finish. If we hold an advisory lock in this session
		// that blocks another transaction, the index creation operation will deadlock and fail.
		//
		// Note that we assume idempotency on this unlock function.
		if err := unlock(nil); err != nil {
			return false, err
		}
		unlocked = true

		// Index is currently being created. Wait a small time and check the index status again. We don't
		// want to take any action here while the other proceses is working.
		if exists && status.Phase != nil {
			select {
			case <-time.After(indexPollInterval):
				continue pollIndexStatusLoop

			case <-ctx.Done():
				return unlocked, ctx.Err()
			}
		}

		// Create the index. Ignore duplicate table/index already exists errors. This can happen if there
		// is a race between two migrator instances fighting to create the same index. Just silence the
		// error within the `createIndex` function (so we prevent a spurious migration log failure entry)
		// and set a flag indicating a to retry the operation. We retry instead of returning so that we
		// do not prematurely begin to apply the next migration, which may assume the successful creation
		// of the index.

		var (
			pgErr        pgconn.PgError
			raceDetected bool

			errorFilter = func(err error) error {
				if err == nil {
					return err
				}
				if !errors.As(err, pgErr) || pgErr.Code != "42P07" {
					return err
				}

				raceDetected = true
				return nil
			}
		)

		logger.Info(
			"Creating index concurrently",
			"schema", schemaContext.schema.Name,
			"migrationID", definition.ID,
			"tableName", tableName,
			"indexName", indexName,
		)

		createIndex := func() error {
			return errorFilter(schemaContext.store.Up(ctx, definition))
		}
		if err := schemaContext.store.WithMigrationLog(ctx, definition, true, createIndex); err != nil {
			return false, err
		} else if raceDetected {
			continue
		}

		return unlocked, nil
	}
}

// filterAppliedDefinitions returns a subset of the given definition slice. A definition will be included
// in the resulting slice if we're migrating up and the migration is not applied, or if we're migrating down
// and the migration is applied.
//
// The resulting slice will have the same relative order as the input slice. This function does not alter
// the input slice.
func filterAppliedDefinitions(
	schemaVersion schemaVersion,
	operation MigrationOperation,
	definitions []definition.Definition,
) []definition.Definition {
	up := operation.Type == MigrationOperationTypeTargetedUp
	appliedVersionMap := intSet(schemaVersion.appliedVersions)

	filtered := make([]definition.Definition, 0, len(definitions))
	for _, definition := range definitions {
		if _, ok := appliedVersionMap[definition.ID]; ok == up {
			// Either
			// - needs to be applied and already applied, or
			// - needs to be unapplied and not currently applied.
			continue
		}

		filtered = append(filtered, definition)
	}

	return filtered
}

// renderIndexStatus returns a slice of interface pairs describing the given index status for use in a
// call to logger. If the index is currently being created, the progress of the create operation will be
// summarized.
func renderIndexStatus(progress storetypes.IndexStatus) (logPairs []interface{}) {
	if progress.Phase == nil {
		return []interface{}{
			"in-progress", false,
		}
	}

	index := -1
	for i, phase := range storetypes.CreateIndexConcurrentlyPhases {
		if phase == *progress.Phase {
			index = i
			break
		}
	}

	return []interface{}{
		"in-progress", true,
		"phase", *progress.Phase,
		"phases", fmt.Sprintf("%d of %d", index, len(storetypes.CreateIndexConcurrentlyPhases)),
		"lockers", fmt.Sprintf("%d of %d", progress.LockersDone, progress.LockersTotal),
		"blocks", fmt.Sprintf("%d of %d", progress.BlocksDone, progress.BlocksTotal),
		"tuples", fmt.Sprintf("%d of %d", progress.TuplesDone, progress.TuplesTotal),
	}
}
