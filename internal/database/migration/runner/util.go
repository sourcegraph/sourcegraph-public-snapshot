package runner

import (
	"context"
	"strconv"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/database/migration/definition"
)

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
) (pendingDefinitions, failedDefinitions []definition.Definition, _ error) {
	for _, definition := range definitions {
		if definition.IsCreateIndexConcurrently {
			tableName := definition.IndexMetadata.TableName
			indexName := definition.IndexMetadata.IndexName

			if status, ok, err := schemaContext.store.IndexStatus(ctx, tableName, indexName); err != nil {
				return nil, nil, err
			} else if ok && status.Phase != nil {
				pendingDefinitions = append(pendingDefinitions, definition)
				continue
			}
		}

		failedDefinitions = append(failedDefinitions, definition)
	}

	return pendingDefinitions, failedDefinitions, nil
}

func extractIDs(definitions []definition.Definition) []int {
	ids := make([]int, 0, len(definitions))
	for _, definition := range definitions {
		ids = append(ids, definition.ID)
	}

	return ids
}

func intSet(vs []int) map[int]struct{} {
	m := make(map[int]struct{}, len(vs))
	for _, v := range vs {
		m[v] = struct{}{}
	}

	return m
}

func intsToStrings(ints []int) []string {
	strs := make([]string, 0, len(ints))
	for _, value := range ints {
		strs = append(strs, strconv.Itoa(value))
	}

	return strs
}

func wait(ctx context.Context, duration time.Duration) error {
	select {
	case <-time.After(duration):
		return nil

	case <-ctx.Done():
		return ctx.Err()
	}
}
