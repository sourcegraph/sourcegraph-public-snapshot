package runner

import (
	"context"
	"sync"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/database/migration/definition"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/schemas"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/storetypes"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type Runner struct {
	storeFactories map[string]StoreFactory
	schemas        []*schemas.Schema
}

type StoreFactory func(ctx context.Context) (Store, error)

func NewRunner(storeFactories map[string]StoreFactory) *Runner {
	return NewRunnerWithSchemas(storeFactories, schemas.Schemas)
}

func NewRunnerWithSchemas(storeFactories map[string]StoreFactory, schemas []*schemas.Schema) *Runner {
	return &Runner{
		storeFactories: storeFactories,
		schemas:        schemas,
	}
}

type schemaContext struct {
	schema               *schemas.Schema
	store                Store
	initialSchemaVersion schemaVersion
}

type schemaVersion struct {
	appliedVersions []int
	pendingVersions []int
	failedVersions  []int
}

type visitFunc func(ctx context.Context, schemaContext schemaContext) error

// Store returns the store associated with the given schema.
func (r *Runner) Store(ctx context.Context, schemaName string) (Store, error) {
	if factory, ok := r.storeFactories[schemaName]; ok {
		return factory(ctx)

	}

	return nil, errors.Newf("unknown store %q", schemaName)
}

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
		for _, schema := range r.schemas {
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
	appliedVersions, pendingVersions, failedVersions, err := store.Versions(ctx)
	if err != nil {
		return schemaVersion{}, err
	}

	return schemaVersion{
		appliedVersions,
		pendingVersions,
		failedVersions,
	}, nil
}

type lockedVersionCallback func(
	schemaVersion schemaVersion,
	byState definitionsByState,
	earlyUnlock unlockFunc,
) error

type unlockFunc func(err error) error

// withLockedSchemaState attempts to take an advisory lock, then re-checks the version of the
// database. The resulting schema state is passed to the given function. The advisory lock
// will be released on function exit, but the callback may explicitly release the lock earlier.
//
// This method returns a true-valued flag if it should be re-invoked by the caller.
func (r *Runner) withLockedSchemaState(
	ctx context.Context,
	schemaContext schemaContext,
	definitions []definition.Definition,
	f lockedVersionCallback,
) (retry bool, _ error) {
	// Take an advisory lock to determine if there are any migrator instances currently
	// running queries unrelated to non-concurrent index creation. This will block until
	// we are able to gain the lock.
	unlock, err := r.pollLock(ctx, schemaContext)
	if err != nil {
		return false, err
	} else {
		defer func() { err = unlock(err) }()
	}

	// Re-fetch the current schema of the database now that we hold the lock. This may differ
	// from our original assumption if another migrator is running concurrently.
	schemaVersion, err := r.fetchVersion(ctx, schemaContext.schema.Name, schemaContext.store)
	if err != nil {
		return false, err
	}

	// Filter out any unlisted migrations (most likely future upgrades) and group them by status.
	byState := groupByState(schemaVersion, definitions)

	logger.Info(
		"Checked current schema state",
		"schema", schemaContext.schema.Name,
		"appliedVersions", extractIDs(byState.applied),
		"pendingVersions", extractIDs(byState.pending),
		"failedVersions", extractIDs(byState.failed),
	)

	// Detect failed migrations, and determine if we need to wait longer for concurrent migrator
	// instances to finish their current work.
	if retry, err := validateSchemaState(ctx, schemaContext, byState); err != nil {
		return false, err
	} else if retry {
		// An index is currently being created. We return true here to flag to the caller that
		// we should wait a small time, then be re-invoked. We don't want to take any action
		// here while the other proceses is working.
		return true, nil
	}

	// Invoke the callback with the current schema state
	return false, f(schemaVersion, byState, unlock)
}

const lockPollInterval = time.Second
const lockPollLogRatio = 5

