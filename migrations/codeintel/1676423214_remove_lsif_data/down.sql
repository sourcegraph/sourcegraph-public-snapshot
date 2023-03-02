CREATE TABLE IF NOT EXISTS lsif_data_definitions (
    dump_id integer NOT NULL,
    scheme text NOT NULL,
    identifier text NOT NULL,
    data bytea,
    schema_version integer NOT NULL,
    num_locations integer NOT NULL,
    PRIMARY KEY (dump_id, scheme, identifier)
);

COMMENT ON TABLE lsif_data_definitions IS 'Associates (document, range) pairs with the import monikers attached to the range.';
COMMENT ON COLUMN lsif_data_definitions.dump_id IS 'The identifier of the associated dump in the lsif_uploads table (state=completed).';
COMMENT ON COLUMN lsif_data_definitions.scheme IS 'The moniker scheme.';
COMMENT ON COLUMN lsif_data_definitions.identifier IS 'The moniker identifier.';
COMMENT ON COLUMN lsif_data_definitions.data IS 'A gob-encoded payload conforming to an array of [LocationData](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@3.26/-/blob/enterprise/lib/codeintel/semantic/types.go#L106:6) types.';
COMMENT ON COLUMN lsif_data_definitions.schema_version IS 'The schema version of this row - used to determine presence and encoding of data.';
COMMENT ON COLUMN lsif_data_definitions.num_locations IS 'The number of locations stored in the data field.';

CREATE TABLE IF NOT EXISTS lsif_data_definitions_schema_versions (
    dump_id integer NOT NULL,
    min_schema_version integer,
    max_schema_version integer,
    PRIMARY KEY (dump_id)
);

COMMENT ON TABLE lsif_data_definitions_schema_versions IS 'Tracks the range of schema_versions for each upload in the lsif_data_definitions table.';
COMMENT ON COLUMN lsif_data_definitions_schema_versions.dump_id IS 'The identifier of the associated dump in the lsif_uploads table.';
COMMENT ON COLUMN lsif_data_definitions_schema_versions.min_schema_version IS 'A lower-bound on the `lsif_data_definitions.schema_version` where `lsif_data_definitions.dump_id = dump_id`.';
COMMENT ON COLUMN lsif_data_definitions_schema_versions.max_schema_version IS 'An upper-bound on the `lsif_data_definitions.schema_version` where `lsif_data_definitions.dump_id = dump_id`.';

CREATE TABLE IF NOT EXISTS lsif_data_documents (
    dump_id integer NOT NULL,
    path text NOT NULL,
    data bytea,
    schema_version integer NOT NULL,
    num_diagnostics integer NOT NULL,
    ranges bytea,
    hovers bytea,
    monikers bytea,
    packages bytea,
    diagnostics bytea,
    PRIMARY KEY (dump_id, path)
);

COMMENT ON TABLE lsif_data_documents IS 'Stores reference, hover text, moniker, and diagnostic data about a particular text document witin a dump.';
COMMENT ON COLUMN lsif_data_documents.dump_id IS 'The identifier of the associated dump in the lsif_uploads table (state=completed).';
COMMENT ON COLUMN lsif_data_documents.path IS 'The path of the text document relative to the associated dump root.';
COMMENT ON COLUMN lsif_data_documents.data IS 'A gob-encoded payload conforming to the [DocumentData](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@3.26/-/blob/enterprise/lib/codeintel/semantic/types.go#L13:6) type. This field is being migrated across ranges, hovers, monikers, packages, and diagnostics columns and will be removed in a future release of Sourcegraph.';
COMMENT ON COLUMN lsif_data_documents.schema_version IS 'The schema version of this row - used to determine presence and encoding of data.';
COMMENT ON COLUMN lsif_data_documents.num_diagnostics IS 'The number of diagnostics stored in the data field.';
COMMENT ON COLUMN lsif_data_documents.ranges IS 'A gob-encoded payload conforming to the [Ranges](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@3.26/-/blob/enterprise/lib/codeintel/semantic/types.go#L14:2) field of the DocumentDatatype.';
COMMENT ON COLUMN lsif_data_documents.hovers IS 'A gob-encoded payload conforming to the [HoversResults](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@3.26/-/blob/enterprise/lib/codeintel/semantic/types.go#L15:2) field of the DocumentDatatype.';
COMMENT ON COLUMN lsif_data_documents.monikers IS 'A gob-encoded payload conforming to the [Monikers](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@3.26/-/blob/enterprise/lib/codeintel/semantic/types.go#L16:2) field of the DocumentDatatype.';
COMMENT ON COLUMN lsif_data_documents.packages IS 'A gob-encoded payload conforming to the [PackageInformation](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@3.26/-/blob/enterprise/lib/codeintel/semantic/types.go#L17:2) field of the DocumentDatatype.';
COMMENT ON COLUMN lsif_data_documents.diagnostics IS 'A gob-encoded payload conforming to the [Diagnostics](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@3.26/-/blob/enterprise/lib/codeintel/semantic/types.go#L18:2) field of the DocumentDatatype.';

