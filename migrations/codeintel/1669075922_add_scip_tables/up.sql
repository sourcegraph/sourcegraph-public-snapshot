CREATE TABLE IF NOT EXISTS codeintel_scip_metadata (
    id bigserial,
    upload_id integer NOT NULL,
    tool_name text NOT NULL,
    tool_version text NOT NULL,
    tool_arguments text[] NOT NULL,
    PRIMARY KEY(id)
);

COMMENT ON TABLE codeintel_scip_metadata IS 'Global metadatadata about a single processed upload.';
COMMENT ON COLUMN codeintel_scip_metadata.id IS 'An auto-generated identifier.';
COMMENT ON COLUMN codeintel_scip_metadata.upload_id IS 'The identifier of the upload that provided this SCIP index.';
COMMENT ON COLUMN codeintel_scip_metadata.tool_name IS 'Name of the indexer that produced this index.';
COMMENT ON COLUMN codeintel_scip_metadata.tool_version IS 'Version of the indexer that produced this index.';
COMMENT ON COLUMN codeintel_scip_metadata.tool_arguments IS 'Command-line arguments that were used to invoke this indexer.';

CREATE TABLE IF NOT EXISTS codeintel_scip_documents(
    id bigserial,
    payload_hash bytea NOT NULL UNIQUE,
    schema_version integer NOT NULL,
    raw_scip_payload bytea NOT NULL,
    PRIMARY KEY(id)
);

COMMENT ON TABLE codeintel_scip_documents IS 'A lookup of SCIP [Document](https://sourcegraph.com/search?q=context:%40sourcegraph/all+repo:%5Egithub%5C.com/sourcegraph/scip%24+file:%5Escip%5C.proto+message+Document&patternType=standard) payloads by their hash.';
COMMENT ON COLUMN codeintel_scip_documents.id IS 'An auto-generated identifier. This column is used as a foreign key target to reduce occurrences of the full payload hash value.';
COMMENT ON COLUMN codeintel_scip_documents.payload_hash IS 'A deterministic hash of the raw SCIP payload. We use this as a unique value to enforce deduplication between indexes with the same document data.';
COMMENT ON COLUMN codeintel_scip_documents.schema_version IS 'The schema version of this row - used to determine presence and encoding of (future) denormalized data.';
COMMENT ON COLUMN codeintel_scip_documents.raw_scip_payload IS 'The raw, canonicalized SCIP [Document](https://sourcegraph.com/search?q=context:%40sourcegraph/all+repo:%5Egithub%5C.com/sourcegraph/scip%24+file:%5Escip%5C.proto+message+Document&patternType=standard) payload.';

CREATE TABLE IF NOT EXISTS codeintel_scip_document_lookup(
    id bigserial,
    upload_id integer NOT NULL,
    document_path text NOT NULL,
    document_id bigint NOT NULL,
    PRIMARY KEY (id),
    UNIQUE (upload_id, document_path),
    CONSTRAINT codeintel_scip_document_lookup_document_id_fk FOREIGN KEY(document_id) REFERENCES codeintel_scip_documents(id)
);

CREATE INDEX IF NOT EXISTS codeintel_scip_document_lookup_document_id ON codeintel_scip_document_lookup USING hash(document_id);

COMMENT ON TABLE codeintel_scip_document_lookup IS 'A mapping from file paths to document references within a particular SCIP index.';
COMMENT ON COLUMN codeintel_scip_document_lookup.id IS 'An auto-generated identifier. This column is used as a foreign key target to reduce occurrences of the full document path value.';
COMMENT ON COLUMN codeintel_scip_document_lookup.upload_id IS 'The identifier of the upload that provided this SCIP index.';
COMMENT ON COLUMN codeintel_scip_document_lookup.document_path IS 'The file path to the document relative to the root of the index.';
COMMENT ON COLUMN codeintel_scip_document_lookup.document_id IS 'The foreign key to the shared document payload (see the table [`codeintel_scip_document_lookup`](#table-publiccodeintel_scip_document_lookup)).';

