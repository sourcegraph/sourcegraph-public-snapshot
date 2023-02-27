CREATE FUNCTION get_file_extension(path text) RETURNS text
    LANGUAGE plpgsql IMMUTABLE
    AS $_$ BEGIN
    RETURN substring(path FROM '\.([^\.]*)$');
END; $_$;

CREATE FUNCTION lsif_data_docs_search_private_delete() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
UPDATE lsif_data_apidocs_num_search_results_private SET count = count - (select count(*) from oldtbl);
RETURN NULL;
END $$;

CREATE FUNCTION lsif_data_docs_search_private_insert() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
UPDATE lsif_data_apidocs_num_search_results_private SET count = count + (select count(*) from newtbl);
RETURN NULL;
END $$;

CREATE FUNCTION lsif_data_docs_search_public_delete() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
UPDATE lsif_data_apidocs_num_search_results_public SET count = count - (select count(*) from oldtbl);
RETURN NULL;
END $$;

CREATE FUNCTION lsif_data_docs_search_public_insert() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
UPDATE lsif_data_apidocs_num_search_results_public SET count = count + (select count(*) from newtbl);
RETURN NULL;
END $$;

CREATE FUNCTION lsif_data_documentation_pages_delete() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
UPDATE lsif_data_apidocs_num_pages SET count = count - (select count(*) from oldtbl);
UPDATE lsif_data_apidocs_num_dumps SET count = count - (select count(DISTINCT dump_id) from oldtbl);
UPDATE lsif_data_apidocs_num_dumps_indexed SET count = count - (select count(DISTINCT dump_id) from oldtbl WHERE search_indexed='true');
RETURN NULL;
END $$;

CREATE FUNCTION lsif_data_documentation_pages_insert() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
UPDATE lsif_data_apidocs_num_pages SET count = count + (select count(*) from newtbl);
UPDATE lsif_data_apidocs_num_dumps SET count = count + (select count(DISTINCT dump_id) from newtbl);
UPDATE lsif_data_apidocs_num_dumps_indexed SET count = count + (select count(DISTINCT dump_id) from newtbl WHERE search_indexed='true');
RETURN NULL;
END $$;

CREATE FUNCTION lsif_data_documentation_pages_update() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
WITH
    beforeIndexed AS (SELECT count(DISTINCT dump_id) FROM oldtbl WHERE search_indexed='true'),
    afterIndexed AS (SELECT count(DISTINCT dump_id) FROM newtbl WHERE search_indexed='true')
UPDATE lsif_data_apidocs_num_dumps_indexed SET count=count + ((select * from afterIndexed) - (select * from beforeIndexed));
RETURN NULL;
END $$;

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

CREATE TABLE lsif_data_apidocs_num_dumps (
    count bigint
);

CREATE TABLE lsif_data_apidocs_num_dumps_indexed (
    count bigint
);

CREATE TABLE lsif_data_apidocs_num_pages (
    count bigint
);

CREATE TABLE lsif_data_apidocs_num_search_results_private (
    count bigint
);

CREATE TABLE lsif_data_apidocs_num_search_results_public (
    count bigint
);

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

CREATE TABLE lsif_data_docs_search_current_private (
    repo_id integer NOT NULL,
    dump_root text NOT NULL,
    lang_name_id integer NOT NULL,
    dump_id integer NOT NULL,
    last_cleanup_scan_at timestamp with time zone DEFAULT now() NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    id integer NOT NULL
);

COMMENT ON TABLE lsif_data_docs_search_current_private IS 'A table indicating the most current search index for a unique repository, root, and language.';

COMMENT ON COLUMN lsif_data_docs_search_current_private.repo_id IS 'The repository identifier of the associated dump.';

COMMENT ON COLUMN lsif_data_docs_search_current_private.dump_root IS 'The root of the associated dump.';

COMMENT ON COLUMN lsif_data_docs_search_current_private.lang_name_id IS 'The interned index name of the associated dump.';

COMMENT ON COLUMN lsif_data_docs_search_current_private.dump_id IS 'The associated dump identifier.';

COMMENT ON COLUMN lsif_data_docs_search_current_private.last_cleanup_scan_at IS 'The last time this record was checked as part of a data retention scan.';

COMMENT ON COLUMN lsif_data_docs_search_current_private.created_at IS 'The time this record was inserted. The records with the latest created_at value for the same repository, root, and language is the only visible one and others will be deleted asynchronously.';

CREATE SEQUENCE lsif_data_docs_search_current_private_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE lsif_data_docs_search_current_private_id_seq OWNED BY lsif_data_docs_search_current_private.id;

CREATE TABLE lsif_data_docs_search_current_public (
    repo_id integer NOT NULL,
    dump_root text NOT NULL,
    lang_name_id integer NOT NULL,
    dump_id integer NOT NULL,
    last_cleanup_scan_at timestamp with time zone DEFAULT now() NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    id integer NOT NULL
);