CREATE TABLE IF NOT EXISTS lsif_data_documents_schema_versions (
    dump_id integer NOT NULL,
    min_schema_version integer,
    max_schema_version integer,
    PRIMARY KEY (dump_id)
);

COMMENT ON TABLE lsif_data_documents_schema_versions IS 'Tracks the range of schema_versions for each upload in the lsif_data_documents table.';
COMMENT ON COLUMN lsif_data_documents_schema_versions.dump_id IS 'The identifier of the associated dump in the lsif_uploads table.';
COMMENT ON COLUMN lsif_data_documents_schema_versions.min_schema_version IS 'A lower-bound on the `lsif_data_documents.schema_version` where `lsif_data_documents.dump_id = dump_id`.';
COMMENT ON COLUMN lsif_data_documents_schema_versions.max_schema_version IS 'An upper-bound on the `lsif_data_documents.schema_version` where `lsif_data_documents.dump_id = dump_id`.';

CREATE TABLE IF NOT EXISTS lsif_data_implementations (
    dump_id integer NOT NULL,
    scheme text NOT NULL,
    identifier text NOT NULL,
    data bytea,
    schema_version integer NOT NULL,
    num_locations integer NOT NULL,
    PRIMARY KEY (dump_id, scheme, identifier)
);

COMMENT ON TABLE lsif_data_implementations IS 'Associates (document, range) pairs with the implementation monikers attached to the range.';
COMMENT ON COLUMN lsif_data_implementations.dump_id IS 'The identifier of the associated dump in the lsif_uploads table (state=completed).';
COMMENT ON COLUMN lsif_data_implementations.scheme IS 'The moniker scheme.';
COMMENT ON COLUMN lsif_data_implementations.identifier IS 'The moniker identifier.';
COMMENT ON COLUMN lsif_data_implementations.data IS 'A gob-encoded payload conforming to an array of [LocationData](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@3.26/-/blob/enterprise/lib/codeintel/semantic/types.go#L106:6) types.';
COMMENT ON COLUMN lsif_data_implementations.schema_version IS 'The schema version of this row - used to determine presence and encoding of data.';
COMMENT ON COLUMN lsif_data_implementations.num_locations IS 'The number of locations stored in the data field.';

CREATE TABLE IF NOT EXISTS lsif_data_implementations_schema_versions (
    dump_id integer NOT NULL,
    min_schema_version integer,
    max_schema_version integer,
    PRIMARY KEY (dump_id)
);

COMMENT ON TABLE lsif_data_implementations_schema_versions IS 'Tracks the range of schema_versions for each upload in the lsif_data_implementations table.';
COMMENT ON COLUMN lsif_data_implementations_schema_versions.dump_id IS 'The identifier of the associated dump in the lsif_uploads table.';
COMMENT ON COLUMN lsif_data_implementations_schema_versions.min_schema_version IS 'A lower-bound on the `lsif_data_implementations.schema_version` where `lsif_data_implementations.dump_id = dump_id`.';
COMMENT ON COLUMN lsif_data_implementations_schema_versions.max_schema_version IS 'An upper-bound on the `lsif_data_implementations.schema_version` where `lsif_data_implementations.dump_id = dump_id`.';

CREATE TABLE IF NOT EXISTS lsif_data_metadata (
    dump_id integer NOT NULL,
    num_result_chunks integer,
    PRIMARY KEY (dump_id)
);

COMMENT ON TABLE lsif_data_metadata IS 'Stores the number of result chunks associated with a dump.';
COMMENT ON COLUMN lsif_data_metadata.dump_id IS 'The identifier of the associated dump in the lsif_uploads table (state=completed).';
COMMENT ON COLUMN lsif_data_metadata.num_result_chunks IS 'A bound of populated indexes in the lsif_data_result_chunks table for the associated dump. This value is used to hash identifiers into the result chunk index to which they belong.';

CREATE TABLE IF NOT EXISTS lsif_data_references (
    dump_id integer NOT NULL,
    scheme text NOT NULL,
    identifier text NOT NULL,
    data bytea,
    schema_version integer NOT NULL,
    num_locations integer NOT NULL,
    PRIMARY KEY (dump_id, scheme, identifier)
);

COMMENT ON TABLE lsif_data_references IS 'Associates (document, range) pairs with the export monikers attached to the range.';
COMMENT ON COLUMN lsif_data_references.dump_id IS 'The identifier of the associated dump in the lsif_uploads table (state=completed).';
COMMENT ON COLUMN lsif_data_references.scheme IS 'The moniker scheme.';
COMMENT ON COLUMN lsif_data_references.identifier IS 'The moniker identifier.';
COMMENT ON COLUMN lsif_data_references.data IS 'A gob-encoded payload conforming to an array of [LocationData](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@3.26/-/blob/enterprise/lib/codeintel/semantic/types.go#L106:6) types.';
COMMENT ON COLUMN lsif_data_references.schema_version IS 'The schema version of this row - used to determine presence and encoding of data.';
COMMENT ON COLUMN lsif_data_references.num_locations IS 'The number of locations stored in the data field.';

