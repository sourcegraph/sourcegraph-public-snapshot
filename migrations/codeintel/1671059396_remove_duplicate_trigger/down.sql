CREATE OR REPLACE FUNCTION update_codeintel_scip_documents_schema_versions_insert() RETURNS trigger
    LANGUAGE plpgsql
    AS $$ BEGIN
    INSERT INTO codeintel_scip_documents_schema_versions
    SELECT
        newtab.upload_id,
        MIN(codeintel_scip_documents.schema_version) as min_schema_version,
        MAX(codeintel_scip_documents.schema_version) as max_schema_version
    FROM newtab
    JOIN codeintel_scip_documents ON codeintel_scip_documents.id = newtab.document_id
    GROUP BY newtab.upload_id
    ON CONFLICT (upload_id) DO UPDATE SET
        -- Update with min(old_min, new_min) and max(old_max, new_max)
        min_schema_version = LEAST(codeintel_scip_documents_schema_versions.min_schema_version, EXCLUDED.min_schema_version),
        max_schema_version = GREATEST(codeintel_scip_documents_schema_versions.max_schema_version, EXCLUDED.max_schema_version);
    RETURN NULL;
END $$;

DROP TRIGGER IF EXISTS codeintel_scip_documents_schema_versions_insert ON codeintel_scip_document_lookup;
CREATE TRIGGER codeintel_scip_documents_schema_versions_insert AFTER INSERT ON codeintel_scip_document_lookup 
REFERENCING NEW TABLE AS newtab FOR EACH STATEMENT EXECUTE FUNCTION update_codeintel_scip_documents_schema_versions_insert();
