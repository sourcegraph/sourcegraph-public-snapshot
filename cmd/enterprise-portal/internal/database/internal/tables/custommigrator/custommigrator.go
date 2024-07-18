package custommigrator

import "gorm.io/gorm"

type CustomTableMigrator interface {
	// RunCustomMigrations is called after all other migrations have been run.
	// It can implement custom migrations.
	RunCustomMigrations(migrator gorm.Migrator) error
}
