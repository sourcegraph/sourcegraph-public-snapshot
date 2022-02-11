package migration

// CommittedAtMigrationID is the primary key of the migration record handled by an instance of
// committedAtMigrator. This is associated with the out-of-band migration record inserted in the
// file migrations/frontend/1528395817_lsif_uploads_committed_at.up.sql.
const CommittedAtMigrationID = 8

// ReferenceCountMigrationID is the primary key of the migration record handled by an instance of
// referenceCountMigrator. This is associated with the out-of-band migration record inserted in the
// file migrations/frontend/1528395873_lsif_upload_reference_counts_oob_migration.up.sql.
const ReferenceCountMigrationID = 11
