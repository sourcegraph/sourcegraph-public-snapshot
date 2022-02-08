package runner

import (
	"context"
	"strconv"

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

// validateSchemaState inspects the given definitions grouped by state and returns an error if
// the database is in a dirty state (i.e., contains failed migrations or pending migrations).
func validateSchemaState(
	ctx context.Context,
	schemaContext schemaContext,
	byState definitionsByState,
) error {
	if len(byState.failed) > 0 {
		// Explicit failures require administrator intervention
		return newDirtySchemaError(schemaContext.schema.Name, byState.failed)
	}

	if len(byState.pending) > 0 {
		// Explicit failures require administrator intervention
		return newDirtySchemaError(schemaContext.schema.Name, byState.pending)
	}

	return nil
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