COMMENT ON TABLE lsif_data_docs_search_current_public IS 'A table indicating the most current search index for a unique repository, root, and language.';

COMMENT ON COLUMN lsif_data_docs_search_current_public.repo_id IS 'The repository identifier of the associated dump.';

COMMENT ON COLUMN lsif_data_docs_search_current_public.dump_root IS 'The root of the associated dump.';

COMMENT ON COLUMN lsif_data_docs_search_current_public.lang_name_id IS 'The interned index name of the associated dump.';

COMMENT ON COLUMN lsif_data_docs_search_current_public.dump_id IS 'The associated dump identifier.';

COMMENT ON COLUMN lsif_data_docs_search_current_public.last_cleanup_scan_at IS 'The last time this record was checked as part of a data retention scan.';

COMMENT ON COLUMN lsif_data_docs_search_current_public.created_at IS 'The time this record was inserted. The records with the latest created_at value for the same repository, root, and language is the only visible one and others will be deleted asynchronously.';

CREATE SEQUENCE lsif_data_docs_search_current_public_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE lsif_data_docs_search_current_public_id_seq OWNED BY lsif_data_docs_search_current_public.id;

CREATE TABLE lsif_data_docs_search_lang_names_private (
    id bigint NOT NULL,
    lang_name text NOT NULL,
    tsv tsvector NOT NULL
);

COMMENT ON TABLE lsif_data_docs_search_lang_names_private IS 'Each unique language name being stored in the API docs search index.';

COMMENT ON COLUMN lsif_data_docs_search_lang_names_private.id IS 'The ID of the language name.';

COMMENT ON COLUMN lsif_data_docs_search_lang_names_private.lang_name IS 'The lowercase language name (go, java, etc.) OR, if unknown, the LSIF indexer name.';

COMMENT ON COLUMN lsif_data_docs_search_lang_names_private.tsv IS 'Indexed tsvector for the lang_name field. Crafted for ordered, case, and punctuation sensitivity, see data_write_documentation.go:textSearchVector.';

CREATE SEQUENCE lsif_data_docs_search_lang_names_private_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE lsif_data_docs_search_lang_names_private_id_seq OWNED BY lsif_data_docs_search_lang_names_private.id;

CREATE TABLE lsif_data_docs_search_lang_names_public (
    id bigint NOT NULL,
    lang_name text NOT NULL,
    tsv tsvector NOT NULL
);

COMMENT ON TABLE lsif_data_docs_search_lang_names_public IS 'Each unique language name being stored in the API docs search index.';

COMMENT ON COLUMN lsif_data_docs_search_lang_names_public.id IS 'The ID of the language name.';

COMMENT ON COLUMN lsif_data_docs_search_lang_names_public.lang_name IS 'The lowercase language name (go, java, etc.) OR, if unknown, the LSIF indexer name.';

COMMENT ON COLUMN lsif_data_docs_search_lang_names_public.tsv IS 'Indexed tsvector for the lang_name field. Crafted for ordered, case, and punctuation sensitivity, see data_write_documentation.go:textSearchVector.';

CREATE SEQUENCE lsif_data_docs_search_lang_names_public_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE lsif_data_docs_search_lang_names_public_id_seq OWNED BY lsif_data_docs_search_lang_names_public.id;

CREATE TABLE lsif_data_docs_search_private (
    id bigint NOT NULL,
    repo_id integer NOT NULL,
    dump_id integer NOT NULL,
    dump_root text NOT NULL,
    path_id text NOT NULL,
    detail text NOT NULL,
    lang_name_id integer NOT NULL,
    repo_name_id integer NOT NULL,
    tags_id integer NOT NULL,
    search_key text NOT NULL,
    search_key_tsv tsvector NOT NULL,
    search_key_reverse_tsv tsvector NOT NULL,
    label text NOT NULL,
    label_tsv tsvector NOT NULL,
    label_reverse_tsv tsvector NOT NULL
);

COMMENT ON TABLE lsif_data_docs_search_private IS 'A tsvector search index over API documentation (private repos only)';

COMMENT ON COLUMN lsif_data_docs_search_private.id IS 'The row ID of the search result.';

COMMENT ON COLUMN lsif_data_docs_search_private.repo_id IS 'The repo ID, from the main app DB repo table. Used to search over a select set of repos by ID.';

COMMENT ON COLUMN lsif_data_docs_search_private.dump_id IS 'The identifier of the associated dump in the lsif_uploads table (state=completed).';

COMMENT ON COLUMN lsif_data_docs_search_private.dump_root IS 'Identical to lsif_dumps.root; The working directory of the indexer image relative to the repository root.';

COMMENT ON COLUMN lsif_data_docs_search_private.path_id IS 'The fully qualified documentation page path ID, e.g. including "#section". See GraphQL codeintel.schema:documentationPage for what this is.';

