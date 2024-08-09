-- This migration adds the tenant_id column in a way which doesn't require
-- updating every row. The value is null and an out of band migration will set
-- it to the default. A later migration will enforce tenant_id to be set.

-- Temporary function to deduplicate the logic required for each table:
CREATE OR REPLACE FUNCTION migrate_table(table_name text)
RETURNS void AS $$
BEGIN
    EXECUTE format('ALTER TABLE %I ADD COLUMN IF NOT EXISTS tenant_id integer REFERENCES tenants(id) ON UPDATE CASCADE ON DELETE CASCADE;', table_name);
END;
$$ LANGUAGE plpgsql;

SELECT migrate_table('codeintel_last_reconcile');
SELECT migrate_table('codeintel_scip_document_lookup');
SELECT migrate_table('codeintel_scip_document_lookup_schema_versions');
SELECT migrate_table('codeintel_scip_documents');
SELECT migrate_table('codeintel_scip_documents_dereference_logs');
SELECT migrate_table('codeintel_scip_metadata');
SELECT migrate_table('codeintel_scip_symbol_names');
SELECT migrate_table('codeintel_scip_symbols');
SELECT migrate_table('codeintel_scip_symbols_schema_versions');
SELECT migrate_table('rockskip_ancestry');
SELECT migrate_table('rockskip_repos');
SELECT migrate_table('rockskip_symbols');

-- Explicitly excluded tables
-- migration_logs :: about DB

DROP FUNCTION migrate_table(text);
