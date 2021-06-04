BEGIN;

CREATE EXTENSION IF NOT EXISTS pg_stat_statements;

COMMENT ON EXTENSION pg_stat_statements IS 'track execution statistics of all SQL statements executed';

CREATE TABLE lsif_data_definitions (
    dump_id integer NOT NULL,
    scheme text NOT NULL,
    identifier text NOT NULL,
    data bytea
);

COMMENT ON TABLE lsif_data_definitions IS 'Associates (document, range) pairs with the import monikers attached to the range.';

COMMENT ON COLUMN lsif_data_definitions.dump_id IS 'The identifier of the associated dump in the lsif_uploads table (state=completed).';

COMMENT ON COLUMN lsif_data_definitions.scheme IS 'The moniker scheme.';

COMMENT ON COLUMN lsif_data_definitions.identifier IS 'The moniker identifier.';

COMMENT ON COLUMN lsif_data_definitions.data IS 'A gob-encoded payload conforming to an array of [LocationData](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@3.23/-/blob/enterprise/internal/codeintel/stores/lsifstore/types.go#L100:6) types.';

CREATE TABLE lsif_data_documents (
    dump_id integer NOT NULL,
    path text NOT NULL,
    data bytea,
    schema_version integer NOT NULL,
    num_diagnostics integer NOT NULL
);

COMMENT ON TABLE lsif_data_documents IS 'Stores reference, hover text, moniker, and diagnostic data about a particular text document witin a dump.';

COMMENT ON COLUMN lsif_data_documents.dump_id IS 'The identifier of the associated dump in the lsif_uploads table (state=completed).';

COMMENT ON COLUMN lsif_data_documents.path IS 'The path of the text document relative to the associated dump root.';

COMMENT ON COLUMN lsif_data_documents.data IS 'A gob-encoded payload conforming to the [DocumentData](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@3.23/-/blob/enterprise/internal/codeintel/stores/lsifstore/types.go#L13:6) type.';

COMMENT ON COLUMN lsif_data_documents.schema_version IS 'The schema version of this row - used to determine presence and encoding of data.';

COMMENT ON COLUMN lsif_data_documents.num_diagnostics IS 'The number of diagnostics stored in the data field.';

CREATE TABLE lsif_data_metadata (
    dump_id integer NOT NULL,
    num_result_chunks integer
);

COMMENT ON TABLE lsif_data_metadata IS 'Stores the number of result chunks associated with a dump.';

COMMENT ON COLUMN lsif_data_metadata.dump_id IS 'The identifier of the associated dump in the lsif_uploads table (state=completed).';

COMMENT ON COLUMN lsif_data_metadata.num_result_chunks IS 'A bound of populated indexes in the lsif_data_result_chunks table for the associated dump. This value is used to hash identifiers into the result chunk index to which they belong.';

CREATE TABLE lsif_data_references (
    dump_id integer NOT NULL,
    scheme text NOT NULL,
    identifier text NOT NULL,
    data bytea
);

COMMENT ON TABLE lsif_data_references IS 'Associates (document, range) pairs with the export monikers attached to the range.';

COMMENT ON COLUMN lsif_data_references.dump_id IS 'The identifier of the associated dump in the lsif_uploads table (state=completed).';

COMMENT ON COLUMN lsif_data_references.scheme IS 'The moniker scheme.';

COMMENT ON COLUMN lsif_data_references.identifier IS 'The moniker identifier.';

COMMENT ON COLUMN lsif_data_references.data IS 'A gob-encoded payload conforming to an array of [LocationData](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@3.23/-/blob/enterprise/internal/codeintel/stores/lsifstore/types.go#L100:6) types.';

CREATE TABLE lsif_data_result_chunks (
    dump_id integer NOT NULL,
    idx integer NOT NULL,
    data bytea
);

COMMENT ON TABLE lsif_data_result_chunks IS 'Associates result set identifiers with the (document path, range identifier) pairs that compose the set.';

COMMENT ON COLUMN lsif_data_result_chunks.dump_id IS 'The identifier of the associated dump in the lsif_uploads table (state=completed).';

COMMENT ON COLUMN lsif_data_result_chunks.idx IS 'The unique result chunk index within the associated dump. Every result set identifier present should hash to this index (modulo lsif_data_metadata.num_result_chunks).';

COMMENT ON COLUMN lsif_data_result_chunks.data IS 'A gob-encoded payload conforming to the [ResultChunkData](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@3.23/-/blob/enterprise/internal/codeintel/stores/lsifstore/types.go#L70:6) type.';

ALTER TABLE ONLY lsif_data_definitions
    ADD CONSTRAINT lsif_data_definitions_pkey PRIMARY KEY (dump_id, scheme, identifier);

ALTER TABLE ONLY lsif_data_documents
    ADD CONSTRAINT lsif_data_documents_pkey PRIMARY KEY (dump_id, path);

ALTER TABLE ONLY lsif_data_metadata
    ADD CONSTRAINT lsif_data_metadata_pkey PRIMARY KEY (dump_id);

ALTER TABLE ONLY lsif_data_references
    ADD CONSTRAINT lsif_data_references_pkey PRIMARY KEY (dump_id, scheme, identifier);

ALTER TABLE ONLY lsif_data_result_chunks
    ADD CONSTRAINT lsif_data_result_chunks_pkey PRIMARY KEY (dump_id, idx);

COMMIT;