COMMENT ON COLUMN lsif_data_docs_search_private.detail IS 'The detail string (e.g. the full function signature and its docs). See protocol/documentation.go:Documentation';

COMMENT ON COLUMN lsif_data_docs_search_private.lang_name_id IS 'The programming language (or indexer name) that produced the result. Foreign key into lsif_data_docs_search_lang_names_private.';

COMMENT ON COLUMN lsif_data_docs_search_private.repo_name_id IS 'The repository name that produced the result. Foreign key into lsif_data_docs_search_repo_names_private.';

COMMENT ON COLUMN lsif_data_docs_search_private.tags_id IS 'The tags from the documentation node. Foreign key into lsif_data_docs_search_tags_private.';

COMMENT ON COLUMN lsif_data_docs_search_private.search_key IS 'The search key generated by the indexer, e.g. mux.Router.ServeHTTP. It is language-specific, and likely unique within a repository (but not always.) See protocol/documentation.go:Documentation.SearchKey';

COMMENT ON COLUMN lsif_data_docs_search_private.search_key_tsv IS 'Indexed tsvector for the search_key field. Crafted for ordered, case, and punctuation sensitivity, see data_write_documentation.go:textSearchVector.';

COMMENT ON COLUMN lsif_data_docs_search_private.search_key_reverse_tsv IS 'Indexed tsvector for the reverse of the search_key field, for suffix lexeme/word matching. Crafted for ordered, case, and punctuation sensitivity, see data_write_documentation.go:textSearchVector.';

COMMENT ON COLUMN lsif_data_docs_search_private.label IS 'The label string of the result, e.g. a one-line function signature. See protocol/documentation.go:Documentation';

COMMENT ON COLUMN lsif_data_docs_search_private.label_tsv IS 'Indexed tsvector for the label field. Crafted for ordered, case, and punctuation sensitivity, see data_write_documentation.go:textSearchVector.';

COMMENT ON COLUMN lsif_data_docs_search_private.label_reverse_tsv IS 'Indexed tsvector for the reverse of the label field, for suffix lexeme/word matching. Crafted for ordered, case, and punctuation sensitivity, see data_write_documentation.go:textSearchVector.';

CREATE SEQUENCE lsif_data_docs_search_private_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE lsif_data_docs_search_private_id_seq OWNED BY lsif_data_docs_search_private.id;

CREATE TABLE lsif_data_docs_search_public (
    id bigint NOT NULL,
    repo_id integer NOT NULL,
    dump_id integer NOT NULL,
    dump_root text NOT NULL,
    path_id text NOT NULL,
    detail text NOT NULL,
    lang_name_id integer NOT NULL,
    repo_name_id integer NOT NULL,
    tags_id integer NOT NULL,
    search_key text NOT NULL,
    search_key_tsv tsvector NOT NULL,
    search_key_reverse_tsv tsvector NOT NULL,
    label text NOT NULL,
    label_tsv tsvector NOT NULL,
    label_reverse_tsv tsvector NOT NULL
);

COMMENT ON TABLE lsif_data_docs_search_public IS 'A tsvector search index over API documentation (public repos only)';

COMMENT ON COLUMN lsif_data_docs_search_public.id IS 'The row ID of the search result.';

COMMENT ON COLUMN lsif_data_docs_search_public.repo_id IS 'The repo ID, from the main app DB repo table. Used to search over a select set of repos by ID.';

COMMENT ON COLUMN lsif_data_docs_search_public.dump_id IS 'The identifier of the associated dump in the lsif_uploads table (state=completed).';

COMMENT ON COLUMN lsif_data_docs_search_public.dump_root IS 'Identical to lsif_dumps.root; The working directory of the indexer image relative to the repository root.';

COMMENT ON COLUMN lsif_data_docs_search_public.path_id IS 'The fully qualified documentation page path ID, e.g. including "#section". See GraphQL codeintel.schema:documentationPage for what this is.';

COMMENT ON COLUMN lsif_data_docs_search_public.detail IS 'The detail string (e.g. the full function signature and its docs). See protocol/documentation.go:Documentation';

COMMENT ON COLUMN lsif_data_docs_search_public.lang_name_id IS 'The programming language (or indexer name) that produced the result. Foreign key into lsif_data_docs_search_lang_names_public.';

COMMENT ON COLUMN lsif_data_docs_search_public.repo_name_id IS 'The repository name that produced the result. Foreign key into lsif_data_docs_search_repo_names_public.';

COMMENT ON COLUMN lsif_data_docs_search_public.tags_id IS 'The tags from the documentation node. Foreign key into lsif_data_docs_search_tags_public.';

COMMENT ON COLUMN lsif_data_docs_search_public.search_key IS 'The search key generated by the indexer, e.g. mux.Router.ServeHTTP. It is language-specific, and likely unique within a repository (but not always.) See protocol/documentation.go:Documentation.SearchKey';

