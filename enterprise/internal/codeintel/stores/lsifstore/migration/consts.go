package migration

// DiagnosticsCountMigrationID is the primary key of the migration record handled by an
// instance of diagnosticsCountMigrator. This is associated with the out-of-band migration
// record inserted in migrations/frontend/1528395786_diagnostic_counts_migration.up.sql.
const DiagnosticsCountMigrationID = 1

// DefinitionsCountMigrationID is the primary key of the migration record handled by an
// instance of locationsCountMigrator. This is associated with the out-of-band migration
// record inserted in migrations/frontend/1528395807_lsif_locations_migration.up.sql.
const DefinitionsCountMigrationID = 4

// ReferencesCountMigrationID is the primary key of the migration record handled by an
// instance of locationsCountMigrator. This is associated with the out-of-band migration
// record inserted in migrations/frontend/1528395807_lsif_locations_migration.up.sql.
const ReferencesCountMigrationID = 5

// DocumentColumnSplitMigrationID is the primary key of the migration record handled by an
// instance of documentColumnSplitMigrator. This explodes the data payload into several
// columns by type. This is associated with the out-of-band migration record inserted in
// migrations/frontend/1528395810_split_document_payload.up.sql.
const DocumentColumnSplitMigrationID = 7
