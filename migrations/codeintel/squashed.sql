CREATE EXTENSION IF NOT EXISTS intarray;

COMMENT ON EXTENSION intarray IS 'functions, operators, and index support for 1-D arrays of integers';

CREATE EXTENSION IF NOT EXISTS pg_stat_statements;

COMMENT ON EXTENSION pg_stat_statements IS 'track execution statistics of all SQL statements executed';

CREATE EXTENSION IF NOT EXISTS pg_trgm;

COMMENT ON EXTENSION pg_trgm IS 'text similarity measurement and index searching based on trigrams';

CREATE FUNCTION get_file_extension(path text) RETURNS text
    LANGUAGE plpgsql IMMUTABLE
    AS $_$ BEGIN
    RETURN substring(path FROM '\.([^\.]*)$');
END; $_$;

CREATE FUNCTION path_prefixes(path text) RETURNS text[]
    LANGUAGE plpgsql IMMUTABLE
    AS $$ BEGIN
    RETURN (
        SELECT array_agg(array_to_string(components[:len], '/')) prefixes
        FROM
            (SELECT regexp_split_to_array(path, E'/') components) t,
            generate_series(1, array_length(components, 1)) AS len
    );
END; $$;

CREATE FUNCTION singleton(value text) RETURNS text[]
    LANGUAGE plpgsql IMMUTABLE
    AS $$ BEGIN
    RETURN ARRAY[value];
END; $$;

CREATE FUNCTION singleton_integer(value integer) RETURNS integer[]
    LANGUAGE plpgsql IMMUTABLE
    AS $$ BEGIN
    RETURN ARRAY[value];
END; $$;

CREATE FUNCTION update_codeintel_scip_document_lookup_schema_versions_insert() RETURNS trigger
    LANGUAGE plpgsql
    AS $$ BEGIN
    INSERT INTO codeintel_scip_document_lookup_schema_versions
    SELECT
        upload_id,
        MIN(schema_version) as min_schema_version,
        MAX(schema_version) as max_schema_version
    FROM newtab
    JOIN codeintel_scip_documents ON codeintel_scip_documents.id = newtab.document_id
    GROUP BY newtab.upload_id
    ON CONFLICT (upload_id) DO UPDATE SET
        -- Update with min(old_min, new_min) and max(old_max, new_max)
        min_schema_version = LEAST(codeintel_scip_document_lookup_schema_versions.min_schema_version, EXCLUDED.min_schema_version),
        max_schema_version = GREATEST(codeintel_scip_document_lookup_schema_versions.max_schema_version, EXCLUDED.max_schema_version);
    RETURN NULL;
END $$;

CREATE FUNCTION update_codeintel_scip_documents_dereference_logs_delete() RETURNS trigger
    LANGUAGE plpgsql
    AS $$ BEGIN
    INSERT INTO codeintel_scip_documents_dereference_logs (document_id)
    SELECT document_id FROM oldtab;
    RETURN NULL;
END $$;

CREATE FUNCTION update_codeintel_scip_documents_schema_versions_insert() RETURNS trigger
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

CREATE FUNCTION update_codeintel_scip_symbols_schema_versions_insert() RETURNS trigger
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

CREATE FUNCTION update_lsif_data_definitions_schema_versions_insert() RETURNS trigger
    LANGUAGE plpgsql
    AS $$ BEGIN
    INSERT INTO
        lsif_data_definitions_schema_versions
    SELECT
        dump_id,
        MIN(schema_version) as min_schema_version,
        MAX(schema_version) as max_schema_version
    FROM
        newtab
    GROUP BY
        dump_id
    ON CONFLICT (dump_id) DO UPDATE SET
        -- Update with min(old_min, new_min) and max(old_max, new_max)
        min_schema_version = LEAST(lsif_data_definitions_schema_versions.min_schema_version, EXCLUDED.min_schema_version),
        max_schema_version = GREATEST(lsif_data_definitions_schema_versions.max_schema_version, EXCLUDED.max_schema_version);

    RETURN NULL;
END $$;

CREATE FUNCTION update_lsif_data_documents_schema_versions_insert() RETURNS trigger
    LANGUAGE plpgsql
    AS $$ BEGIN
    INSERT INTO
        lsif_data_documents_schema_versions
    SELECT
        dump_id,
        MIN(schema_version) as min_schema_version,
        MAX(schema_version) as max_schema_version
    FROM
        newtab
    GROUP BY
        dump_id
    ON CONFLICT (dump_id) DO UPDATE SET
        -- Update with min(old_min, new_min) and max(old_max, new_max)
        min_schema_version = LEAST(lsif_data_documents_schema_versions.min_schema_version, EXCLUDED.min_schema_version),
        max_schema_version = GREATEST(lsif_data_documents_schema_versions.max_schema_version, EXCLUDED.max_schema_version);

    RETURN NULL;
END $$;

