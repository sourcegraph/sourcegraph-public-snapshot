package runner

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/database/migration/definition"
	"github.com/sourcegraph/sourcegraph/lib/errors"
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
		return errors.Newf("multiple operations defined on the same schema")
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
// method will attempt to coordinate with other concurrently running instances and may block while
// attempting to acquire a lock. An error is returned only if user intervention is deemed a necessity,
// the "dirty database" condition, or on context cancellation.
func (r *Runner) runSchema(ctx context.Context, operation MigrationOperation, schemaContext schemaContext) error {
	// First, rewrite operations into a smaller set of operations we'll handle below. This call converts
	// upgrade and revert operations into targeted up and down operations.
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

	return r.applyMigrations(ctx, operation, schemaContext, definitions)
}

// applyMigrations attempts to take an advisory lock, then re-checks the version of the database. If there
// are still migrations to apply from the given definitions, they are applied in-order.
func (r *Runner) applyMigrations(
	ctx context.Context,
	operation MigrationOperation,
	schemaContext schemaContext,
	definitions []definition.Definition,
) error {
	up := operation.Type == MigrationOperationTypeTargetedUp

	callback := func(schemaVersion schemaVersion, _ definitionsByState) error {
		// Filter the set of definitions we still need to apply given our new view of the schema
		definitions := filterAppliedDefinitions(schemaVersion, operation, definitions)
		if len(definitions) == 0 {
			return nil
		}

		logger.Info(
			"Applying migrations",
			"schema", schemaContext.schema.Name,
			"up", up,
			"count", len(definitions),
		)

		for _, definition := range definitions {
			// Apply all other types of migrations uniformly
			if err := r.applyMigration(ctx, schemaContext, operation, definition); err != nil {
				return err
			}
		}

		return nil
	}

	return r.withLockedSchemaState(ctx, schemaContext, definitions, callback)
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
