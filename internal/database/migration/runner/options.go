package runner

type Options struct {
	Operations []MigrationOperation

	// Parallel controls whether we run schema migrations concurrently or not. By default,
	// we run schema migrations sequentially. This is to ensure that in testing, where the
	// same database can be targetted by multiple schemas, we do not hit errors that occur
	// when trying to install Postgres extensions concurrently (which do not seem txn-safe).
	Parallel bool
}

type MigrationOperation struct {
	SchemaName      string
	Up              bool
	TargetMigration int
}
