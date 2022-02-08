package runner

import (
	"context"
)

func (r *Runner) Validate(ctx context.Context, schemaNames ...string) error {
	return r.forEachSchema(ctx, schemaNames, func(ctx context.Context, schemaContext schemaContext) error {
		return r.validateSchema(ctx, schemaContext)
	})
}

// validateSchema returns a non-nil error value if the target database schema is not in the state
// expected by the given schema context. This method will block if there are relevant migrations
// in progress.
func (r *Runner) validateSchema(ctx context.Context, schemaContext schemaContext) error {
	definitions := schemaContext.schema.Definitions.All()

	// Quickly determine with our initial schema version if we are up to date. If so, we won't need
	// to take an advisory lock and poll index creation status below.
	byState := groupByState(schemaContext.initialSchemaVersion, definitions)
	if len(byState.pending) == 0 && len(byState.failed) == 0 && len(byState.applied) == len(definitions) {
		return nil
	}

	// Take an advisory lock, then re-check the version of the database. If there are still migrations
	// to apply from the given definitions, return an error.
	return r.withLockedSchemaState(ctx, schemaContext, definitions, func(schemaVersion schemaVersion, byState definitionsByState) error {
		if len(byState.applied) != len(definitions) {
			// Return an error if all expected schemas have not been applied
			return newOutOfDateError(schemaContext, schemaVersion)
		}

		return nil
	})
}
