package runner

import (
	"fmt"
	"strings"
)

type SchemaOutOfDateError struct {
	schemaName      string
	currentVersion  int
	expectedVersion int
}

func (e *SchemaOutOfDateError) Error() string {
	return (instructionalError{
		class:       "schema out of date",
		description: fmt.Sprintf("expected schema %q to be at or above version %d, currently at version %d\n", e.schemaName, e.expectedVersion, e.currentVersion),
		instructions: strings.Join([]string{
			`This software expects a migrator instance to have run on this schema prior to the deployment of this process.`,
			`If this error is occurring directly after an upgrade, roll back your instance to the previous versiona nd ensure the migrator instance runs successfully prior attempting to re-upgrade.`,
		}, " "),
	}).Error()
}

type instructionalError struct {
	class        string
	description  string
	instructions string
}

func (e instructionalError) Error() string {
	return fmt.Sprintf("%s: %s\n\n%s\n", e.class, e.description, e.instructions)
}

// errDirtyDatabase occurs when a database schema is marked as dirty but there does not
// appear to be any running instance currently migrating that schema. This occurs when
// a previous attempt to migrate the schema had not successfully completed and requires
// intervention of a site administrator.
var errDirtyDatabase = instructionalError{
	class:       "dirty database",
	description: "schema is marked as dirty but no migrator instance appears to be running",
	instructions: strings.Join([]string{
		`The target schema is marked as dirty and no other migration operation is seen running on this schema.`,
		`The last migration operation over this schema has failed (or, at least, the migrator instance issuing that migration has died).`,
		`Please contact support@sourcegraph.com for further assistance.`,
	}, " "),
}

// errMigrationContention occurs when the migrator refuses to operate on a schema as there
// appears to be other migrator instances performing other concurrent operations over the
// same schema. This error only occurs when downgrading or upgrading to a particular version,
// and not on the happy-path use case of "upgrade to latest".
var errMigrationContention = instructionalError{
	class:       "migration contention",
	description: "concurrent migrator instances appear to be running on this schema",
	instructions: strings.Join([]string{
		`We have detected other migrations operations occurring on this schema and opted to abort this operation.`,
		`The state of the database is likely different than what was known at the time this command was issued.`,
		`Please check the state of your target database and re-issue this command (ensuring correct arguments).`,
	}, " "),
}
