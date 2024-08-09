-- Temporary function to deduplicate the logic required for each table:
CREATE OR REPLACE FUNCTION migrate_table(table_name text)
RETURNS void AS $$
BEGIN
    EXECUTE format('ALTER TABLE %I DROP COLUMN IF EXISTS tenant_id;', table_name);
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

DROP FUNCTION migrate_table(text);