CREATE FUNCTION update_lsif_data_implementations_schema_versions_insert() RETURNS trigger
    LANGUAGE plpgsql
    AS $$ BEGIN
    INSERT INTO
        lsif_data_implementations_schema_versions
    SELECT
        dump_id,
        MIN(schema_version) as min_schema_version,
        MAX(schema_version) as max_schema_version
    FROM
        newtab
    GROUP BY
        dump_id
    ON CONFLICT (dump_id) DO UPDATE SET
        -- Update with min(old_min, new_min) and max(old_max, new_max)
        min_schema_version = LEAST   (lsif_data_implementations_schema_versions.min_schema_version, EXCLUDED.min_schema_version),
        max_schema_version = GREATEST(lsif_data_implementations_schema_versions.max_schema_version, EXCLUDED.max_schema_version);

    RETURN NULL;
END $$;

CREATE FUNCTION update_lsif_data_references_schema_versions_insert() RETURNS trigger
    LANGUAGE plpgsql
    AS $$ BEGIN
    INSERT INTO
        lsif_data_references_schema_versions
    SELECT
        dump_id,
        MIN(schema_version) as min_schema_version,
        MAX(schema_version) as max_schema_version
    FROM
        newtab
    GROUP BY
        dump_id
    ON CONFLICT (dump_id) DO UPDATE SET
        -- Update with min(old_min, new_min) and max(old_max, new_max)
        min_schema_version = LEAST(lsif_data_references_schema_versions.min_schema_version, EXCLUDED.min_schema_version),
        max_schema_version = GREATEST(lsif_data_references_schema_versions.max_schema_version, EXCLUDED.max_schema_version);

    RETURN NULL;
END $$;

CREATE TABLE codeintel_last_reconcile (
    dump_id integer NOT NULL,
    last_reconcile_at timestamp with time zone NOT NULL
);

COMMENT ON TABLE codeintel_last_reconcile IS 'Stores the last time processed LSIF data was reconciled with the other database.';

CREATE TABLE codeintel_scip_document_lookup (
    id bigint NOT NULL,
    upload_id integer NOT NULL,
    document_path text NOT NULL,
    document_id bigint NOT NULL
);

COMMENT ON TABLE codeintel_scip_document_lookup IS 'A mapping from file paths to document references within a particular SCIP index.';

COMMENT ON COLUMN codeintel_scip_document_lookup.id IS 'An auto-generated identifier. This column is used as a foreign key target to reduce occurrences of the full document path value.';

COMMENT ON COLUMN codeintel_scip_document_lookup.upload_id IS 'The identifier of the upload that provided this SCIP index.';

COMMENT ON COLUMN codeintel_scip_document_lookup.document_path IS 'The file path to the document relative to the root of the index.';

COMMENT ON COLUMN codeintel_scip_document_lookup.document_id IS 'The foreign key to the shared document payload (see the table [`codeintel_scip_document_lookup`](#table-publiccodeintel_scip_document_lookup)).';

CREATE SEQUENCE codeintel_scip_document_lookup_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE codeintel_scip_document_lookup_id_seq OWNED BY codeintel_scip_document_lookup.id;

CREATE TABLE codeintel_scip_document_lookup_schema_versions (
    upload_id integer NOT NULL,
    min_schema_version integer,
    max_schema_version integer
);

COMMENT ON TABLE codeintel_scip_document_lookup_schema_versions IS 'Tracks the range of `schema_versions` values associated with each SCIP index in the [`codeintel_scip_document_lookup`](#table-publiccodeintel_scip_document_lookup) table.';

COMMENT ON COLUMN codeintel_scip_document_lookup_schema_versions.upload_id IS 'The identifier of the associated SCIP index.';

COMMENT ON COLUMN codeintel_scip_document_lookup_schema_versions.min_schema_version IS 'A lower-bound on the `schema_version` values of the records in the table [`codeintel_scip_document_lookup`](#table-publiccodeintel_scip_document_lookup) where the `upload_id` column matches the associated SCIP index.';

COMMENT ON COLUMN codeintel_scip_document_lookup_schema_versions.max_schema_version IS 'An upper-bound on the `schema_version` values of the records in the table [`codeintel_scip_document_lookup`](#table-publiccodeintel_scip_document_lookup) where the `upload_id` column matches the associated SCIP index.';

CREATE TABLE codeintel_scip_documents (
    id bigint NOT NULL,
    payload_hash bytea NOT NULL,
    schema_version integer NOT NULL,
    raw_scip_payload bytea NOT NULL
);

COMMENT ON TABLE codeintel_scip_documents IS 'A lookup of SCIP [Document](https://sourcegraph.com/search?q=context:%40sourcegraph/all+repo:%5Egithub%5C.com/sourcegraph/scip%24+file:%5Escip%5C.proto+message+Document&patternType=standard) payloads by their hash.';

COMMENT ON COLUMN codeintel_scip_documents.id IS 'An auto-generated identifier. This column is used as a foreign key target to reduce occurrences of the full payload hash value.';

