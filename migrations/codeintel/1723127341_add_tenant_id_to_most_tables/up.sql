-- This migration adds the tenant_id column in a way which doesn't require
-- updating every row. The value is null and an out of band migration will set
-- it to the default. A later migration will enforce tenant_id to be set.
--
-- We wrap each call with its own transaction. This is since mutliple
-- statements in a single call to postgres is run in its own transaction by
-- default. We want to run each alter without joining the table locks with
-- eachother to avoid deadlock.

-- Temporary function to deduplicate the logic required for each table:
CREATE OR REPLACE FUNCTION migrate_add_tenant_id_codeintel(table_name text)
RETURNS void AS $$
BEGIN
    EXECUTE format('ALTER TABLE %I ADD COLUMN IF NOT EXISTS tenant_id integer REFERENCES tenants(id) ON UPDATE CASCADE ON DELETE CASCADE;', table_name);
END;
$$ LANGUAGE plpgsql;

BEGIN; SELECT migrate_add_tenant_id_codeintel('codeintel_last_reconcile'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_codeintel('codeintel_scip_document_lookup'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_codeintel('codeintel_scip_document_lookup_schema_versions'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_codeintel('codeintel_scip_documents'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_codeintel('codeintel_scip_documents_dereference_logs'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_codeintel('codeintel_scip_metadata'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_codeintel('codeintel_scip_symbol_names'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_codeintel('codeintel_scip_symbols'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_codeintel('codeintel_scip_symbols_schema_versions'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_codeintel('rockskip_ancestry'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_codeintel('rockskip_repos'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_codeintel('rockskip_symbols'); COMMIT;

-- Explicitly excluded tables
-- migration_logs :: about DB

DROP FUNCTION migrate_add_tenant_id_codeintel(text);