CREATE TABLE IF NOT EXISTS lsif_data_references_schema_versions (
    dump_id integer NOT NULL,
    min_schema_version integer,
    max_schema_version integer,
    PRIMARY KEY (dump_id)
);

COMMENT ON TABLE lsif_data_references_schema_versions IS 'Tracks the range of schema_versions for each upload in the lsif_data_references table.';
COMMENT ON COLUMN lsif_data_references_schema_versions.dump_id IS 'The identifier of the associated dump in the lsif_uploads table.';
COMMENT ON COLUMN lsif_data_references_schema_versions.min_schema_version IS 'A lower-bound on the `lsif_data_references.schema_version` where `lsif_data_references.dump_id = dump_id`.';
COMMENT ON COLUMN lsif_data_references_schema_versions.max_schema_version IS 'An upper-bound on the `lsif_data_references.schema_version` where `lsif_data_references.dump_id = dump_id`.';

CREATE TABLE IF NOT EXISTS lsif_data_result_chunks (
    dump_id integer NOT NULL,
    idx integer NOT NULL,
    data bytea,
    PRIMARY KEY (dump_id, idx)
);

COMMENT ON TABLE lsif_data_result_chunks IS 'Associates result set identifiers with the (document path, range identifier) pairs that compose the set.';
COMMENT ON COLUMN lsif_data_result_chunks.dump_id IS 'The identifier of the associated dump in the lsif_uploads table (state=completed).';
COMMENT ON COLUMN lsif_data_result_chunks.idx IS 'The unique result chunk index within the associated dump. Every result set identifier present should hash to this index (modulo lsif_data_metadata.num_result_chunks).';
COMMENT ON COLUMN lsif_data_result_chunks.data IS 'A gob-encoded payload conforming to the [ResultChunkData](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@3.26/-/blob/enterprise/lib/codeintel/semantic/types.go#L76:6) type.';

CREATE INDEX IF NOT EXISTS lsif_data_definitions_dump_id_schema_version ON lsif_data_definitions USING btree (dump_id, schema_version);
CREATE INDEX IF NOT EXISTS lsif_data_definitions_schema_versions_dump_id_schema_version_bo ON lsif_data_definitions_schema_versions USING btree (dump_id, min_schema_version, max_schema_version);
CREATE INDEX IF NOT EXISTS lsif_data_documents_dump_id_schema_version ON lsif_data_documents USING btree (dump_id, schema_version);
CREATE INDEX IF NOT EXISTS lsif_data_documents_schema_versions_dump_id_schema_version_boun ON lsif_data_documents_schema_versions USING btree (dump_id, min_schema_version, max_schema_version);
CREATE INDEX IF NOT EXISTS lsif_data_implementations_dump_id_schema_version ON lsif_data_implementations USING btree (dump_id, schema_version);
CREATE INDEX IF NOT EXISTS lsif_data_implementations_schema_versions_dump_id_schema_versio ON lsif_data_implementations_schema_versions USING btree (dump_id, min_schema_version, max_schema_version);
CREATE INDEX IF NOT EXISTS lsif_data_references_dump_id_schema_version ON lsif_data_references USING btree (dump_id, schema_version);
CREATE INDEX IF NOT EXISTS lsif_data_references_schema_versions_dump_id_schema_version_bou ON lsif_data_references_schema_versions USING btree (dump_id, min_schema_version, max_schema_version);

DROP TRIGGER IF EXISTS lsif_data_definitions_schema_versions_insert ON lsif_data_definitions;
DROP TRIGGER IF EXISTS lsif_data_documents_schema_versions_insert ON lsif_data_documents;
DROP TRIGGER IF EXISTS lsif_data_implementations_schema_versions_insert ON lsif_data_implementations;
DROP TRIGGER IF EXISTS lsif_data_references_schema_versions_insert ON lsif_data_references;

CREATE TRIGGER lsif_data_definitions_schema_versions_insert AFTER INSERT ON lsif_data_definitions REFERENCING NEW TABLE AS newtab FOR EACH STATEMENT EXECUTE FUNCTION update_lsif_data_definitions_schema_versions_insert();
CREATE TRIGGER lsif_data_documents_schema_versions_insert AFTER INSERT ON lsif_data_documents REFERENCING NEW TABLE AS newtab FOR EACH STATEMENT EXECUTE FUNCTION update_lsif_data_documents_schema_versions_insert();
CREATE TRIGGER lsif_data_implementations_schema_versions_insert AFTER INSERT ON lsif_data_implementations REFERENCING NEW TABLE AS newtab FOR EACH STATEMENT EXECUTE FUNCTION update_lsif_data_implementations_schema_versions_insert();
CREATE TRIGGER lsif_data_references_schema_versions_insert AFTER INSERT ON lsif_data_references REFERENCING NEW TABLE AS newtab FOR EACH STATEMENT EXECUTE FUNCTION update_lsif_data_references_schema_versions_insert();