COMMENT ON COLUMN codeintel_scip_documents.payload_hash IS 'A deterministic hash of the raw SCIP payload. We use this as a unique value to enforce deduplication between indexes with the same document data.';

COMMENT ON COLUMN codeintel_scip_documents.schema_version IS 'The schema version of this row - used to determine presence and encoding of (future) denormalized data.';

COMMENT ON COLUMN codeintel_scip_documents.raw_scip_payload IS 'The raw, canonicalized SCIP [Document](https://sourcegraph.com/search?q=context:%40sourcegraph/all+repo:%5Egithub%5C.com/sourcegraph/scip%24+file:%5Escip%5C.proto+message+Document&patternType=standard) payload.';

CREATE TABLE codeintel_scip_documents_dereference_logs (
    id bigint NOT NULL,
    document_id bigint NOT NULL,
    last_removal_time timestamp with time zone DEFAULT now() NOT NULL
);

COMMENT ON TABLE codeintel_scip_documents_dereference_logs IS 'A list of document rows that were recently dereferenced by the deletion of an index.';

COMMENT ON COLUMN codeintel_scip_documents_dereference_logs.document_id IS 'The identifier of the document that was dereferenced.';

COMMENT ON COLUMN codeintel_scip_documents_dereference_logs.last_removal_time IS 'The time that the log entry was inserted.';

CREATE SEQUENCE codeintel_scip_documents_dereference_logs_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE codeintel_scip_documents_dereference_logs_id_seq OWNED BY codeintel_scip_documents_dereference_logs.id;

CREATE SEQUENCE codeintel_scip_documents_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE codeintel_scip_documents_id_seq OWNED BY codeintel_scip_documents.id;

CREATE TABLE codeintel_scip_documents_schema_versions (
    upload_id integer NOT NULL,
    min_schema_version integer,
    max_schema_version integer
);

COMMENT ON TABLE codeintel_scip_documents_schema_versions IS 'Tracks the range of `schema_versions` values associated with each document referenced from the [`codeintel_scip_document_lookup`](#table-publiccodeintel_scip_document_lookup) table.';

COMMENT ON COLUMN codeintel_scip_documents_schema_versions.upload_id IS 'The identifier of the associated SCIP index.';

COMMENT ON COLUMN codeintel_scip_documents_schema_versions.min_schema_version IS 'A lower-bound on the `schema_version` values of the document records referenced from the table [`codeintel_scip_document_lookup`](#table-publiccodeintel_scip_document_lookup) where the `upload_id` column matches the associated SCIP index.';

COMMENT ON COLUMN codeintel_scip_documents_schema_versions.max_schema_version IS 'An upper-bound on the `schema_version` values of the document records referenced from the table [`codeintel_scip_document_lookup`](#table-publiccodeintel_scip_document_lookup) where the `upload_id` column matches the associated SCIP index.';

CREATE TABLE codeintel_scip_metadata (
    id bigint NOT NULL,
    upload_id integer NOT NULL,
    tool_name text NOT NULL,
    tool_version text NOT NULL,
    tool_arguments text[] NOT NULL,
    text_document_encoding text NOT NULL,
    protocol_version integer NOT NULL
);

COMMENT ON TABLE codeintel_scip_metadata IS 'Global metadatadata about a single processed upload.';

COMMENT ON COLUMN codeintel_scip_metadata.id IS 'An auto-generated identifier.';

COMMENT ON COLUMN codeintel_scip_metadata.upload_id IS 'The identifier of the upload that provided this SCIP index.';

COMMENT ON COLUMN codeintel_scip_metadata.tool_name IS 'Name of the indexer that produced this index.';

COMMENT ON COLUMN codeintel_scip_metadata.tool_version IS 'Version of the indexer that produced this index.';

COMMENT ON COLUMN codeintel_scip_metadata.tool_arguments IS 'Command-line arguments that were used to invoke this indexer.';

COMMENT ON COLUMN codeintel_scip_metadata.text_document_encoding IS 'The encoding of the text documents within this index. May affect range boundaries.';

COMMENT ON COLUMN codeintel_scip_metadata.protocol_version IS 'The version of the SCIP protocol used to encode this index.';

CREATE SEQUENCE codeintel_scip_metadata_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE codeintel_scip_metadata_id_seq OWNED BY codeintel_scip_metadata.id;

CREATE TABLE codeintel_scip_symbol_names (
    id integer NOT NULL,
    upload_id integer NOT NULL,
    name_segment text NOT NULL,
    prefix_id integer
);

COMMENT ON TABLE codeintel_scip_symbol_names IS 'Stores a prefix tree of symbol names within a particular upload.';

COMMENT ON COLUMN codeintel_scip_symbol_names.id IS 'An identifier unique within the index for this symbol name segment.';

COMMENT ON COLUMN codeintel_scip_symbol_names.upload_id IS 'The identifier of the upload that provided this SCIP index.';