COMMENT ON COLUMN lsif_data_docs_search_public.search_key_tsv IS 'Indexed tsvector for the search_key field. Crafted for ordered, case, and punctuation sensitivity, see data_write_documentation.go:textSearchVector.';

COMMENT ON COLUMN lsif_data_docs_search_public.search_key_reverse_tsv IS 'Indexed tsvector for the reverse of the search_key field, for suffix lexeme/word matching. Crafted for ordered, case, and punctuation sensitivity, see data_write_documentation.go:textSearchVector.';

COMMENT ON COLUMN lsif_data_docs_search_public.label IS 'The label string of the result, e.g. a one-line function signature. See protocol/documentation.go:Documentation';

COMMENT ON COLUMN lsif_data_docs_search_public.label_tsv IS 'Indexed tsvector for the label field. Crafted for ordered, case, and punctuation sensitivity, see data_write_documentation.go:textSearchVector.';

COMMENT ON COLUMN lsif_data_docs_search_public.label_reverse_tsv IS 'Indexed tsvector for the reverse of the label field, for suffix lexeme/word matching. Crafted for ordered, case, and punctuation sensitivity, see data_write_documentation.go:textSearchVector.';

CREATE SEQUENCE lsif_data_docs_search_public_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE lsif_data_docs_search_public_id_seq OWNED BY lsif_data_docs_search_public.id;

CREATE TABLE lsif_data_docs_search_repo_names_private (
    id bigint NOT NULL,
    repo_name text NOT NULL,
    tsv tsvector NOT NULL,
    reverse_tsv tsvector NOT NULL
);

COMMENT ON TABLE lsif_data_docs_search_repo_names_private IS 'Each unique repository name being stored in the API docs search index.';

COMMENT ON COLUMN lsif_data_docs_search_repo_names_private.id IS 'The ID of the repository name.';

COMMENT ON COLUMN lsif_data_docs_search_repo_names_private.repo_name IS 'The fully qualified name of the repository.';

COMMENT ON COLUMN lsif_data_docs_search_repo_names_private.tsv IS 'Indexed tsvector for the lang_name field. Crafted for ordered, case, and punctuation sensitivity, see data_write_documentation.go:textSearchVector.';

COMMENT ON COLUMN lsif_data_docs_search_repo_names_private.reverse_tsv IS 'Indexed tsvector for the reverse of the lang_name field, for suffix lexeme/word matching. Crafted for ordered, case, and punctuation sensitivity, see data_write_documentation.go:textSearchVector.';

CREATE SEQUENCE lsif_data_docs_search_repo_names_private_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE lsif_data_docs_search_repo_names_private_id_seq OWNED BY lsif_data_docs_search_repo_names_private.id;

CREATE TABLE lsif_data_docs_search_repo_names_public (
    id bigint NOT NULL,
    repo_name text NOT NULL,
    tsv tsvector NOT NULL,
    reverse_tsv tsvector NOT NULL
);

COMMENT ON TABLE lsif_data_docs_search_repo_names_public IS 'Each unique repository name being stored in the API docs search index.';

COMMENT ON COLUMN lsif_data_docs_search_repo_names_public.id IS 'The ID of the repository name.';

COMMENT ON COLUMN lsif_data_docs_search_repo_names_public.repo_name IS 'The fully qualified name of the repository.';

COMMENT ON COLUMN lsif_data_docs_search_repo_names_public.tsv IS 'Indexed tsvector for the lang_name field. Crafted for ordered, case, and punctuation sensitivity, see data_write_documentation.go:textSearchVector.';

COMMENT ON COLUMN lsif_data_docs_search_repo_names_public.reverse_tsv IS 'Indexed tsvector for the reverse of the lang_name field, for suffix lexeme/word matching. Crafted for ordered, case, and punctuation sensitivity, see data_write_documentation.go:textSearchVector.';

CREATE SEQUENCE lsif_data_docs_search_repo_names_public_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE lsif_data_docs_search_repo_names_public_id_seq OWNED BY lsif_data_docs_search_repo_names_public.id;

CREATE TABLE lsif_data_docs_search_tags_private (
    id bigint NOT NULL,
    tags text NOT NULL,
    tsv tsvector NOT NULL
);

COMMENT ON TABLE lsif_data_docs_search_tags_private IS 'Each uniques sequence of space-separated tags being stored in the API docs search index.';

COMMENT ON COLUMN lsif_data_docs_search_tags_private.id IS 'The ID of the tags.';

COMMENT ON COLUMN lsif_data_docs_search_tags_private.tags IS 'The full sequence of space-separated tags. See protocol/documentation.go:Documentation';

