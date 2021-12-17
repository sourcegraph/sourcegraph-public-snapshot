package runner

import (
	"context"
	"fmt"

	"github.com/sourcegraph/sourcegraph/internal/database/migration/definition"
)

func (r *Runner) Validate(ctx context.Context, schemaNames ...string) error {
	return r.forEachSchema(ctx, schemaNames, func(ctx context.Context, schemaName string, schemaContext schemaContext) error {
		return r.validateSchema(ctx, schemaName, schemaContext)
	})
}

func (r *Runner) validateSchema(ctx context.Context, schemaName string, schemaContext schemaContext) error {
	// If database version is strictly newer, then we have a deployment in-process. The current
	// instance has what it needs to run, so we should be good with that. Do not crash here if
	// the database is is dirty, as that would cause a troublesome deployment to cause outages
	// on the old instance (which seems stressful, don't do that).
	if newer, err := isDatabaseNewer(schemaContext.schemaVersion.version, schemaContext.schema.Definitions); err != nil {
		return err
	} else if newer {
		return nil
	}

	version, dirty, err := r.waitForMigration(ctx, schemaName, schemaContext)
	if err != nil {
		return err
	}

	// Note: No migrator instances seem to be running indicating that the dirty flag indicates
	// an actual migration failure that needs attention from a site administrator. We'll handle
	// the dirty flag selectively below.

	definitions, err := schemaContext.schema.Definitions.UpFrom(version, 0)
	if err != nil {
		// An error here means we might just be a very old instance. In order to figure out what
		// version we expect to be at, we re-query from a "blank" database so that we can take
		// populate the definitions variable in the error construction in the function below.

		return withAllDefinitions(schemaContext.schema.Definitions, func(allDefinitions []definition.Definition) error {
			if len(allDefinitions) == 0 {
				return err
			}

			return &SchemaOutOfDateError{
				schemaName:      schemaName,
				currentVersion:  version,
				expectedVersion: allDefinitions[len(allDefinitions)-1].ID,
			}
		})
	}
	if dirty {
		// Check again to see if the database is newer and ignore the dirty flag here as wlel.
		if newer, err := isDatabaseNewer(version, schemaContext.schema.Definitions); err != nil {
			return err
		} else if newer {
			return nil
		}

		// We have migrations to run but won't be able to run them
		return errDirtyDatabase
	}

	if len(definitions) == 0 {
		// No migrations to run, up to date
		return nil
	}

	return &SchemaOutOfDateError{
		schemaName:      schemaName,
		currentVersion:  version,
		expectedVersion: definitions[len(definitions)-1].ID,
	}
}

// waitForMigration polls the store for the version while taking an advisory lock. We do
// this while a migrator seems to be running concurrently so that we do not fail fast on
// applications that would succeed after the migration finishes.
func (r *Runner) waitForMigration(ctx context.Context, schemaName string, schemaContext schemaContext) (int, bool, error) {
	version, dirty := schemaContext.schemaVersion.version, schemaContext.schemaVersion.dirty

	for dirty {
		// While the previous version of the schema we queried was marked as dirty, we
		// will block until we can acquire the migration lock, then re-check the version.
		newVersion, newDirty, err := r.lockedVersion(ctx, schemaContext)
		if err != nil {
			return 0, false, err
		}
		if newVersion == version {
			// Version didn't change, no migrator instance was running and we were able
			// to acquire the lock right away. Break from this loop otherwise we'll just
			// be busy-querying the same state.
			break
		}

		// Version changed, check again
		version, dirty = newVersion, newDirty
	}

	return version, dirty, nil
}

func (r *Runner) lockedVersion(ctx context.Context, schemaContext schemaContext) (_ int, _ bool, err error) {
	if locked, unlock, err := schemaContext.store.Lock(ctx); err != nil {
		return 0, false, err
	} else if !locked {
		return 0, false, fmt.Errorf("failed to acquire migration lock")
	} else {
		defer func() { err = unlock(err) }()
	}

	return r.fetchVersion(ctx, schemaContext.schema.Name, schemaContext.store)
}

func withAllDefinitions(definitions *definition.Definitions, f func(definitions []definition.Definition) error) error {
	allDefinitions, err := definitions.UpFrom(0, 0)
	if err != nil {
		return err
	}

	return f(allDefinitions)
}

// isDatabaseNewer returns true if the given version is strictly larger than the maximum migration
// identifier we expect to have applied. Returns an error only on malformed schema definitions.
func isDatabaseNewer(version int, definitions *definition.Definitions) (bool, error) {
	found := false
	if err := withAllDefinitions(definitions, func(allDefinitions []definition.Definition) error {
		if len(allDefinitions) != 0 {
			if expectedMinimumMigration := allDefinitions[len(allDefinitions)-1].ID; expectedMinimumMigration < version {
				found = true
			}
		}

		return nil
	}); err != nil {
		return false, err
	}

	return found, nil
}