COMMENT ON COLUMN codeintel_scip_symbol_names.name_segment IS 'The portion of the symbol name that is unique to this symbol and its children.';

COMMENT ON COLUMN codeintel_scip_symbol_names.prefix_id IS 'The identifier of the segment that forms the prefix of this symbol, if any.';

CREATE TABLE codeintel_scip_symbols (
    upload_id integer NOT NULL,
    document_lookup_id bigint NOT NULL,
    schema_version integer NOT NULL,
    definition_ranges bytea,
    reference_ranges bytea,
    implementation_ranges bytea,
    type_definition_ranges bytea,
    symbol_id integer NOT NULL
);

COMMENT ON TABLE codeintel_scip_symbols IS 'A mapping from SCIP [Symbol names](https://sourcegraph.com/search?q=context:%40sourcegraph/all+repo:%5Egithub%5C.com/sourcegraph/scip%24+file:%5Escip%5C.proto+message+Symbol&patternType=standard) to path and ranges where that symbol occurs within a particular SCIP index.';

COMMENT ON COLUMN codeintel_scip_symbols.upload_id IS 'The identifier of the upload that provided this SCIP index.';

COMMENT ON COLUMN codeintel_scip_symbols.document_lookup_id IS 'A reference to the `id` column of [`codeintel_scip_document_lookup`](#table-publiccodeintel_scip_document_lookup). Joining on this table yields the document path relative to the index root.';

COMMENT ON COLUMN codeintel_scip_symbols.schema_version IS 'The schema version of this row - used to determine presence and encoding of denormalized data.';

COMMENT ON COLUMN codeintel_scip_symbols.definition_ranges IS 'An encoded set of ranges within the associated document that have a **definition** relationship to the associated symbol.';

COMMENT ON COLUMN codeintel_scip_symbols.reference_ranges IS 'An encoded set of ranges within the associated document that have a **reference** relationship to the associated symbol.';

COMMENT ON COLUMN codeintel_scip_symbols.implementation_ranges IS 'An encoded set of ranges within the associated document that have a **implementation** relationship to the associated symbol.';

COMMENT ON COLUMN codeintel_scip_symbols.type_definition_ranges IS 'An encoded set of ranges within the associated document that have a **type definition** relationship to the associated symbol.';

COMMENT ON COLUMN codeintel_scip_symbols.symbol_id IS 'The identifier of the segment that terminates the name of this symbol. See the table [`codeintel_scip_symbol_names`](#table-publiccodeintel_scip_symbol_names) on how to reconstruct the full symbol name.';

CREATE TABLE codeintel_scip_symbols_schema_versions (
    upload_id integer NOT NULL,
    min_schema_version integer,
    max_schema_version integer
);

COMMENT ON TABLE codeintel_scip_symbols_schema_versions IS 'Tracks the range of `schema_versions` for each index in the [`codeintel_scip_symbols`](#table-publiccodeintel_scip_symbols) table.';

COMMENT ON COLUMN codeintel_scip_symbols_schema_versions.upload_id IS 'The identifier of the associated SCIP index.';

COMMENT ON COLUMN codeintel_scip_symbols_schema_versions.min_schema_version IS 'A lower-bound on the `schema_version` values of the records in the table [`codeintel_scip_symbols`](#table-publiccodeintel_scip_symbols) where the `upload_id` column matches the associated SCIP index.';

COMMENT ON COLUMN codeintel_scip_symbols_schema_versions.max_schema_version IS 'An upper-bound on the `schema_version` values of the records in the table [`codeintel_scip_symbols`](#table-publiccodeintel_scip_symbols) where the `upload_id` column matches the associated SCIP index.';

CREATE TABLE lsif_data_definitions (
    dump_id integer NOT NULL,
    scheme text NOT NULL,
    identifier text NOT NULL,
    data bytea,
    schema_version integer NOT NULL,
    num_locations integer NOT NULL
);

COMMENT ON TABLE lsif_data_definitions IS 'Associates (document, range) pairs with the import monikers attached to the range.';

COMMENT ON COLUMN lsif_data_definitions.dump_id IS 'The identifier of the associated dump in the lsif_uploads table (state=completed).';

COMMENT ON COLUMN lsif_data_definitions.scheme IS 'The moniker scheme.';

COMMENT ON COLUMN lsif_data_definitions.identifier IS 'The moniker identifier.';

COMMENT ON COLUMN lsif_data_definitions.data IS 'A gob-encoded payload conforming to an array of [LocationData](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@3.26/-/blob/enterprise/lib/codeintel/semantic/types.go#L106:6) types.';

COMMENT ON COLUMN lsif_data_definitions.schema_version IS 'The schema version of this row - used to determine presence and encoding of data.';

COMMENT ON COLUMN lsif_data_definitions.num_locations IS 'The number of locations stored in the data field.';

CREATE TABLE lsif_data_definitions_schema_versions (
    dump_id integer NOT NULL,
    min_schema_version integer,
    max_schema_version integer
);