CREATE TABLE IF NOT EXISTS codeintel_scip_document_lookup_schema_versions (
    upload_id integer NOT NULL,
    min_schema_version integer,
    max_schema_version integer,
    PRIMARY KEY(upload_id)
);

COMMENT ON TABLE codeintel_scip_document_lookup_schema_versions IS 'Tracks the range of `schema_versions` values associated with each SCIP index in the [`codeintel_scip_document_lookup`](#table-publiccodeintel_scip_document_lookup) table.';
COMMENT ON COLUMN codeintel_scip_document_lookup_schema_versions.upload_id IS 'The identifier of the associated SCIP index.';
COMMENT ON COLUMN codeintel_scip_document_lookup_schema_versions.min_schema_version IS 'A lower-bound on the `schema_version` values of the records in the table [`codeintel_scip_document_lookup`](#table-publiccodeintel_scip_document_lookup) where the `upload_id` column matches the associated SCIP index.';
COMMENT ON COLUMN codeintel_scip_document_lookup_schema_versions.max_schema_version IS 'An upper-bound on the `schema_version` values of the records in the table [`codeintel_scip_document_lookup`](#table-publiccodeintel_scip_document_lookup) where the `upload_id` column matches the associated SCIP index.';

CREATE OR REPLACE FUNCTION update_codeintel_scip_document_lookup_schema_versions_insert() RETURNS trigger
    LANGUAGE plpgsql
    AS $$ BEGIN
    INSERT INTO codeintel_scip_document_lookup_schema_versions
    SELECT
        upload_id,
        MIN(documents.schema_version) as min_schema_version,
        MAX(documents.schema_version) as max_schema_version
    FROM newtab
    JOIN codeintel_scip_documents ON codeintel_scip_documents.id = newtab.document_id
    GROUP BY newtab.upload_id
    ON CONFLICT (upload_id) DO UPDATE SET
        -- Update with min(old_min, new_min) and max(old_max, new_max)
        min_schema_version = LEAST(codeintel_scip_document_lookup_schema_versions.min_schema_version, EXCLUDED.min_schema_version),
        max_schema_version = GREATEST(codeintel_scip_document_lookup_schema_versions.max_schema_version, EXCLUDED.max_schema_version);
    RETURN NULL;
END $$;

DROP TRIGGER IF EXISTS codeintel_scip_document_lookup_schema_versions_insert ON codeintel_scip_document_lookup_schema_versions;
CREATE TRIGGER codeintel_scip_document_lookup_schema_versions_insert AFTER INSERT ON codeintel_scip_document_lookup_schema_versions
REFERENCING NEW TABLE AS newtab FOR EACH STATEMENT EXECUTE FUNCTION update_codeintel_scip_document_lookup_schema_versions_insert();

CREATE TABLE IF NOT EXISTS codeintel_scip_symbols(
    upload_id integer NOT NULL,
    symbol_name text NOT NULL,
    document_lookup_id bigint NOT NULL,
    schema_version integer NOT NULL,
    definition_ranges bytea,
    reference_ranges bytea,
    implementation_ranges bytea,
    type_definition_ranges bytea,
    PRIMARY KEY (upload_id, symbol_name, document_lookup_id),
    CONSTRAINT codeintel_scip_symbols_document_lookup_id_fk FOREIGN KEY(document_lookup_id) REFERENCES codeintel_scip_document_lookup(id) ON DELETE CASCADE
);

