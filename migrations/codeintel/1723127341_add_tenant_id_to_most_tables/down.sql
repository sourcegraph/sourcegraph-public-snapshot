-- Temporary function to deduplicate the logic required for each table:
CREATE OR REPLACE FUNCTION migrate_add_tenant_id_codeintel(table_name text)
RETURNS void AS $$
BEGIN
    EXECUTE format('ALTER TABLE %I DROP COLUMN IF EXISTS tenant_id;', table_name);
END;
$$ LANGUAGE plpgsql;

SELECT migrate_add_tenant_id_codeintel('codeintel_last_reconcile');
SELECT migrate_add_tenant_id_codeintel('codeintel_scip_document_lookup');
SELECT migrate_add_tenant_id_codeintel('codeintel_scip_document_lookup_schema_versions');
SELECT migrate_add_tenant_id_codeintel('codeintel_scip_documents');
SELECT migrate_add_tenant_id_codeintel('codeintel_scip_documents_dereference_logs');
SELECT migrate_add_tenant_id_codeintel('codeintel_scip_metadata');
SELECT migrate_add_tenant_id_codeintel('codeintel_scip_symbol_names');
SELECT migrate_add_tenant_id_codeintel('codeintel_scip_symbols');
SELECT migrate_add_tenant_id_codeintel('codeintel_scip_symbols_schema_versions');
SELECT migrate_add_tenant_id_codeintel('rockskip_ancestry');
SELECT migrate_add_tenant_id_codeintel('rockskip_repos');
SELECT migrate_add_tenant_id_codeintel('rockskip_symbols');

DROP FUNCTION migrate_add_tenant_id_codeintel(text);
