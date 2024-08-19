-- Temporary function to deduplicate the logic required for each table:
CREATE OR REPLACE FUNCTION migrate_add_tenant_id_codeintel(table_name text)
RETURNS void AS $$
BEGIN
    EXECUTE format('ALTER TABLE %I DROP COLUMN IF EXISTS tenant_id;', table_name);
END;
$$ LANGUAGE plpgsql;

SELECT migrate_add_tenant_id_codeintel('codeintel_last_reconcile'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_codeintel('codeintel_scip_document_lookup'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_codeintel('codeintel_scip_document_lookup_schema_versions'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_codeintel('codeintel_scip_documents'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_codeintel('codeintel_scip_documents_dereference_logs'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_codeintel('codeintel_scip_metadata'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_codeintel('codeintel_scip_symbol_names'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_codeintel('codeintel_scip_symbols'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_codeintel('codeintel_scip_symbols_schema_versions'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_codeintel('rockskip_ancestry'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_codeintel('rockskip_repos'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_codeintel('rockskip_symbols'); COMMIT AND CHAIN;

DROP FUNCTION migrate_add_tenant_id_codeintel(text);