COMMENT ON COLUMN lsif_data_docs_search_tags_private.tsv IS 'Indexed tsvector for the tags field. Crafted for ordered, case, and punctuation sensitivity, see data_write_documentation.go:textSearchVector.';

CREATE SEQUENCE lsif_data_docs_search_tags_private_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE lsif_data_docs_search_tags_private_id_seq OWNED BY lsif_data_docs_search_tags_private.id;

CREATE TABLE lsif_data_docs_search_tags_public (
    id bigint NOT NULL,
    tags text NOT NULL,
    tsv tsvector NOT NULL
);

COMMENT ON TABLE lsif_data_docs_search_tags_public IS 'Each uniques sequence of space-separated tags being stored in the API docs search index.';

COMMENT ON COLUMN lsif_data_docs_search_tags_public.id IS 'The ID of the tags.';

COMMENT ON COLUMN lsif_data_docs_search_tags_public.tags IS 'The full sequence of space-separated tags. See protocol/documentation.go:Documentation';

COMMENT ON COLUMN lsif_data_docs_search_tags_public.tsv IS 'Indexed tsvector for the tags field. Crafted for ordered, case, and punctuation sensitivity, see data_write_documentation.go:textSearchVector.';

CREATE SEQUENCE lsif_data_docs_search_tags_public_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE lsif_data_docs_search_tags_public_id_seq OWNED BY lsif_data_docs_search_tags_public.id;

CREATE TABLE lsif_data_documentation_mappings (
    dump_id integer NOT NULL,
    path_id text NOT NULL,
    result_id integer NOT NULL,
    file_path text
);

COMMENT ON TABLE lsif_data_documentation_mappings IS 'Maps documentation path IDs to their corresponding integral documentationResult vertex IDs, which are unique within a dump.';

COMMENT ON COLUMN lsif_data_documentation_mappings.dump_id IS 'The identifier of the associated dump in the lsif_uploads table (state=completed).';

COMMENT ON COLUMN lsif_data_documentation_mappings.path_id IS 'The documentation page path ID, see see GraphQL codeintel.schema:documentationPage for what this is.';

COMMENT ON COLUMN lsif_data_documentation_mappings.result_id IS 'The documentationResult vertex ID.';

COMMENT ON COLUMN lsif_data_documentation_mappings.file_path IS 'The document file path for the documentationResult, if any. e.g. the path to the file where the symbol described by this documentationResult is located, if it is a symbol.';

CREATE TABLE lsif_data_documentation_pages (
    dump_id integer NOT NULL,
    path_id text NOT NULL,
    data bytea,
    search_indexed boolean DEFAULT false
);

COMMENT ON TABLE lsif_data_documentation_pages IS 'Associates documentation pathIDs to their documentation page hierarchy chunk.';

COMMENT ON COLUMN lsif_data_documentation_pages.dump_id IS 'The identifier of the associated dump in the lsif_uploads table (state=completed).';

COMMENT ON COLUMN lsif_data_documentation_pages.path_id IS 'The documentation page path ID, see see GraphQL codeintel.schema:documentationPage for what this is.';

COMMENT ON COLUMN lsif_data_documentation_pages.data IS 'A gob-encoded payload conforming to a `type DocumentationPageData struct` pointer (lib/codeintel/semantic/types.go)';

CREATE TABLE lsif_data_documentation_path_info (
    dump_id integer NOT NULL,
    path_id text NOT NULL,
    data bytea
);

COMMENT ON TABLE lsif_data_documentation_path_info IS 'Associates documentation page pathIDs to information about what is at that pathID, its immediate children, etc.';

COMMENT ON COLUMN lsif_data_documentation_path_info.dump_id IS 'The identifier of the associated dump in the lsif_uploads table (state=completed).';

COMMENT ON COLUMN lsif_data_documentation_path_info.path_id IS 'The documentation page path ID, see see GraphQL codeintel.schema:documentationPage for what this is.';

COMMENT ON COLUMN lsif_data_documentation_path_info.data IS 'A gob-encoded payload conforming to a `type DocumentationPathInoData struct` pointer (lib/codeintel/semantic/types.go)';

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

ALTER TABLE ONLY lsif_data_docs_search_current_private ALTER COLUMN id SET DEFAULT nextval('lsif_data_docs_search_current_private_id_seq'::regclass);

ALTER TABLE ONLY lsif_data_docs_search_current_public ALTER COLUMN id SET DEFAULT nextval('lsif_data_docs_search_current_public_id_seq'::regclass);

ALTER TABLE ONLY lsif_data_docs_search_lang_names_private ALTER COLUMN id SET DEFAULT nextval('lsif_data_docs_search_lang_names_private_id_seq'::regclass);

ALTER TABLE ONLY lsif_data_docs_search_lang_names_public ALTER COLUMN id SET DEFAULT nextval('lsif_data_docs_search_lang_names_public_id_seq'::regclass);