COMMENT ON TABLE codeintel_scip_symbols IS 'A mapping from SCIP [Symbol names](https://sourcegraph.com/search?q=context:%40sourcegraph/all+repo:%5Egithub%5C.com/sourcegraph/scip%24+file:%5Escip%5C.proto+message+Symbol&patternType=standard) to path and ranges where that symbol occurs within a particular SCIP index.';
COMMENT ON COLUMN codeintel_scip_symbols.upload_id IS 'The identifier of the upload that provided this SCIP index.';
COMMENT ON COLUMN codeintel_scip_symbols.symbol_name IS 'The SCIP [Symbol names](https://sourcegraph.com/search?q=context:%40sourcegraph/all+repo:%5Egithub%5C.com/sourcegraph/scip%24+file:%5Escip%5C.proto+message+Symbol&patternType=standard).';
COMMENT ON COLUMN codeintel_scip_symbols.document_lookup_id IS 'A reference to the `id` column of [`codeintel_scip_document_lookup`](#table-publiccodeintel_scip_document_lookup). Joining on this table yields the document path relative to the index root.';
COMMENT ON COLUMN codeintel_scip_symbols.schema_version IS 'The schema version of this row - used to determine presence and encoding of denormalized data.';
COMMENT ON COLUMN codeintel_scip_symbols.definition_ranges IS 'An encoded set of ranges within the associated document that have a **definition** relationship to the associated symbol.';
COMMENT ON COLUMN codeintel_scip_symbols.reference_ranges IS 'An encoded set of ranges within the associated document that have a **reference** relationship to the associated symbol.';
COMMENT ON COLUMN codeintel_scip_symbols.implementation_ranges IS 'An encoded set of ranges within the associated document that have a **implementation** relationship to the associated symbol.';
COMMENT ON COLUMN codeintel_scip_symbols.type_definition_ranges IS 'An encoded set of ranges within the associated document that have a **type definition** relationship to the associated symbol.';

CREATE TABLE IF NOT EXISTS codeintel_scip_symbols_schema_versions (
    upload_id integer NOT NULL,
    min_schema_version integer,
    max_schema_version integer,
    PRIMARY KEY(upload_id)
);

COMMENT ON TABLE codeintel_scip_symbols_schema_versions IS 'Tracks the range of `schema_versions` for each index in the [`codeintel_scip_symbols`](#table-publiccodeintel_scip_symbols) table.';
COMMENT ON COLUMN codeintel_scip_symbols_schema_versions.upload_id IS 'The identifier of the associated SCIP index.';
COMMENT ON COLUMN codeintel_scip_symbols_schema_versions.min_schema_version IS 'A lower-bound on the `schema_version` values of the records in the table [`codeintel_scip_symbols`](#table-publiccodeintel_scip_symbols) where the `upload_id` column matches the associated SCIP index.';
COMMENT ON COLUMN codeintel_scip_symbols_schema_versions.max_schema_version IS 'An upper-bound on the `schema_version` values of the records in the table [`codeintel_scip_symbols`](#table-publiccodeintel_scip_symbols) where the `upload_id` column matches the associated SCIP index.';

CREATE OR REPLACE FUNCTION update_codeintel_scip_symbols_schema_versions_insert() RETURNS trigger
    LANGUAGE plpgsql
    AS $$ BEGIN
    INSERT INTO codeintel_scip_symbols_schema_versions
    SELECT
        upload_id,
        MIN(schema_version) as min_schema_version,
        MAX(schema_version) as max_schema_version
    FROM newtab
    GROUP BY upload_id
    ON CONFLICT (upload_id) DO UPDATE SET
        -- Update with min(old_min, new_min) and max(old_max, new_max)
        min_schema_version = LEAST(codeintel_scip_symbols_schema_versions.min_schema_version, EXCLUDED.min_schema_version),
        max_schema_version = GREATEST(codeintel_scip_symbols_schema_versions.max_schema_version, EXCLUDED.max_schema_version);
    RETURN NULL;
END $$;

DROP TRIGGER IF EXISTS codeintel_scip_symbols_schema_versions_insert ON codeintel_scip_symbols_schema_versions;
CREATE TRIGGER codeintel_scip_symbols_schema_versions_insert AFTER INSERT ON codeintel_scip_symbols_schema_versions
REFERENCING NEW TABLE AS newtab FOR EACH STATEMENT EXECUTE FUNCTION update_codeintel_scip_symbols_schema_versions_insert();