// pollLock will attempt to acquire a session-level advisory lock while the given context has not
// been canceled. The caller must eventually invoke the unlock function on successful acquisition
// of the lock.
func (r *Runner) pollLock(ctx context.Context, schemaContext schemaContext) (unlock func(err error) error, _ error) {
	numWaits := 0

	for {
		if acquired, unlock, err := schemaContext.store.TryLock(ctx); err != nil {
			return nil, err
		} else if acquired {
			return unlock, nil
		}

		if numWaits%lockPollLogRatio == 0 {
			logger.Info(
				"Schema migration lock is currently held - will re-attempt to acquire lock",
				"schema", schemaContext.schema.Name,
			)
		}

		if err := wait(ctx, lockPollInterval); err != nil {
			return nil, err
		}

		numWaits++
	}
}

type definitionsByState struct {
	applied []definition.Definition
	pending []definition.Definition
	failed  []definition.Definition
}

// groupByState returns the the given definitions grouped by their status (applied, pending, failed) as
// indicated by the current schema.
func groupByState(schemaVersion schemaVersion, definitions []definition.Definition) definitionsByState {
	appliedVersionsMap := intSet(schemaVersion.appliedVersions)
	failedVersionsMap := intSet(schemaVersion.failedVersions)
	pendingVersionsMap := intSet(schemaVersion.pendingVersions)

	states := definitionsByState{}
	for _, definition := range definitions {
		if _, ok := appliedVersionsMap[definition.ID]; ok {
			states.applied = append(states.applied, definition)
		}
		if _, ok := pendingVersionsMap[definition.ID]; ok {
			states.pending = append(states.pending, definition)
		}
		if _, ok := failedVersionsMap[definition.ID]; ok {
			states.failed = append(states.failed, definition)
		}
	}

	return states
}

// validateSchemaState inspects the given definitions grouped by state and determines if the schema
// state should be re-queried (when `retry` is true). This function returns an error if the database
// is in a dirty state (contains failed migrations or pending migrations without a backing query).
func validateSchemaState(
	ctx context.Context,
	schemaContext schemaContext,
	byState definitionsByState,
) (retry bool, _ error) {
	if len(byState.failed) > 0 {
		// Explicit failures require administrator intervention
		return false, newDirtySchemaError(schemaContext.schema.Name, byState.failed)
	}

	if len(byState.pending) > 0 {
		// We are currently holding the lock, so any migrations that are "pending" are either
		// dead and the migrator instance has died before finishing the operation, or they're
		// active concurrent index creation operations. We'll partition this set into those two
		// groups and determine what to do.
		if pendingDefinitions, failedDefinitions, err := partitionPendingMigrations(ctx, schemaContext, byState.pending); err != nil {
			return false, err
		} else if len(failedDefinitions) > 0 {
			// Explicit failures require administrator intervention
			return false, newDirtySchemaError(schemaContext.schema.Name, failedDefinitions)
		} else if len(pendingDefinitions) > 0 {
			return true, nil
		}
	}

	return false, nil
}

type definitionWithStatus struct {
	definition  definition.Definition
	indexStatus storetypes.IndexStatus
}

// partitionPendingMigrations partitions the given migrations into two sets: the set of pending
// migration definitions, which includes migrations with visible and active create index operation
// running in the database, and the set of filed migration definitions, which includes migrations
// which are marked as pending but do not appear as active.
//
// This function assumes that the migration advisory lock is held.
func partitionPendingMigrations(
	ctx context.Context,
	schemaContext schemaContext,
	definitions []definition.Definition,
) (pendingDefinitions []definitionWithStatus, failedDefinitions []definition.Definition, _ error) {
	for _, definition := range definitions {
		if definition.IsCreateIndexConcurrently {
			tableName := definition.IndexMetadata.TableName
			indexName := definition.IndexMetadata.IndexName

			if indexStatus, ok, err := schemaContext.store.IndexStatus(ctx, tableName, indexName); err != nil {
				return nil, nil, err
			} else if ok && indexStatus.Phase != nil {
				pendingDefinitions = append(pendingDefinitions, definitionWithStatus{definition, indexStatus})
				continue
			}
		}

		failedDefinitions = append(failedDefinitions, definition)
	}

	return pendingDefinitions, failedDefinitions, nil
}