ALTER TABLE ONLY lsif_data_docs_search_private ALTER COLUMN id SET DEFAULT nextval('lsif_data_docs_search_private_id_seq'::regclass);

ALTER TABLE ONLY lsif_data_docs_search_public ALTER COLUMN id SET DEFAULT nextval('lsif_data_docs_search_public_id_seq'::regclass);

ALTER TABLE ONLY lsif_data_docs_search_repo_names_private ALTER COLUMN id SET DEFAULT nextval('lsif_data_docs_search_repo_names_private_id_seq'::regclass);

ALTER TABLE ONLY lsif_data_docs_search_repo_names_public ALTER COLUMN id SET DEFAULT nextval('lsif_data_docs_search_repo_names_public_id_seq'::regclass);

ALTER TABLE ONLY lsif_data_docs_search_tags_private ALTER COLUMN id SET DEFAULT nextval('lsif_data_docs_search_tags_private_id_seq'::regclass);

ALTER TABLE ONLY lsif_data_docs_search_tags_public ALTER COLUMN id SET DEFAULT nextval('lsif_data_docs_search_tags_public_id_seq'::regclass);

ALTER TABLE ONLY rockskip_ancestry ALTER COLUMN id SET DEFAULT nextval('rockskip_ancestry_id_seq'::regclass);

ALTER TABLE ONLY rockskip_repos ALTER COLUMN id SET DEFAULT nextval('rockskip_repos_id_seq'::regclass);

ALTER TABLE ONLY rockskip_symbols ALTER COLUMN id SET DEFAULT nextval('rockskip_symbols_id_seq'::regclass);

ALTER TABLE ONLY lsif_data_definitions
    ADD CONSTRAINT lsif_data_definitions_pkey PRIMARY KEY (dump_id, scheme, identifier);

ALTER TABLE ONLY lsif_data_definitions_schema_versions
    ADD CONSTRAINT lsif_data_definitions_schema_versions_pkey PRIMARY KEY (dump_id);

ALTER TABLE ONLY lsif_data_docs_search_current_private
    ADD CONSTRAINT lsif_data_docs_search_current_private_pkey PRIMARY KEY (id);

ALTER TABLE ONLY lsif_data_docs_search_current_public
    ADD CONSTRAINT lsif_data_docs_search_current_public_pkey PRIMARY KEY (id);

ALTER TABLE ONLY lsif_data_docs_search_lang_names_private
    ADD CONSTRAINT lsif_data_docs_search_lang_names_private_lang_name_key UNIQUE (lang_name);

ALTER TABLE ONLY lsif_data_docs_search_lang_names_private
    ADD CONSTRAINT lsif_data_docs_search_lang_names_private_pkey PRIMARY KEY (id);

ALTER TABLE ONLY lsif_data_docs_search_lang_names_public
    ADD CONSTRAINT lsif_data_docs_search_lang_names_public_lang_name_key UNIQUE (lang_name);

ALTER TABLE ONLY lsif_data_docs_search_lang_names_public
    ADD CONSTRAINT lsif_data_docs_search_lang_names_public_pkey PRIMARY KEY (id);

ALTER TABLE ONLY lsif_data_docs_search_private
    ADD CONSTRAINT lsif_data_docs_search_private_pkey PRIMARY KEY (id);

ALTER TABLE ONLY lsif_data_docs_search_public
    ADD CONSTRAINT lsif_data_docs_search_public_pkey PRIMARY KEY (id);

ALTER TABLE ONLY lsif_data_docs_search_repo_names_private
    ADD CONSTRAINT lsif_data_docs_search_repo_names_private_pkey PRIMARY KEY (id);

ALTER TABLE ONLY lsif_data_docs_search_repo_names_private
    ADD CONSTRAINT lsif_data_docs_search_repo_names_private_repo_name_key UNIQUE (repo_name);

ALTER TABLE ONLY lsif_data_docs_search_repo_names_public
    ADD CONSTRAINT lsif_data_docs_search_repo_names_public_pkey PRIMARY KEY (id);

ALTER TABLE ONLY lsif_data_docs_search_repo_names_public
    ADD CONSTRAINT lsif_data_docs_search_repo_names_public_repo_name_key UNIQUE (repo_name);

ALTER TABLE ONLY lsif_data_docs_search_tags_private
    ADD CONSTRAINT lsif_data_docs_search_tags_private_pkey PRIMARY KEY (id);

ALTER TABLE ONLY lsif_data_docs_search_tags_private
    ADD CONSTRAINT lsif_data_docs_search_tags_private_tags_key UNIQUE (tags);

ALTER TABLE ONLY lsif_data_docs_search_tags_public
    ADD CONSTRAINT lsif_data_docs_search_tags_public_pkey PRIMARY KEY (id);

ALTER TABLE ONLY lsif_data_docs_search_tags_public
    ADD CONSTRAINT lsif_data_docs_search_tags_public_tags_key UNIQUE (tags);

