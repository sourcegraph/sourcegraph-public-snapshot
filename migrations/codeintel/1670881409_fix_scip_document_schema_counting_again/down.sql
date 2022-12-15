-- Add shard id to documents
ALTER TABLE codeintel_scip_documents ADD COLUMN IF NOT EXISTS metadata_shard_id integer NOT NULL DEFAULT floor(random() * 128 + 1)::integer;
COMMENT ON COLUMN codeintel_scip_documents.metadata_shard_id IS 'A randomly generated integer used to arbitrarily bucket groups of documents for things like expiration checks and data migrations.';

-- Replace table and triggers
DROP TABLE IF EXISTS codeintel_scip_documents_schema_versions;
CREATE TABLE codeintel_scip_documents_schema_versions (
    metadata_shard_id integer NOT NULL,
    min_schema_version integer,
    max_schema_version integer,
    PRIMARY KEY(metadata_shard_id)
);

COMMENT ON TABLE codeintel_scip_documents_schema_versions IS 'Tracks the range of `schema_versions` values associated with each document metadata shard in the [`codeintel_scip_documents`](#table-publiccodeintel_scip_documents) table.';
COMMENT ON COLUMN codeintel_scip_documents_schema_versions.metadata_shard_id IS 'The identifier of the associated document metadata shard.';
COMMENT ON COLUMN codeintel_scip_documents_schema_versions.min_schema_version IS 'A lower-bound on the `schema_version` values of the records in the table [`codeintel_scip_documents`](#table-publiccodeintel_scip_documents) where the `metadata_shard_id` column matches the associated document metadata shard.';
COMMENT ON COLUMN codeintel_scip_documents_schema_versions.max_schema_version IS 'An upper-bound on the `schema_version` values of the records in the table [`codeintel_scip_documents`](#table-publiccodeintel_scip_documents) where the `metadata_shard_id` column matches the associated document metadata shard.';

CREATE OR REPLACE FUNCTION update_codeintel_scip_documents_schema_versions_insert() RETURNS trigger
    LANGUAGE plpgsql
    AS $$ BEGIN
    INSERT INTO codeintel_scip_documents_schema_versions
    SELECT
        newtab.metadata_shard_id,
        MIN(codeintel_scip_documents.schema_version) as min_schema_version,
        MAX(codeintel_scip_documents.schema_version) as max_schema_version
    FROM newtab
    JOIN codeintel_scip_documents ON codeintel_scip_documents.metadata_shard_id = newtab.metadata_shard_id
    GROUP BY newtab.metadata_shard_id
    ON CONFLICT (metadata_shard_id) DO UPDATE SET
        -- Update with min(old_min, new_min) and max(old_max, new_max)
        min_schema_version = LEAST(codeintel_scip_documents_schema_versions.min_schema_version, EXCLUDED.min_schema_version),
        max_schema_version = GREATEST(codeintel_scip_documents_schema_versions.max_schema_version, EXCLUDED.max_schema_version);
    RETURN NULL;
END $$;

DROP TRIGGER IF EXISTS codeintel_scip_documents_schema_versions_insert ON codeintel_scip_documents;
DROP TRIGGER IF EXISTS codeintel_scip_documents_schema_versions_insert ON codeintel_scip_document_lookup;
DROP TRIGGER IF EXISTS codeintel_scip_documents_schema_versions_insert ON codeintel_scip_documents_schema_versions;
CREATE TRIGGER codeintel_scip_documents_schema_versions_insert AFTER INSERT ON codeintel_scip_documents
REFERENCING NEW TABLE AS newtab FOR EACH STATEMENT EXECUTE FUNCTION update_codeintel_scip_documents_schema_versions_insert();