COMMENT ON TABLE lsif_data_definitions_schema_versions IS 'Tracks the range of schema_versions for each upload in the lsif_data_definitions table.';

COMMENT ON COLUMN lsif_data_definitions_schema_versions.dump_id IS 'The identifier of the associated dump in the lsif_uploads table.';

COMMENT ON COLUMN lsif_data_definitions_schema_versions.min_schema_version IS 'A lower-bound on the `lsif_data_definitions.schema_version` where `lsif_data_definitions.dump_id = dump_id`.';

COMMENT ON COLUMN lsif_data_definitions_schema_versions.max_schema_version IS 'An upper-bound on the `lsif_data_definitions.schema_version` where `lsif_data_definitions.dump_id = dump_id`.';

CREATE TABLE lsif_data_documents (
    dump_id integer NOT NULL,
    path text NOT NULL,
    data bytea,
    schema_version integer NOT NULL,
    num_diagnostics integer NOT NULL,
    ranges bytea,
    hovers bytea,
    monikers bytea,
    packages bytea,
    diagnostics bytea
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

CREATE TABLE lsif_data_documents_schema_versions (
    dump_id integer NOT NULL,
    min_schema_version integer,
    max_schema_version integer
);

COMMENT ON TABLE lsif_data_documents_schema_versions IS 'Tracks the range of schema_versions for each upload in the lsif_data_documents table.';

COMMENT ON COLUMN lsif_data_documents_schema_versions.dump_id IS 'The identifier of the associated dump in the lsif_uploads table.';

COMMENT ON COLUMN lsif_data_documents_schema_versions.min_schema_version IS 'A lower-bound on the `lsif_data_documents.schema_version` where `lsif_data_documents.dump_id = dump_id`.';

COMMENT ON COLUMN lsif_data_documents_schema_versions.max_schema_version IS 'An upper-bound on the `lsif_data_documents.schema_version` where `lsif_data_documents.dump_id = dump_id`.';

CREATE TABLE lsif_data_implementations (
    dump_id integer NOT NULL,
    scheme text NOT NULL,
    identifier text NOT NULL,
    data bytea,
    schema_version integer NOT NULL,
    num_locations integer NOT NULL
);

COMMENT ON TABLE lsif_data_implementations IS 'Associates (document, range) pairs with the implementation monikers attached to the range.';

COMMENT ON COLUMN lsif_data_implementations.dump_id IS 'The identifier of the associated dump in the lsif_uploads table (state=completed).';

COMMENT ON COLUMN lsif_data_implementations.scheme IS 'The moniker scheme.';

COMMENT ON COLUMN lsif_data_implementations.identifier IS 'The moniker identifier.';

COMMENT ON COLUMN lsif_data_implementations.data IS 'A gob-encoded payload conforming to an array of [LocationData](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@3.26/-/blob/enterprise/lib/codeintel/semantic/types.go#L106:6) types.';

COMMENT ON COLUMN lsif_data_implementations.schema_version IS 'The schema version of this row - used to determine presence and encoding of data.';

COMMENT ON COLUMN lsif_data_implementations.num_locations IS 'The number of locations stored in the data field.';

CREATE TABLE lsif_data_implementations_schema_versions (
    dump_id integer NOT NULL,
    min_schema_version integer,
    max_schema_version integer
);

COMMENT ON TABLE lsif_data_implementations_schema_versions IS 'Tracks the range of schema_versions for each upload in the lsif_data_implementations table.';

COMMENT ON COLUMN lsif_data_implementations_schema_versions.dump_id IS 'The identifier of the associated dump in the lsif_uploads table.';

COMMENT ON COLUMN lsif_data_implementations_schema_versions.min_schema_version IS 'A lower-bound on the `lsif_data_implementations.schema_version` where `lsif_data_implementations.dump_id = dump_id`.';

COMMENT ON COLUMN lsif_data_implementations_schema_versions.max_schema_version IS 'An upper-bound on the `lsif_data_implementations.schema_version` where `lsif_data_implementations.dump_id = dump_id`.';

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
    data bytea,
    schema_version integer NOT NULL,
    num_locations integer NOT NULL
);

COMMENT ON TABLE lsif_data_references IS 'Associates (document, range) pairs with the export monikers attached to the range.';

COMMENT ON COLUMN lsif_data_references.dump_id IS 'The identifier of the associated dump in the lsif_uploads table (state=completed).';

COMMENT ON COLUMN lsif_data_references.scheme IS 'The moniker scheme.';

COMMENT ON COLUMN lsif_data_references.identifier IS 'The moniker identifier.';

COMMENT ON COLUMN lsif_data_references.data IS 'A gob-encoded payload conforming to an array of [LocationData](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@3.26/-/blob/enterprise/lib/codeintel/semantic/types.go#L106:6) types.';

COMMENT ON COLUMN lsif_data_references.schema_version IS 'The schema version of this row - used to determine presence and encoding of data.';

