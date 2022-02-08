package runner

import (
	"context"
	"sync"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/database/migration/definition"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/schemas"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type Runner struct {
	storeFactories map[string]StoreFactory
}

type StoreFactory func(ctx context.Context) (Store, error)

func NewRunner(storeFactories map[string]StoreFactory) *Runner {
	return &Runner{
		storeFactories: storeFactories,
	}
}

type schemaContext struct {
	schema               *schemas.Schema
	store                Store
	initialSchemaVersion schemaVersion
}

type schemaVersion struct {
	version         int
	dirty           bool
	appliedVersions []int
	pendingVersions []int
	failedVersions  []int
}

type visitFunc func(ctx context.Context, schemaContext schemaContext) error

// forEachSchema invokes the given function once for each schema in the given list, with
// store instances initialized for each given schema name. Each function invocation occurs
// concurrently. Errors from each invocation are collected and returned. An error from one
// goroutine will not cancel the progress of another.
func (r *Runner) forEachSchema(ctx context.Context, schemaNames []string, visitor visitFunc) error {
	// Create map of relevant schemas keyed by name
	schemaMap, err := r.prepareSchemas(schemaNames)
	if err != nil {
		return err
	}

	// Create map of migration stores keyed by name
	storeMap, err := r.prepareStores(ctx, schemaNames)
	if err != nil {
		return err
	}

	// Create map of versions keyed by name
	versionMap, err := r.fetchVersions(ctx, storeMap)
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	errorCh := make(chan error, len(schemaNames))

	for _, schemaName := range schemaNames {
		wg.Add(1)

		go func(schemaName string) {
			defer wg.Done()

			errorCh <- visitor(ctx, schemaContext{
				schema:               schemaMap[schemaName],
				store:                storeMap[schemaName],
				initialSchemaVersion: versionMap[schemaName],
			})
		}(schemaName)
	}

	wg.Wait()
	close(errorCh)

	var errs *errors.MultiError
	for err := range errorCh {
		if err != nil {
			errs = errors.Append(errs, err)
		}
	}

	return errs.ErrorOrNil()
}

func (r *Runner) prepareSchemas(schemaNames []string) (map[string]*schemas.Schema, error) {
	schemaMap := make(map[string]*schemas.Schema, len(schemaNames))

	for _, targetSchemaName := range schemaNames {
		for _, schema := range schemas.Schemas {
			if schema.Name == targetSchemaName {
				schemaMap[schema.Name] = schema
				break
			}
		}
	}

	// Ensure that all supplied schema names are valid
	for _, schemaName := range schemaNames {
		if _, ok := schemaMap[schemaName]; !ok {
			return nil, errors.Newf("unknown schema %q", schemaName)
		}
	}

	return schemaMap, nil
}

func (r *Runner) prepareStores(ctx context.Context, schemaNames []string) (map[string]Store, error) {
	storeMap := make(map[string]Store, len(schemaNames))

	for _, schemaName := range schemaNames {
		storeFactory, ok := r.storeFactories[schemaName]
		if !ok {
			return nil, errors.Newf("unknown schema %q", schemaName)
		}

		store, err := storeFactory(ctx)
		if err != nil {
			return nil, err
		}

		storeMap[schemaName] = store
	}

	return storeMap, nil
}

func (r *Runner) fetchVersions(ctx context.Context, storeMap map[string]Store) (map[string]schemaVersion, error) {
	versions := make(map[string]schemaVersion, len(storeMap))

	for schemaName, store := range storeMap {
		schemaVersion, err := r.fetchVersion(ctx, schemaName, store)
		if err != nil {
			return nil, err
		}

		versions[schemaName] = schemaVersion
	}

	return versions, nil
}

func (r *Runner) fetchVersion(ctx context.Context, schemaName string, store Store) (schemaVersion, error) {
	version, dirty, _, err := store.Version(ctx)
	if err != nil {
		return schemaVersion{}, err
	}
	appliedVersions, pendingVersions, failedVersions, err := store.Versions(ctx)
	if err != nil {
		return schemaVersion{}, err
	}

	logger.Info(
		"Checked current version",
		"schema", schemaName,
		"version", version,
		"dirty", dirty,
		"appliedVersions", appliedVersions,
		"pendingVersions", pendingVersions,
		"failedVersions", failedVersions,
	)

	return schemaVersion{
		version,
		dirty,
		appliedVersions,
		pendingVersions,
		failedVersions,
	}, nil
}

const lockPollInterval = time.Second

// pollLock will attempt to acquire a session-level advisory lock while the given context has not
// been canceled. The caller must eventually invoke the unlock function on successful acquisition
// of the lock.
func (r *Runner) pollLock(ctx context.Context, store Store) (unlock func(err error) error, _ error) {
	for {
		if acquired, unlock, err := store.TryLock(ctx); err != nil {
			return nil, err
		} else if acquired {
			return unlock, nil
		}

		select {
		case <-time.After(lockPollInterval):
			continue
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
}

type lockedVersionCallback func(
	schemaVersion schemaVersion,
	byState definitionsByState,
) error

// withLockedSchemaState attempts to take an advisory lock, then re-checks the version of the
// database. The resulting schema state is passed to the given function. The advisory lock
// will be released on function exit.
func (r *Runner) withLockedSchemaState(
	ctx context.Context,
	schemaContext schemaContext,
	definitions []definition.Definition,
	f lockedVersionCallback,
) (err error) {
	// Take an advisory lock to determine if there are any migrator instances currently
	// running queries unrelated to non-concurrent index creation. This will block until
	// we are able to gain the lock.
	unlock, err := r.pollLock(ctx, schemaContext.store)
	if err != nil {
		return err
	} else {
		defer func() { err = unlock(err) }()
	}

	// Re-fetch the current schema of the database now that we hold the lock. This may differ
	// from our original assumption if another migrator is running concurrently.
	schemaVersion, err := r.fetchVersion(ctx, schemaContext.schema.Name, schemaContext.store)
	if err != nil {
		return err
	}
	byState := groupByState(schemaVersion, definitions)

	// Detect failed migrations prior to the callback
	if err := validateSchemaState(ctx, schemaContext, byState); err != nil {
		return err
	}

	// Invoke the callback with the current schema state
	return f(schemaVersion, byState)
}
