-- Restore table and triggers
DROP TABLE codeintel_scip_documents_schema_versions;
CREATE TABLE IF NOT EXISTS codeintel_scip_documents_schema_versions (
    upload_id integer NOT NULL,
    min_schema_version integer,
    max_schema_version integer,
    PRIMARY KEY(upload_id)
);

COMMENT ON TABLE codeintel_scip_documents_schema_versions IS 'Tracks the range of `schema_versions` values associated with each SCIP index in the [`codeintel_scip_documents`](#table-publiccodeintel_scip_documents) table.';
COMMENT ON COLUMN codeintel_scip_documents_schema_versions.upload_id IS 'The identifier of the associated SCIP index.';
COMMENT ON COLUMN codeintel_scip_documents_schema_versions.min_schema_version IS 'A lower-bound on the `schema_version` values of the records in the table [`codeintel_scip_documents`](#table-publiccodeintel_scip_documents) where the `upload_id` column matches the associated SCIP index.';
COMMENT ON COLUMN codeintel_scip_documents_schema_versions.max_schema_version IS 'An upper-bound on the `schema_version` values of the records in the table [`codeintel_scip_documents`](#table-publiccodeintel_scip_documents) where the `upload_id` column matches the associated SCIP index.';

CREATE OR REPLACE FUNCTION update_codeintel_scip_documents_schema_versions_insert() RETURNS trigger
    LANGUAGE plpgsql
    AS $$ BEGIN
    INSERT INTO codeintel_scip_documents_schema_versions
    SELECT
        upload_id,
        MIN(documents.schema_version) as min_schema_version,
        MAX(documents.schema_version) as max_schema_version
    FROM newtab
    JOIN codeintel_scip_documents ON codeintel_scip_documents.id = newtab.document_id
    GROUP BY newtab.upload_id
    ON CONFLICT (upload_id) DO UPDATE SET
        -- Update with min(old_min, new_min) and max(old_max, new_max)
        min_schema_version = LEAST(codeintel_scip_documents_schema_versions.min_schema_version, EXCLUDED.min_schema_version),
        max_schema_version = GREATEST(codeintel_scip_documents_schema_versions.max_schema_version, EXCLUDED.max_schema_version);
    RETURN NULL;
END $$;

DROP TRIGGER IF EXISTS codeintel_scip_documents_schema_versions_insert ON codeintel_scip_documents;
DROP TRIGGER IF EXISTS codeintel_scip_documents_schema_versions_insert ON codeintel_scip_documents_schema_versions;

-- Restore documents table
ALTER TABLE codeintel_scip_documents DROP COLUMN IF EXISTS metadata_shard_id;
