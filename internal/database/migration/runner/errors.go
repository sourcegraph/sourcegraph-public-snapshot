package runner

import "fmt"

type SchemaOutOfDateError struct {
	schemaName      string
	currentVersion  int
	expectedVersion int
}

func (e *SchemaOutOfDateError) Error() string {
	return fmt.Sprintf("expected schema %q to be migrated to version %d, currently %d\n", e.schemaName, e.expectedVersion, e.currentVersion)
}
