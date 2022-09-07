package oobmigration

import "context"

// Migrator handles migrating data from one format into another in a way that cannot easily
// be done via the in-band migration mechanism. This may be due to a large amount of data, or
// a process that requires the results of an external API or non-SQL-compatible encoding
// (e.g., gob-encode or gzipped payloads).
type Migrator interface {
	// Progress returns a percentage (in the range range [0, 1]) of data records that need
	// to be migrated in the up direction. A value of 0 means that no data has been changedk.
	// A value of 1 means that the underlying data has been completely migrated. A value < 1
	// denotes that a future invocation of the Up method may affect additional data, excluding
	// error conditions and prerequisite migrations. A value > 0 denotes that a future invocation
	// of the Down method may affect additional data.
	Progress(ctx context.Context, applyReverse bool) (float64, error)

	// Up runs a batch of the migration. This method is called repeatedly until the Progress
	// method reports completion. Errors returned from this method will be associated with the
	// migration record.
	Up(ctx context.Context) error

	// Down runs a batch of the migration in reverse. This does not need to be implemented
	// for migrations which are non-destructive. A non-destructive migration only adds data,
	// and does not transform fields that were read by previous versions of Sourcegraph and
	// therefore do not need to be undone prior to a downgrade.
	Down(ctx context.Context) error
}