COMMENT ON COLUMN lsif_data_references.num_locations IS 'The number of locations stored in the data field.';

CREATE TABLE lsif_data_references_schema_versions (
    dump_id integer NOT NULL,
    min_schema_version integer,
    max_schema_version integer
);

COMMENT ON TABLE lsif_data_references_schema_versions IS 'Tracks the range of schema_versions for each upload in the lsif_data_references table.';

COMMENT ON COLUMN lsif_data_references_schema_versions.dump_id IS 'The identifier of the associated dump in the lsif_uploads table.';

COMMENT ON COLUMN lsif_data_references_schema_versions.min_schema_version IS 'A lower-bound on the `lsif_data_references.schema_version` where `lsif_data_references.dump_id = dump_id`.';

COMMENT ON COLUMN lsif_data_references_schema_versions.max_schema_version IS 'An upper-bound on the `lsif_data_references.schema_version` where `lsif_data_references.dump_id = dump_id`.';

CREATE TABLE lsif_data_result_chunks (
    dump_id integer NOT NULL,
    idx integer NOT NULL,
    data bytea
);

COMMENT ON TABLE lsif_data_result_chunks IS 'Associates result set identifiers with the (document path, range identifier) pairs that compose the set.';

COMMENT ON COLUMN lsif_data_result_chunks.dump_id IS 'The identifier of the associated dump in the lsif_uploads table (state=completed).';

COMMENT ON COLUMN lsif_data_result_chunks.idx IS 'The unique result chunk index within the associated dump. Every result set identifier present should hash to this index (modulo lsif_data_metadata.num_result_chunks).';

COMMENT ON COLUMN lsif_data_result_chunks.data IS 'A gob-encoded payload conforming to the [ResultChunkData](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@3.26/-/blob/enterprise/lib/codeintel/semantic/types.go#L76:6) type.';

CREATE TABLE rockskip_ancestry (
    id integer NOT NULL,
    repo_id integer NOT NULL,
    commit_id character varying(40) NOT NULL,
    height integer NOT NULL,
    ancestor integer NOT NULL
);

CREATE SEQUENCE rockskip_ancestry_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE rockskip_ancestry_id_seq OWNED BY rockskip_ancestry.id;

CREATE TABLE rockskip_repos (
    id integer NOT NULL,
    repo text NOT NULL,
    last_accessed_at timestamp with time zone NOT NULL
);

CREATE SEQUENCE rockskip_repos_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE rockskip_repos_id_seq OWNED BY rockskip_repos.id;

CREATE TABLE rockskip_symbols (
    id integer NOT NULL,
    added integer[] NOT NULL,
    deleted integer[] NOT NULL,
    repo_id integer NOT NULL,
    path text NOT NULL,
    name text NOT NULL
);

CREATE SEQUENCE rockskip_symbols_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE rockskip_symbols_id_seq OWNED BY rockskip_symbols.id;

ALTER TABLE ONLY codeintel_scip_document_lookup ALTER COLUMN id SET DEFAULT nextval('codeintel_scip_document_lookup_id_seq'::regclass);

ALTER TABLE ONLY codeintel_scip_documents ALTER COLUMN id SET DEFAULT nextval('codeintel_scip_documents_id_seq'::regclass);

ALTER TABLE ONLY codeintel_scip_documents_dereference_logs ALTER COLUMN id SET DEFAULT nextval('codeintel_scip_documents_dereference_logs_id_seq'::regclass);

ALTER TABLE ONLY codeintel_scip_metadata ALTER COLUMN id SET DEFAULT nextval('codeintel_scip_metadata_id_seq'::regclass);

ALTER TABLE ONLY rockskip_ancestry ALTER COLUMN id SET DEFAULT nextval('rockskip_ancestry_id_seq'::regclass);

ALTER TABLE ONLY rockskip_repos ALTER COLUMN id SET DEFAULT nextval('rockskip_repos_id_seq'::regclass);

ALTER TABLE ONLY rockskip_symbols ALTER COLUMN id SET DEFAULT nextval('rockskip_symbols_id_seq'::regclass);

ALTER TABLE ONLY codeintel_last_reconcile
    ADD CONSTRAINT codeintel_last_reconcile_dump_id_key UNIQUE (dump_id);

ALTER TABLE ONLY codeintel_scip_document_lookup
    ADD CONSTRAINT codeintel_scip_document_lookup_pkey PRIMARY KEY (id);

ALTER TABLE ONLY codeintel_scip_document_lookup_schema_versions
    ADD CONSTRAINT codeintel_scip_document_lookup_schema_versions_pkey PRIMARY KEY (upload_id);

ALTER TABLE ONLY codeintel_scip_document_lookup
    ADD CONSTRAINT codeintel_scip_document_lookup_upload_id_document_path_key UNIQUE (upload_id, document_path);

ALTER TABLE ONLY codeintel_scip_documents_dereference_logs
    ADD CONSTRAINT codeintel_scip_documents_dereference_logs_pkey PRIMARY KEY (id);

