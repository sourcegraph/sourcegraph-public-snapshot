package shared

import "github.com/sourcegraph/sourcegraph/internal/database/migration/definition"

// StitchedMigration represents a "virtual" migration graph constructed over time.
type StitchedMigration struct {
	// Definitions is a graph formed by concatenating and canonicalizing schema migration graphs over
	// several releases. This should contain all migrations defined in the associated version range.
	Definitions *definition.Definitions

	// BoundsByRev is a map from version to the identifiers of the root and leaf migrations defined at
	// that revision.
	BoundsByRev map[string]MigrationBounds
}

// MigrationBounds indicates version boundaries within a StitchedMigration.
type MigrationBounds struct {
	RootID      int
	LeafIDs     []int
	PreCreation bool
}

// IndexStatus describes the state of an index. Is{Valid,Ready,Live} is taken
// from the `pg_index` system table. If the index is currently being created,
// then the remaining reference fields will be populated describing the index
// creation progress.
type IndexStatus struct {
	IsValid      bool
	IsReady      bool
	IsLive       bool
	Phase        *string
	LockersDone  *int
	LockersTotal *int
	BlocksDone   *int
	BlocksTotal  *int
	TuplesDone   *int
	TuplesTotal  *int
}

// CreateIndexConcurrentlyPhases is an ordered list of phases that occur during
// a CREATE INDEX CONCURRENTLY operation. The phase of an ongoing operation can
// found in the system view `view pg_stat_progress_create_index` (since PG 12).
//
// If the phase value found in the system view may not match these values exactly
// and may only indicate a prefix. The phase may have more specific information
// following the initial phase description. Do not compare phase values exactly.
//
// See https://www.postgresql.org/docs/12/progress-reporting.html#CREATE-INDEX-PROGRESS-REPORTING.
var CreateIndexConcurrentlyPhases = []string{
	"initializing",
	"waiting for writers before build",
	"building index",
	"waiting for writers before validation",
	"index validation: scanning index",
	"index validation: sorting tuples",
	"index validation: scanning table",
	"waiting for old snapshots",
	"waiting for readers before marking dead",
	"waiting for readers before dropping",
}
