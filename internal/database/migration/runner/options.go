package runner

import (
	"fmt"

	"github.com/inconshreveable/log15"
)

type Options struct {
	Operations []MigrationOperation

	// Parallel controls whether we run schema migrations concurrently or not. By default,
	// we run schema migrations sequentially. This is to ensure that in testing, where the
	// same database can be targeted by multiple schemas, we do not hit errors that occur
	// when trying to install Postgres extensions concurrently (which do not seem txn-safe).
	Parallel bool
}

type MigrationOperation struct {
	SchemaName    string
	Type          MigrationOperationType
	TargetVersion int
}

type MigrationOperationType int

const (
	MigrationOperationTypeTargetedUp MigrationOperationType = iota
	MigrationOperationTypeTargetedDown
	MigrationOperationTypeUpgrade
	MigrationOperationTypeRevert
)

func desugarOperation(schemaContext schemaContext, operation MigrationOperation) (MigrationOperation, error) {
	switch operation.Type {
	case MigrationOperationTypeUpgrade:
		return desugarUpgrade(schemaContext, operation)
	case MigrationOperationTypeRevert:
		return desugarRevert(schemaContext, operation)
	}

	return operation, nil
}

// desugarUpgrade converts an "upgrade" operation into a targeted "upto" operation. We only need to
// identify the leaves of the current schema definition to run everything defined.
func desugarUpgrade(schemaContext schemaContext, operation MigrationOperation) (MigrationOperation, error) {
	leafVersions := extractIDs(schemaContext.schema.Definitions.Leaves())
	if len(leafVersions) != 1 {
		return MigrationOperation{}, fmt.Errorf("nothing to upgrade")
	}

	log15.Info(
		"Desugaring `upgrade` to `targeted up` operation",
		"schema", operation.SchemaName,
		"leafVersions", leafVersions,
	)

	return MigrationOperation{
		SchemaName:    operation.SchemaName,
		Type:          MigrationOperationTypeTargetedUp,
		TargetVersion: leafVersions[0],
	}, nil
}

// desugarRevert converts a "revert" operation into a targeted "down" operation. A revert operation
// is primarily meant to support "undo" capability in local development when testing a single migration
// (or linear chain of migrations).
//
// This function selects to undo the migration that has no applied children. Repeated application of the
// revert operation should "pop" off the last migration applied. This function will give up if the revert
// is ambiguous, which can happen once a migration with multiple parents has been reverted. More complex
// down migrations can be run with an explicit "down" operation.
func desugarRevert(schemaContext schemaContext, operation MigrationOperation) (MigrationOperation, error) {
	definitions := schemaContext.schema.Definitions
	leafVersions := []int{schemaContext.initialSchemaVersion.version}

	log15.Info(
		"Desugaring `revert` to `targeted down` operation",
		"schema", operation.SchemaName,
		"appliedLeafVersions", leafVersions,
	)

	// We want to revert leafVersions[0], so we need to migrate "down" its parents.
	// That operation will undo any applied proper descendants of this parent set, which
	// should consist of exactly this target version.
	definition, ok := definitions.GetByID(leafVersions[0])
	if !ok {
		return MigrationOperation{}, fmt.Errorf("unknown version %d", leafVersions[0])
	}
	if len(definition.Parents) != 1 {
		return MigrationOperation{}, fmt.Errorf("expected one parent")
	}

	return MigrationOperation{
		SchemaName:    operation.SchemaName,
		Type:          MigrationOperationTypeTargetedDown,
		TargetVersion: definition.Parents[0],
	}, nil
}
