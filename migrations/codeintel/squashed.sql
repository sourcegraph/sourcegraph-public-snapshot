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

ALTER TABLE ONLY rockskip_ancestry ALTER COLUMN id SET DEFAULT nextval('rockskip_ancestry_id_seq'::regclass);

ALTER TABLE ONLY rockskip_repos ALTER COLUMN id SET DEFAULT nextval('rockskip_repos_id_seq'::regclass);

ALTER TABLE ONLY rockskip_symbols ALTER COLUMN id SET DEFAULT nextval('rockskip_symbols_id_seq'::regclass);

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

CREATE TRIGGER lsif_data_definitions_schema_versions_insert AFTER INSERT ON lsif_data_definitions REFERENCING NEW TABLE AS newtab FOR EACH STATEMENT EXECUTE FUNCTION update_lsif_data_definitions_schema_versions_insert();

CREATE TRIGGER lsif_data_documents_schema_versions_insert AFTER INSERT ON lsif_data_documents REFERENCING NEW TABLE AS newtab FOR EACH STATEMENT EXECUTE FUNCTION update_lsif_data_documents_schema_versions_insert();

CREATE TRIGGER lsif_data_implementations_schema_versions_insert AFTER INSERT ON lsif_data_implementations REFERENCING NEW TABLE AS newtab FOR EACH STATEMENT EXECUTE FUNCTION update_lsif_data_implementations_schema_versions_insert();

CREATE TRIGGER lsif_data_references_schema_versions_insert AFTER INSERT ON lsif_data_references REFERENCING NEW TABLE AS newtab FOR EACH STATEMENT EXECUTE FUNCTION update_lsif_data_references_schema_versions_insert();