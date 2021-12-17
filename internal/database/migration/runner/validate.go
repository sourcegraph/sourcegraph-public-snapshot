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
	version, dirty := schemaContext.schemaVersion.version, schemaContext.schemaVersion.dirty

	for dirty {
		// While the previous version of the schema we queried was marked as dirty, we
		// will block until we can acquire the migration lock, then re-check the version.
		newVersion, newDirty, err := r.lockedVersion(ctx, schemaContext)
		if err != nil {
			return err
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

	withAllDefinitions := func(f func(definitions []definition.Definition) error) error {
		definitions, err := schemaContext.schema.Definitions.UpFrom(0, 0)
		if err != nil {
			return err
		}

		return f(definitions)
	}

	// Note: No migrator instances seem to be running indicating that the dirty flag indicates
	// an actual migration failure that needs attention from a site administrator. We'll handle
	// the dirty flag selectively below.

	definitions, err := schemaContext.schema.Definitions.UpFrom(version, 0)
	if err != nil {
		// An error here means we might just be a very old instance. In order to figure out what
		// version we expect to be at, we re-query from a "blank" database so that we can take
		// populate the definitions variable in the error construction in the function below.

		return withAllDefinitions(func(allDefinitions []definition.Definition) error {
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
	if len(definitions) == 0 {
		if dirty {
			// Check to see if the database is dirty but in a schema state farther along then what
			// we need. In this case, the schema may have been marked dirty during a failed upgrade
			// from the current running instance, which should not generally have any affect on the
			// stability of the active instance.
			//
			// In these cases, we ignore the dirty flag here. The dirty flag error should be obvious
			// to a site administrator via the migrator instance during the deploy.
			return withAllDefinitions(func(allDefinitions []definition.Definition) error {
				if len(allDefinitions) == 0 || version <= allDefinitions[len(allDefinitions)-1].ID {
					return errDirtyDatabase
				}

				return nil
			})
		}

		// No migrations to run, up to date
		return nil
	}
	if dirty {
		// We have migrations to run but won't be able to run them
		return errDirtyDatabase
	}

	return &SchemaOutOfDateError{
		schemaName:      schemaName,
		currentVersion:  version,
		expectedVersion: definitions[len(definitions)-1].ID,
	}
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