ALTER TABLE ONLY codeintel_scip_documents
    ADD CONSTRAINT codeintel_scip_documents_payload_hash_key UNIQUE (payload_hash);

ALTER TABLE ONLY codeintel_scip_documents
    ADD CONSTRAINT codeintel_scip_documents_pkey PRIMARY KEY (id);

ALTER TABLE ONLY codeintel_scip_documents_schema_versions
    ADD CONSTRAINT codeintel_scip_documents_schema_versions_pkey PRIMARY KEY (upload_id);

ALTER TABLE ONLY codeintel_scip_metadata
    ADD CONSTRAINT codeintel_scip_metadata_pkey PRIMARY KEY (id);

ALTER TABLE ONLY codeintel_scip_symbol_names
    ADD CONSTRAINT codeintel_scip_symbol_names_pkey PRIMARY KEY (upload_id, id);

ALTER TABLE ONLY codeintel_scip_symbols
    ADD CONSTRAINT codeintel_scip_symbols_pkey PRIMARY KEY (upload_id, symbol_id, document_lookup_id);

ALTER TABLE ONLY codeintel_scip_symbols_schema_versions
    ADD CONSTRAINT codeintel_scip_symbols_schema_versions_pkey PRIMARY KEY (upload_id);

ALTER TABLE ONLY lsif_data_definitions
    ADD CONSTRAINT lsif_data_definitions_pkey PRIMARY KEY (dump_id, scheme, identifier);

ALTER TABLE ONLY lsif_data_definitions_schema_versions
    ADD CONSTRAINT lsif_data_definitions_schema_versions_pkey PRIMARY KEY (dump_id);

ALTER TABLE ONLY lsif_data_documents
    ADD CONSTRAINT lsif_data_documents_pkey PRIMARY KEY (dump_id, path);

ALTER TABLE ONLY lsif_data_documents_schema_versions
    ADD CONSTRAINT lsif_data_documents_schema_versions_pkey PRIMARY KEY (dump_id);

ALTER TABLE ONLY lsif_data_implementations
    ADD CONSTRAINT lsif_data_implementations_pkey PRIMARY KEY (dump_id, scheme, identifier);

ALTER TABLE ONLY lsif_data_implementations_schema_versions
    ADD CONSTRAINT lsif_data_implementations_schema_versions_pkey PRIMARY KEY (dump_id);

ALTER TABLE ONLY lsif_data_metadata
    ADD CONSTRAINT lsif_data_metadata_pkey PRIMARY KEY (dump_id);

ALTER TABLE ONLY lsif_data_references
    ADD CONSTRAINT lsif_data_references_pkey PRIMARY KEY (dump_id, scheme, identifier);

ALTER TABLE ONLY lsif_data_references_schema_versions
    ADD CONSTRAINT lsif_data_references_schema_versions_pkey PRIMARY KEY (dump_id);

ALTER TABLE ONLY lsif_data_result_chunks
    ADD CONSTRAINT lsif_data_result_chunks_pkey PRIMARY KEY (dump_id, idx);

ALTER TABLE ONLY rockskip_ancestry
    ADD CONSTRAINT rockskip_ancestry_pkey PRIMARY KEY (id);

ALTER TABLE ONLY rockskip_ancestry
    ADD CONSTRAINT rockskip_ancestry_repo_id_commit_id_key UNIQUE (repo_id, commit_id);

ALTER TABLE ONLY rockskip_repos
    ADD CONSTRAINT rockskip_repos_pkey PRIMARY KEY (id);

ALTER TABLE ONLY rockskip_repos
    ADD CONSTRAINT rockskip_repos_repo_key UNIQUE (repo);

ALTER TABLE ONLY rockskip_symbols
    ADD CONSTRAINT rockskip_symbols_pkey PRIMARY KEY (id);

CREATE INDEX codeintel_last_reconcile_last_reconcile_at_dump_id ON codeintel_last_reconcile USING btree (last_reconcile_at, dump_id);

CREATE INDEX codeintel_scip_document_lookup_document_id ON codeintel_scip_document_lookup USING hash (document_id);

CREATE INDEX codeintel_scip_documents_dereference_logs_last_removal_time_doc ON codeintel_scip_documents_dereference_logs USING btree (last_removal_time, document_id);

CREATE INDEX codeintel_scip_symbol_names_upload_id_roots ON codeintel_scip_symbol_names USING btree (upload_id) WHERE (prefix_id IS NULL);

CREATE INDEX codeintel_scip_symbols_document_lookup_id ON codeintel_scip_symbols USING btree (document_lookup_id);

CREATE INDEX codeisdntel_scip_symbol_names_upload_id_children ON codeintel_scip_symbol_names USING btree (upload_id, prefix_id) WHERE (prefix_id IS NOT NULL);

CREATE INDEX lsif_data_definitions_dump_id_schema_version ON lsif_data_definitions USING btree (dump_id, schema_version);