ALTER TABLE ONLY lsif_data_documentation_mappings
    ADD CONSTRAINT lsif_data_documentation_mappings_pkey PRIMARY KEY (dump_id, path_id);

ALTER TABLE ONLY lsif_data_documentation_pages
    ADD CONSTRAINT lsif_data_documentation_pages_pkey PRIMARY KEY (dump_id, path_id);

ALTER TABLE ONLY lsif_data_documentation_path_info
    ADD CONSTRAINT lsif_data_documentation_path_info_pkey PRIMARY KEY (dump_id, path_id);

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

CREATE INDEX lsif_data_docs_search_current_private_last_cleanup_scan_at ON lsif_data_docs_search_current_private USING btree (last_cleanup_scan_at);

CREATE INDEX lsif_data_docs_search_current_private_lookup ON lsif_data_docs_search_current_private USING btree (repo_id, dump_root, lang_name_id, created_at) INCLUDE (dump_id);

CREATE INDEX lsif_data_docs_search_current_public_last_cleanup_scan_at ON lsif_data_docs_search_current_public USING btree (last_cleanup_scan_at);

CREATE INDEX lsif_data_docs_search_current_public_lookup ON lsif_data_docs_search_current_public USING btree (repo_id, dump_root, lang_name_id, created_at) INCLUDE (dump_id);

CREATE INDEX lsif_data_docs_search_lang_names_private_tsv_idx ON lsif_data_docs_search_lang_names_private USING gin (tsv);

CREATE INDEX lsif_data_docs_search_lang_names_public_tsv_idx ON lsif_data_docs_search_lang_names_public USING gin (tsv);

CREATE INDEX lsif_data_docs_search_private_dump_id_idx ON lsif_data_docs_search_private USING btree (dump_id);

CREATE INDEX lsif_data_docs_search_private_dump_root_idx ON lsif_data_docs_search_private USING btree (dump_root);

CREATE INDEX lsif_data_docs_search_private_label_reverse_tsv_idx ON lsif_data_docs_search_private USING gin (label_reverse_tsv);

CREATE INDEX lsif_data_docs_search_private_label_tsv_idx ON lsif_data_docs_search_private USING gin (label_tsv);

CREATE INDEX lsif_data_docs_search_private_repo_id_idx ON lsif_data_docs_search_private USING btree (repo_id);

CREATE INDEX lsif_data_docs_search_private_search_key_reverse_tsv_idx ON lsif_data_docs_search_private USING gin (search_key_reverse_tsv);

CREATE INDEX lsif_data_docs_search_private_search_key_tsv_idx ON lsif_data_docs_search_private USING gin (search_key_tsv);

CREATE INDEX lsif_data_docs_search_public_dump_id_idx ON lsif_data_docs_search_public USING btree (dump_id);

CREATE INDEX lsif_data_docs_search_public_dump_root_idx ON lsif_data_docs_search_public USING btree (dump_root);

CREATE INDEX lsif_data_docs_search_public_label_reverse_tsv_idx ON lsif_data_docs_search_public USING gin (label_reverse_tsv);

CREATE INDEX lsif_data_docs_search_public_label_tsv_idx ON lsif_data_docs_search_public USING gin (label_tsv);

CREATE INDEX lsif_data_docs_search_public_repo_id_idx ON lsif_data_docs_search_public USING btree (repo_id);

CREATE INDEX lsif_data_docs_search_public_search_key_reverse_tsv_idx ON lsif_data_docs_search_public USING gin (search_key_reverse_tsv);

CREATE INDEX lsif_data_docs_search_public_search_key_tsv_idx ON lsif_data_docs_search_public USING gin (search_key_tsv);

CREATE INDEX lsif_data_docs_search_repo_names_private_reverse_tsv_idx ON lsif_data_docs_search_repo_names_private USING gin (reverse_tsv);

CREATE INDEX lsif_data_docs_search_repo_names_private_tsv_idx ON lsif_data_docs_search_repo_names_private USING gin (tsv);

CREATE INDEX lsif_data_docs_search_repo_names_public_reverse_tsv_idx ON lsif_data_docs_search_repo_names_public USING gin (reverse_tsv);

CREATE INDEX lsif_data_docs_search_repo_names_public_tsv_idx ON lsif_data_docs_search_repo_names_public USING gin (tsv);

CREATE INDEX lsif_data_docs_search_tags_private_tsv_idx ON lsif_data_docs_search_tags_private USING gin (tsv);

CREATE INDEX lsif_data_docs_search_tags_public_tsv_idx ON lsif_data_docs_search_tags_public USING gin (tsv);

CREATE UNIQUE INDEX lsif_data_documentation_mappings_inverse_unique_idx ON lsif_data_documentation_mappings USING btree (dump_id, result_id);

