# Clearing code intelligence data

Clear all precise code intelligence data from your instance and start fresh.

## Clearing the `frontend` database

The following commands assume a connection to the `frontend` database. Refer to our guide on [`psql`](run-psql.md) if you do not use an [external database](../external_services/index.md).

If you are clearing **all** code intelligence data, then you will need to clear the metadata information in the frontend database. This will clear the high-level information about code intelligence indexes, but the raw data will remain in the `codeintel-db` database, which will also need to be cleared (see the next section).

```sql
BEGIN;
TRUNCATE lsif_uploads;
TRUNCATE lsif_uploads_audit_logs;
TRUNCATE lsif_uploads_reference_counts;
TRUNCATE codeintel_commit_dates;
TRUNCATE lsif_packages;
TRUNCATE lsif_references;
TRUNCATE lsif_nearest_uploads;
TRUNCATE lsif_nearest_uploads_links;
TRUNCATE lsif_uploads_visible_at_tip;
TRUNCATE lsif_dirty_repositories;
COMMIT;
```

## Clearing the `codeintel-db` database

The following commands assume a connection to the `codeintel-db` database. Refer to our guide on [`psql`](run-psql.md) if you do not use an [external database](../external_services/index.md).

### Clearing LSIF data

Truncate the following tables to clear all LSIF-encoded information from the database. This command can be run when clearing **all** data from the instance, in which case the associated records in the `frontend` database must also be cleared.

This command can also be run after a completed [migration from LSIF to SCIP](lsif_scip_migration.md), in which case these tables should be empty but may retain a number of blocks containing a large number of dead tuples. Truncating these _empty_ tables acts like a `VACUUM FULL` and releases the previously used disk space.

```sql
BEGIN;
TRUNCATE lsif_data_metadata;
TRUNCATE lsif_data_documents;
TRUNCATE lsif_data_documents_schema_versions;
TRUNCATE lsif_data_result_chunks;
TRUNCATE lsif_data_definitions;
TRUNCATE lsif_data_definitions_schema_versions;
TRUNCATE lsif_data_references;
TRUNCATE lsif_data_references_schema_versions;
TRUNCATE lsif_data_implementations;
TRUNCATE lsif_data_implementations_schema_versions;
TRUNCATE codeintel_last_reconcile;
COMMIT;
```

### Clearing SCIP data

Truncate the following tables to clear all SCIP-encoded information from the database. Only run this command if you clearing **all** data from the instance, in which case the associated records in the `frontend` database must also be cleared.

```sql
BEGIN;
TRUNCATE codeintel_scip_metadata;
TRUNCATE codeintel_scip_documents;
TRUNCATE codeintel_scip_documents_schema_versions;
TRUNCATE codeintel_scip_document_lookup;
TRUNCATE codeintel_scip_document_lookup_schema_versions;
TRUNCATE codeintel_scip_documents_dereference_logs;
TRUNCATE codeintel_scip_symbols;
TRUNCATE codeintel_scip_symbols_schema_versions;
TRUNCATE codeintel_scip_symbol_names;
TRUNCATE codeintel_last_reconcile;
COMMIT;
```
