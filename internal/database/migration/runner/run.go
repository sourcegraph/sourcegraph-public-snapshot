package runner

import (
	"context"
	"fmt"

	"github.com/cockroachdb/errors"
	"github.com/inconshreveable/log15"
)

type Options struct {
	Up            bool
	NumMigrations int
	SchemaNames   []string

	// Parallel controls whether we run schema migrations concurrently or not. By default,
	// we run schema migrations sequentially. This is to ensure that in testing, where the
	// same database can be targetted by multiple schemas, we do not hit errors that occur
	// when trying to install Postgres extensions concurrently (which do not seem txn-safe).
	Parallel bool
}

func (r *Runner) Run(ctx context.Context, options Options) error {
	numRoutines := 1
	if options.Parallel {
		numRoutines = len(options.SchemaNames)
	}
	semaphore := make(chan struct{}, numRoutines)

	return r.forEachSchema(ctx, options.SchemaNames, func(ctx context.Context, schemaName string, schemaContext schemaContext) error {
		// Block until we can write into this channel. This ensures that we only have at most
		// the same number of active goroutines as we have slots int he channel's buffer.
		semaphore <- struct{}{}
		defer func() { <-semaphore }()

		if err := r.runSchema(ctx, options, schemaContext); err != nil {
			return errors.Wrapf(err, "failed to run migration for schema %q", schemaName)
		}

		return nil
	})
}

func (r *Runner) runSchema(ctx context.Context, options Options, schemaContext schemaContext) (err error) {
	// Determine if we are upgrading to the latest schema. There are some properties around
	// contention which we want to accept on normal "upgrade to latest" behavior, but want to
	// alert on when a user is downgrading or upgrading to a specific version.
	upgradingToLatest := options.Up && options.NumMigrations == 0

	if !upgradingToLatest {
		// If the database is dirty, then either the last attempted migration had failed,
		// or another migrator is currently running and holding an advisory lock. If we're
		// not migrating to the latest schema, concurrent migrations may have unexpected
		// behavior. In either case, we'll early exit here.
		if schemaContext.schemaVersion.dirty {
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

	version, dirty, err := r.fetchVersion(ctx, schemaContext.schema.Name, schemaContext.store)
	if err != nil {
		return err
	}
	if !upgradingToLatest {
		// Check if another instance changed the schema version before we acquired the
		// lock. If we're not migrating to the latest schema, concurrent migrations may
		// have unexpected behavior. We'll early exit here.
		if version != schemaContext.schemaVersion.version {
			return errMigrationContention
		}
	}
	if dirty {
		// The store layer will refuse to alter a dirty database. We'll return an error
		// here instead of from the store as we can provide a bit instruction to the user
		// at this point.
		return errDirtyDatabase
	}

	if options.Up {
		return r.runSchemaUp(ctx, options, schemaContext)
	}
	return r.runSchemaDown(ctx, options, schemaContext)
}

func (r *Runner) runSchemaUp(ctx context.Context, options Options, schemaContext schemaContext) (err error) {
	log15.Info("Upgrading schema", "schema", schemaContext.schema.Name)

	definitions, err := schemaContext.schema.Definitions.UpFrom(schemaContext.schemaVersion.version, options.NumMigrations)
	if err != nil {
		return err
	}

	for _, definition := range definitions {
		log15.Info("Running up migration", "schema", schemaContext.schema.Name, "migrationID", definition.ID)

		if err := schemaContext.store.Up(ctx, definition); err != nil {
			return errors.Wrapf(err, "failed upgrade migration %d", definition.ID)
		}
	}

	return nil
}

func (r *Runner) runSchemaDown(ctx context.Context, options Options, schemaContext schemaContext) error {
	log15.Info("Downgrading schema", "schema", schemaContext.schema.Name)

	definitions, err := schemaContext.schema.Definitions.DownFrom(schemaContext.schemaVersion.version, options.NumMigrations)
	if err != nil {
		return err
	}

	for _, definition := range definitions {
		log15.Info("Running down migration", "schema", schemaContext.schema.Name, "migrationID", definition.ID)

		if err := schemaContext.store.Down(ctx, definition); err != nil {
			return errors.Wrapf(err, "failed downgrade migration %d", definition.ID)
		}
	}

	return nil
}