CREATE INDEX lsif_data_documentation_pages_dump_id_unindexed ON lsif_data_documentation_pages USING btree (dump_id) WHERE (NOT search_indexed);

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

CREATE TRIGGER lsif_data_docs_search_private_delete AFTER DELETE ON lsif_data_docs_search_private REFERENCING OLD TABLE AS oldtbl FOR EACH STATEMENT EXECUTE FUNCTION lsif_data_docs_search_private_delete();

CREATE TRIGGER lsif_data_docs_search_private_insert AFTER INSERT ON lsif_data_docs_search_private REFERENCING NEW TABLE AS newtbl FOR EACH STATEMENT EXECUTE FUNCTION lsif_data_docs_search_private_insert();

CREATE TRIGGER lsif_data_docs_search_public_delete AFTER DELETE ON lsif_data_docs_search_public REFERENCING OLD TABLE AS oldtbl FOR EACH STATEMENT EXECUTE FUNCTION lsif_data_docs_search_public_delete();

CREATE TRIGGER lsif_data_docs_search_public_insert AFTER INSERT ON lsif_data_docs_search_public REFERENCING NEW TABLE AS newtbl FOR EACH STATEMENT EXECUTE FUNCTION lsif_data_docs_search_public_insert();

CREATE TRIGGER lsif_data_documentation_pages_delete AFTER DELETE ON lsif_data_documentation_pages REFERENCING OLD TABLE AS oldtbl FOR EACH STATEMENT EXECUTE FUNCTION lsif_data_documentation_pages_delete();

CREATE TRIGGER lsif_data_documentation_pages_insert AFTER INSERT ON lsif_data_documentation_pages REFERENCING NEW TABLE AS newtbl FOR EACH STATEMENT EXECUTE FUNCTION lsif_data_documentation_pages_insert();

CREATE TRIGGER lsif_data_documentation_pages_update AFTER UPDATE ON lsif_data_documentation_pages REFERENCING OLD TABLE AS oldtbl NEW TABLE AS newtbl FOR EACH STATEMENT EXECUTE FUNCTION lsif_data_documentation_pages_update();

CREATE TRIGGER lsif_data_documents_schema_versions_insert AFTER INSERT ON lsif_data_documents REFERENCING NEW TABLE AS newtab FOR EACH STATEMENT EXECUTE FUNCTION update_lsif_data_documents_schema_versions_insert();

CREATE TRIGGER lsif_data_implementations_schema_versions_insert AFTER INSERT ON lsif_data_implementations REFERENCING NEW TABLE AS newtab FOR EACH STATEMENT EXECUTE FUNCTION update_lsif_data_implementations_schema_versions_insert();

CREATE TRIGGER lsif_data_references_schema_versions_insert AFTER INSERT ON lsif_data_references REFERENCING NEW TABLE AS newtab FOR EACH STATEMENT EXECUTE FUNCTION update_lsif_data_references_schema_versions_insert();

ALTER TABLE ONLY lsif_data_docs_search_private
    ADD CONSTRAINT lsif_data_docs_search_private_lang_name_id_fk FOREIGN KEY (lang_name_id) REFERENCES lsif_data_docs_search_lang_names_private(id);

ALTER TABLE ONLY lsif_data_docs_search_private
    ADD CONSTRAINT lsif_data_docs_search_private_repo_name_id_fk FOREIGN KEY (repo_name_id) REFERENCES lsif_data_docs_search_repo_names_private(id);

ALTER TABLE ONLY lsif_data_docs_search_private
    ADD CONSTRAINT lsif_data_docs_search_private_tags_id_fk FOREIGN KEY (tags_id) REFERENCES lsif_data_docs_search_tags_private(id);

ALTER TABLE ONLY lsif_data_docs_search_public
    ADD CONSTRAINT lsif_data_docs_search_public_lang_name_id_fk FOREIGN KEY (lang_name_id) REFERENCES lsif_data_docs_search_lang_names_public(id);

ALTER TABLE ONLY lsif_data_docs_search_public
    ADD CONSTRAINT lsif_data_docs_search_public_repo_name_id_fk FOREIGN KEY (repo_name_id) REFERENCES lsif_data_docs_search_repo_names_public(id);

ALTER TABLE ONLY lsif_data_docs_search_public
    ADD CONSTRAINT lsif_data_docs_search_public_tags_id_fk FOREIGN KEY (tags_id) REFERENCES lsif_data_docs_search_tags_public(id);

INSERT INTO lsif_data_apidocs_num_dumps VALUES (0);
INSERT INTO lsif_data_apidocs_num_dumps_indexed VALUES (0);
INSERT INTO lsif_data_apidocs_num_pages VALUES (0);
INSERT INTO lsif_data_apidocs_num_search_results_private VALUES (0);
INSERT INTO lsif_data_apidocs_num_search_results_public VALUES (0);