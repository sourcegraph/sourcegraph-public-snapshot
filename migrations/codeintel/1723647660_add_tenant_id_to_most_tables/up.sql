-- This migration adds the tenant_id column in a way which doesn't require
-- updating every row. The value is null and an out of band migration will set
-- it to the default. A later migration will enforce tenant_id to be set.
--
-- We COMMIT AND CHAIN after each table is altered to prevent a single
-- transaction over all the alters. A single transaction would lead to a
-- deadlock with concurrent application queries.

-- Temporary function to deduplicate the logic required for each table:
CREATE OR REPLACE FUNCTION migrate_add_tenant_id_codeintel(table_name text)
RETURNS void AS $$
BEGIN
    -- ALTER TABLE with a foreign key constraint will _always_ add the
    -- constraint, which means we always require a table lock even if this
    -- migration has run. So we check if the column exists first.
    IF NOT EXISTS (SELECT true
        FROM   pg_attribute
        WHERE  attrelid = quote_ident(table_name)::regclass
        AND    attname = 'tenant_id'
        AND    NOT attisdropped
    ) THEN
        EXECUTE format('ALTER TABLE %I ADD COLUMN IF NOT EXISTS tenant_id integer REFERENCES tenants(id) ON UPDATE CASCADE ON DELETE CASCADE;', table_name);
    END IF;
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

-- Explicitly excluded tables
-- migration_logs :: about DB

DROP FUNCTION migrate_add_tenant_id_codeintel(text);
