package runner

import (
	"strings"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type Options struct {
	Operations []MigrationOperation

	// Parallel controls whether we run schema migrations concurrently or not. By default,
	// we run schema migrations sequentially. This is to ensure that in testing, where the
	// same database can be targeted by multiple schemas, we do not hit errors that occur
	// when trying to install Postgres extensions concurrently (which do not seem txn-safe).
	Parallel bool

	// PrivilegedMode controls how privileged migrations are applied.
	PrivilegedMode PrivilegedMode

	// MatchPrivilegedHash is a function that matches a string indicating a deterministic hash
	// of the set of privileged migrations that should be no-op'd against user-supplied strings
	// given from a previous run with the same migration state. This value is only checked when
	// running up-direction migrations with a privileged mode of `NoopPrivilegedMigrations`.
	MatchPrivilegedHash func(hash string) bool

	// IgnoreSingleDirtyLog controls whether or not to ignore a dirty database in the specific
	// case when the _next_ migration application is the only failure. This is meant to enable
	// a short development loop where the user can re-apply the `up` command without having to
	// create a dummy migration log to proceed.
	IgnoreSingleDirtyLog bool

	// IgnoreSinglePendingLog controls whether or not to ignore a pending migration log in the
	// specific case when the _next_ migration application is the only pending migration. This
	// is meant to enable interruptable upgrades.
	IgnoreSinglePendingLog bool
}

type PrivilegedMode uint

func (m PrivilegedMode) Valid() bool {
	return m < InvalidPrivilegedMode
}

const (
	// ApplyPrivilegedMigrations, the default privileged mode, indicates to the runner that any
	// privileged migrations should be applied along with unprivileged migrations.
	ApplyPrivilegedMigrations PrivilegedMode = iota

	// NoopPrivilegedMigrations, enabled via the -noop-privileged flag, indicates to the runner
	// that any privileged migrations should be skipped, but an entry in the migration logs table
	// should be added. This mode assumes that the user has already applied these migrations by hand.
	NoopPrivilegedMigrations

	// RefusePrivilegedMigrations, enabled via the -unprivileged-only flag, indicates to the runner
	// that any privileged migrations should result in an error. This indicates to the user that
	// these migrations need to be run by hand with elevated permissions before the migration can
	// succeed.
	RefusePrivilegedMigrations

	// InvalidPrivilegedMode indicates an unsupported privileged mode state.
	InvalidPrivilegedMode
)

type MigrationOperation struct {
	SchemaName     string
	Type           MigrationOperationType
	TargetVersions []int
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
		return desugarUpgrade(schemaContext, operation), nil
	case MigrationOperationTypeRevert:
		return desugarRevert(schemaContext, operation)
	}

	return operation, nil
}

// desugarUpgrade converts an "upgrade" operation into a targeted up operation. We only need to
// identify the leaves of the current schema definition to run everything defined.
func desugarUpgrade(schemaContext schemaContext, operation MigrationOperation) MigrationOperation {
	leafVersions := extractIDs(schemaContext.schema.Definitions.Leaves())

	schemaContext.logger.Info(
		"Desugaring `upgrade` to `targeted up` operation",
		log.String("schema", operation.SchemaName),
		log.Ints("leafVersions", leafVersions),
	)

	return MigrationOperation{
		SchemaName:     operation.SchemaName,
		Type:           MigrationOperationTypeTargetedUp,
		TargetVersions: leafVersions,
	}
}

// desugarRevert converts a "revert" operation into a targeted down operation. A revert operation
// is primarily meant to support "undo" capability in local development when testing a single migration
// (or linear chain of migrations).
//
// This function selects to undo the migration that has no applied children. Repeated application of the
// revert operation should "pop" off the last migration applied. This function will give up if the revert
// is ambiguous, which can happen once a migration with multiple parents has been reverted. More complex
// down migrations can be run with an explicit targeted down operation.
func desugarRevert(schemaContext schemaContext, operation MigrationOperation) (MigrationOperation, error) {
	definitions := schemaContext.schema.Definitions
	schemaVersion := schemaContext.initialSchemaVersion

	// Construct a map from migration version to the number of its children that are also applied
	counts := make(map[int]int, len(schemaVersion.appliedVersions))
	for _, version := range schemaVersion.appliedVersions {
		definition, ok := definitions.GetByID(version)
		if !ok {
			continue
		}

		for _, parent := range definition.Parents {
			counts[parent] = counts[parent] + 1
		}

		// Ensure that we have an entry for this definition (but do not modify the count)
		counts[definition.ID] = counts[definition.ID] + 0
	}

	// Find applied migrations with no applied children
	leafVersions := make([]int, 0, len(counts))
	for version, numChildren := range counts {
		if numChildren == 0 {
			leafVersions = append(leafVersions, version)
		}
	}

	schemaContext.logger.Info(
		"Desugaring `revert` to `targeted down` operation",
		log.String("schema", operation.SchemaName),
		log.Ints("appliedLeafVersions", leafVersions),
	)

	switch len(leafVersions) {
	case 1:
		// We want to revert leafVersions[0], so we need to migrate down to its parents.
		// That operation will undo any applied proper descendants of this parent set, which
		// should consist of exactly this target version.
		definition, ok := definitions.GetByID(leafVersions[0])
		if !ok {
			return MigrationOperation{}, errors.Newf("unknown version %d", leafVersions[0])
		}

		return MigrationOperation{
			SchemaName:     operation.SchemaName,
			Type:           MigrationOperationTypeTargetedDown,
			TargetVersions: definition.Parents,
		}, nil

	case 0:
		return MigrationOperation{}, errors.Newf("nothing to revert")

	default:
		return MigrationOperation{}, errors.Newf("ambiguous revert - candidates include %s", strings.Join(intsToStrings(leafVersions), ", "))
	}
}