CREATE INDEX lsif_data_definitions_schema_versions_dump_id_schema_version_bo ON lsif_data_definitions_schema_versions USING btree (dump_id, min_schema_version, max_schema_version);

CREATE INDEX lsif_data_documents_dump_id_schema_version ON lsif_data_documents USING btree (dump_id, schema_version);

CREATE INDEX lsif_data_documents_schema_versions_dump_id_schema_version_boun ON lsif_data_documents_schema_versions USING btree (dump_id, min_schema_version, max_schema_version);

CREATE INDEX lsif_data_implementations_dump_id_schema_version ON lsif_data_implementations USING btree (dump_id, schema_version);

CREATE INDEX lsif_data_implementations_schema_versions_dump_id_schema_versio ON lsif_data_implementations_schema_versions USING btree (dump_id, min_schema_version, max_schema_version);

CREATE INDEX lsif_data_references_dump_id_schema_version ON lsif_data_references USING btree (dump_id, schema_version);

CREATE INDEX lsif_data_references_schema_versions_dump_id_schema_version_bou ON lsif_data_references_schema_versions USING btree (dump_id, min_schema_version, max_schema_version);

CREATE INDEX rockskip_ancestry_repo_commit_id ON rockskip_ancestry USING btree (repo_id, commit_id);

CREATE INDEX rockskip_repos_last_accessed_at ON rockskip_repos USING btree (last_accessed_at);

CREATE INDEX rockskip_repos_repo ON rockskip_repos USING btree (repo);

CREATE INDEX rockskip_symbols_gin ON rockskip_symbols USING gin (singleton_integer(repo_id) gin__int_ops, added gin__int_ops, deleted gin__int_ops, name gin_trgm_ops, singleton(name), singleton(lower(name)), path gin_trgm_ops, singleton(path), path_prefixes(path), singleton(lower(path)), path_prefixes(lower(path)), singleton(get_file_extension(path)), singleton(get_file_extension(lower(path))));

CREATE INDEX rockskip_symbols_repo_id_path_name ON rockskip_symbols USING btree (repo_id, path, name);

CREATE TRIGGER codeintel_scip_document_lookup_schema_versions_insert AFTER INSERT ON codeintel_scip_document_lookup REFERENCING NEW TABLE AS newtab FOR EACH STATEMENT EXECUTE FUNCTION update_codeintel_scip_document_lookup_schema_versions_insert();

CREATE TRIGGER codeintel_scip_documents_dereference_logs_insert AFTER DELETE ON codeintel_scip_document_lookup REFERENCING OLD TABLE AS oldtab FOR EACH STATEMENT EXECUTE FUNCTION update_codeintel_scip_documents_dereference_logs_delete();

CREATE TRIGGER codeintel_scip_documents_schema_versions_insert AFTER INSERT ON codeintel_scip_document_lookup REFERENCING NEW TABLE AS newtab FOR EACH STATEMENT EXECUTE FUNCTION update_codeintel_scip_documents_schema_versions_insert();

CREATE TRIGGER codeintel_scip_symbols_schema_versions_insert AFTER INSERT ON codeintel_scip_symbols REFERENCING NEW TABLE AS newtab FOR EACH STATEMENT EXECUTE FUNCTION update_codeintel_scip_symbols_schema_versions_insert();

CREATE TRIGGER lsif_data_definitions_schema_versions_insert AFTER INSERT ON lsif_data_definitions REFERENCING NEW TABLE AS newtab FOR EACH STATEMENT EXECUTE FUNCTION update_lsif_data_definitions_schema_versions_insert();

CREATE TRIGGER lsif_data_documents_schema_versions_insert AFTER INSERT ON lsif_data_documents REFERENCING NEW TABLE AS newtab FOR EACH STATEMENT EXECUTE FUNCTION update_lsif_data_documents_schema_versions_insert();

CREATE TRIGGER lsif_data_implementations_schema_versions_insert AFTER INSERT ON lsif_data_implementations REFERENCING NEW TABLE AS newtab FOR EACH STATEMENT EXECUTE FUNCTION update_lsif_data_implementations_schema_versions_insert();

CREATE TRIGGER lsif_data_references_schema_versions_insert AFTER INSERT ON lsif_data_references REFERENCING NEW TABLE AS newtab FOR EACH STATEMENT EXECUTE FUNCTION update_lsif_data_references_schema_versions_insert();

ALTER TABLE ONLY codeintel_scip_document_lookup
    ADD CONSTRAINT codeintel_scip_document_lookup_document_id_fk FOREIGN KEY (document_id) REFERENCES codeintel_scip_documents(id);

ALTER TABLE ONLY codeintel_scip_symbols
    ADD CONSTRAINT codeintel_scip_symbols_document_lookup_id_fk FOREIGN KEY (document_lookup_id) REFERENCES codeintel_scip_document_lookup(id) ON DELETE CASCADE;