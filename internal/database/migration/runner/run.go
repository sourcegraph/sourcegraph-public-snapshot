package runner

import (
	"context"
	"fmt"

	"github.com/cockroachdb/errors"

	"github.com/sourcegraph/sourcegraph/internal/database/migration/definition"
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
func (r *Runner) runSchema(ctx context.Context, operation MigrationOperation, schemaContext schemaContext) (err error) {
	// First, rewrite operations into a smaller set of operations we'll handle below. This call converts
	// upgrade and revert operations into targeted up and down operations.
	operation, err = desugarOperation(schemaContext, operation)
	if err != nil {
		return err
	}

	// Determine if we are upgrading to the latest schema. There are some properties around
	// contention which we want to accept on normal "upgrade to latest" behavior, but want to
	// alert on when a user is downgrading or upgrading to a specific version.
	upgradingToLatest := operation.Type == MigrationOperationTypeTargetedUp && operation.TargetVersion == 0

	if !upgradingToLatest {
		// If the database is dirty, then either the last attempted migration had failed,
		// or another migrator is currently running and holding an advisory lock. If we're
		// not migrating to the latest schema, concurrent migrations may have unexpected
		// behavior. In either case, we'll early exit here.
		if schemaContext.initialSchemaVersion.dirty {
			acquired, unlock, err := schemaContext.store.TryLock(ctx)
			if err != nil {
				return err
			}
			defer func() { err = unlock(err) }()

			if !acquired {
				// Some other migration process is holding the lock
				return errMigrationContention
			}

			return errDirtyDatabase
		}
	}

	if acquired, unlock, err := schemaContext.store.Lock(ctx); err != nil {
		return err
	} else if !acquired {
		return fmt.Errorf("failed to acquire migration lock")
	} else {
		defer func() { err = unlock(err) }()
	}

	schemaVersion, err := r.fetchVersion(ctx, schemaContext.schema.Name, schemaContext.store)
	if err != nil {
		return err
	}
	if !upgradingToLatest {
		// Check if another instance changed the schema version before we acquired the
		// lock. If we're not migrating to the latest schema, concurrent migrations may
		// have unexpected behavior. We'll early exit here.
		if schemaVersion.version != schemaContext.initialSchemaVersion.version {
			return errMigrationContention
		}
	}
	if schemaVersion.dirty {
		// The store layer will refuse to alter a dirty database. We'll return an error
		// here instead of from the store as we can provide a bit instruction to the user
		// at this point.
		return errDirtyDatabase
	}

	if operation.Type == MigrationOperationTypeTargetedUp {
		return r.runSchemaUp(ctx, operation, schemaContext)
	}
	return r.runSchemaDown(ctx, operation, schemaContext)
}

func (r *Runner) applyMigrations(ctx context.Context, operation MigrationOperation, schemaContext schemaContext, definitions []definition.Definition) error {
	logger.Info(
		"Applying migrations",
		"schema", schemaContext.schema.Name,
		"type", operation.Type,
		"count", len(definitions),
	)

	for _, definition := range definitions {
		if err := r.applyMigration(ctx, schemaContext, operation, definition); err != nil {
			return err
		}
	}

	return nil
}

func (r *Runner) runSchemaUp(ctx context.Context, operation MigrationOperation, schemaContext schemaContext) (err error) {
	logger.Info("Upgrading schema", "schema", schemaContext.schema.Name)

	definitions, err := schemaContext.schema.Definitions.UpTo(schemaContext.initialSchemaVersion.version, operation.TargetVersion)
	if err != nil {
		return err
	}

	return r.applyMigrations(ctx, operation, schemaContext, definitions)
}

func (r *Runner) runSchemaDown(ctx context.Context, operation MigrationOperation, schemaContext schemaContext) error {
	logger.Info("Downgrading schema", "schema", schemaContext.schema.Name)

	if operation.TargetVersion == 0 {
		operation.TargetVersion = schemaContext.initialSchemaVersion.version - 1
	}

	definitions, err := schemaContext.schema.Definitions.DownTo(schemaContext.initialSchemaVersion.version, operation.TargetVersion)
	if err != nil {
		return err
	}

	return r.applyMigrations(ctx, operation, schemaContext, definitions)
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
