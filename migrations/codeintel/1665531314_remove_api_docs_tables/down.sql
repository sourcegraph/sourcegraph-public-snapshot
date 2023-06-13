CREATE OR REPLACE FUNCTION lsif_data_docs_search_private_delete() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
UPDATE lsif_data_apidocs_num_search_results_private SET count = count - (select count(*) from oldtbl);
RETURN NULL;
END $$;

CREATE OR REPLACE FUNCTION lsif_data_docs_search_private_insert() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
UPDATE lsif_data_apidocs_num_search_results_private SET count = count + (select count(*) from newtbl);
RETURN NULL;
END $$;

CREATE OR REPLACE FUNCTION lsif_data_docs_search_public_delete() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
UPDATE lsif_data_apidocs_num_search_results_public SET count = count - (select count(*) from oldtbl);
RETURN NULL;
END $$;

CREATE OR REPLACE FUNCTION lsif_data_docs_search_public_insert() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
UPDATE lsif_data_apidocs_num_search_results_public SET count = count + (select count(*) from newtbl);
RETURN NULL;
END $$;

CREATE OR REPLACE FUNCTION lsif_data_documentation_pages_delete() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
UPDATE lsif_data_apidocs_num_pages SET count = count - (select count(*) from oldtbl);
UPDATE lsif_data_apidocs_num_dumps SET count = count - (select count(DISTINCT dump_id) from oldtbl);
UPDATE lsif_data_apidocs_num_dumps_indexed SET count = count - (select count(DISTINCT dump_id) from oldtbl WHERE search_indexed='true');
RETURN NULL;
END $$;

CREATE OR REPLACE FUNCTION lsif_data_documentation_pages_insert() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
UPDATE lsif_data_apidocs_num_pages SET count = count + (select count(*) from newtbl);
UPDATE lsif_data_apidocs_num_dumps SET count = count + (select count(DISTINCT dump_id) from newtbl);
UPDATE lsif_data_apidocs_num_dumps_indexed SET count = count + (select count(DISTINCT dump_id) from newtbl WHERE search_indexed='true');
RETURN NULL;
END $$;

CREATE OR REPLACE FUNCTION lsif_data_documentation_pages_update() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
WITH
    beforeIndexed AS (SELECT count(DISTINCT dump_id) FROM oldtbl WHERE search_indexed='true'),
    afterIndexed AS (SELECT count(DISTINCT dump_id) FROM newtbl WHERE search_indexed='true')
UPDATE lsif_data_apidocs_num_dumps_indexed SET count=count + ((select * from afterIndexed) - (select * from beforeIndexed));
RETURN NULL;
END $$;

CREATE TABLE IF NOT EXISTS lsif_data_apidocs_num_dumps (count bigint);
CREATE TABLE IF NOT EXISTS lsif_data_apidocs_num_dumps_indexed (count bigint);
CREATE TABLE IF NOT EXISTS lsif_data_apidocs_num_pages (count bigint);
CREATE TABLE IF NOT EXISTS lsif_data_apidocs_num_search_results_private (count bigint);
CREATE TABLE IF NOT EXISTS lsif_data_apidocs_num_search_results_public (count bigint);

INSERT INTO lsif_data_apidocs_num_dumps VALUES (0);
INSERT INTO lsif_data_apidocs_num_dumps_indexed VALUES (0);
INSERT INTO lsif_data_apidocs_num_pages VALUES (0);
INSERT INTO lsif_data_apidocs_num_search_results_private VALUES (0);
INSERT INTO lsif_data_apidocs_num_search_results_public VALUES (0);

CREATE SEQUENCE IF NOT EXISTS lsif_data_docs_search_current_private_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

CREATE TABLE IF NOT EXISTS lsif_data_docs_search_current_private (
    repo_id integer NOT NULL,
    dump_root text NOT NULL,
    lang_name_id integer NOT NULL,
    dump_id integer NOT NULL,
    last_cleanup_scan_at timestamp with time zone DEFAULT now() NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    id int NOT NULL DEFAULT nextval('lsif_data_docs_search_current_private_id_seq'),
    PRIMARY KEY(id)
);

CREATE SEQUENCE IF NOT EXISTS lsif_data_docs_search_current_public_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

CREATE TABLE IF NOT EXISTS lsif_data_docs_search_current_public (
    repo_id integer NOT NULL,
    dump_root text NOT NULL,
    lang_name_id integer NOT NULL,
    dump_id integer NOT NULL,
    last_cleanup_scan_at timestamp with time zone DEFAULT now() NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    id int NULL DEFAULT nextval('lsif_data_docs_search_current_public_id_seq'),
    PRIMARY KEY(id)
);

CREATE SEQUENCE IF NOT EXISTS lsif_data_docs_search_lang_names_private_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

CREATE TABLE IF NOT EXISTS lsif_data_docs_search_lang_names_private (
    id BIGINT NOT NULL DEFAULT nextval('lsif_data_docs_search_lang_names_private_id_seq'),
    lang_name text NOT NULL,
    tsv tsvector NOT NULL,
    PRIMARY KEY(id),
    UNIQUE(lang_name)
);

ALTER SEQUENCE lsif_data_docs_search_lang_names_private_id_seq OWNED BY lsif_data_docs_search_lang_names_private.id;

CREATE SEQUENCE IF NOT EXISTS lsif_data_docs_search_lang_names_public_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

CREATE TABLE IF NOT EXISTS lsif_data_docs_search_lang_names_public (
    id BIGINT NOT NULL DEFAULT nextval('lsif_data_docs_search_lang_names_public_id_seq'),
    lang_name text NOT NULL,
    tsv tsvector NOT NULL,
    PRIMARY KEY(id),
    UNIQUE(lang_name)
);

ALTER SEQUENCE lsif_data_docs_search_lang_names_public_id_seq OWNED BY lsif_data_docs_search_lang_names_public.id;

CREATE SEQUENCE IF NOT EXISTS lsif_data_docs_search_private_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

CREATE TABLE IF NOT EXISTS lsif_data_docs_search_private (
    id BIGINT NOT NULL DEFAULT nextval('lsif_data_docs_search_private_id_seq'),
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
    label_reverse_tsv tsvector NOT NULL,
    PRIMARY KEY(id)
);

ALTER TABLE lsif_data_docs_search_private
    DROP CONSTRAINT IF EXISTS lsif_data_docs_search_private_lang_name_id_fk;
ALTER TABLE lsif_data_docs_search_private ADD CONSTRAINT lsif_data_docs_search_private_lang_name_id_fk FOREIGN KEY (lang_name_id) REFERENCES lsif_data_docs_search_lang_names_private(id);

ALTER SEQUENCE lsif_data_docs_search_private_id_seq OWNED BY lsif_data_docs_search_private.id;

CREATE SEQUENCE IF NOT EXISTS lsif_data_docs_search_public_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

CREATE TABLE IF NOT EXISTS lsif_data_docs_search_public (
    id BIGINT NOT NULL DEFAULT nextval('lsif_data_docs_search_public_id_seq'),
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
    label_reverse_tsv tsvector NOT NULL,
    PRIMARY KEY(id)
);

ALTER TABLE lsif_data_docs_search_public
    DROP CONSTRAINT IF EXISTS lsif_data_docs_search_public_lang_name_id_fk;
ALTER TABLE lsif_data_docs_search_public ADD CONSTRAINT lsif_data_docs_search_public_lang_name_id_fk FOREIGN KEY (lang_name_id) REFERENCES lsif_data_docs_search_lang_names_public(id);

ALTER SEQUENCE lsif_data_docs_search_public_id_seq OWNED BY lsif_data_docs_search_public.id;

CREATE SEQUENCE IF NOT EXISTS lsif_data_docs_search_repo_names_private_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

CREATE TABLE IF NOT EXISTS lsif_data_docs_search_repo_names_private (
    id BIGINT NOT NULL DEFAULT nextval('lsif_data_docs_search_repo_names_private_id_seq'),
    repo_name text NOT NULL,
    tsv tsvector NOT NULL,
    reverse_tsv tsvector NOT NULL,
    PRIMARY KEY(id),
    UNIQUE(repo_name)
);

ALTER TABLE lsif_data_docs_search_private
    DROP CONSTRAINT IF EXISTS lsif_data_docs_search_private_repo_name_id_fk;
ALTER TABLE lsif_data_docs_search_private ADD CONSTRAINT lsif_data_docs_search_private_repo_name_id_fk FOREIGN KEY (repo_name_id) REFERENCES lsif_data_docs_search_repo_names_private(id);

ALTER SEQUENCE lsif_data_docs_search_repo_names_private_id_seq OWNED BY lsif_data_docs_search_repo_names_private.id;

CREATE SEQUENCE IF NOT EXISTS lsif_data_docs_search_repo_names_public_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

CREATE TABLE IF NOT EXISTS lsif_data_docs_search_repo_names_public (
    id BIGINT NOT NULL DEFAULT nextval('lsif_data_docs_search_repo_names_public_id_seq'),
    repo_name text NOT NULL,
    tsv tsvector NOT NULL,
    reverse_tsv tsvector NOT NULL,
    PRIMARY KEY(id),
    UNIQUE(repo_name)
);
ALTER TABLE lsif_data_docs_search_public
    DROP CONSTRAINT IF EXISTS lsif_data_docs_search_public_repo_name_id_fk;
ALTER TABLE lsif_data_docs_search_public ADD CONSTRAINT lsif_data_docs_search_public_repo_name_id_fk FOREIGN KEY (repo_name_id) REFERENCES lsif_data_docs_search_repo_names_public(id);

ALTER SEQUENCE lsif_data_docs_search_repo_names_public_id_seq OWNED BY lsif_data_docs_search_repo_names_public.id;

CREATE SEQUENCE IF NOT EXISTS lsif_data_docs_search_tags_private_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

CREATE TABLE IF NOT EXISTS lsif_data_docs_search_tags_private (
    id BIGINT NOT NULL DEFAULT nextval('lsif_data_docs_search_tags_private_id_seq'),
    tags text NOT NULL UNIQUE,
    tsv tsvector NOT NULL,
    PRIMARY KEY(id)
);
ALTER TABLE lsif_data_docs_search_private DROP CONSTRAINT IF EXISTS lsif_data_docs_search_private_tags_id_fk;
ALTER TABLE lsif_data_docs_search_private ADD CONSTRAINT lsif_data_docs_search_private_tags_id_fk FOREIGN KEY (tags_id) REFERENCES lsif_data_docs_search_tags_private(id);

ALTER SEQUENCE lsif_data_docs_search_tags_private_id_seq OWNED BY lsif_data_docs_search_tags_private.id;

CREATE SEQUENCE IF NOT EXISTS lsif_data_docs_search_tags_public_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

CREATE TABLE IF NOT EXISTS lsif_data_docs_search_tags_public (
    id BIGINT NOT NULL DEFAULT nextval('lsif_data_docs_search_tags_public_id_seq'),
    tags text NOT NULL UNIQUE,
    tsv tsvector NOT NULL,
    PRIMARY KEY(id),
    UNIQUE(tags)
);
ALTER TABLE lsif_data_docs_search_public DROP CONSTRAINT IF EXISTS lsif_data_docs_search_public_tags_id_fk;
ALTER TABLE lsif_data_docs_search_public ADD CONSTRAINT lsif_data_docs_search_public_tags_id_fk FOREIGN KEY (tags_id) REFERENCES lsif_data_docs_search_tags_public(id);

ALTER SEQUENCE lsif_data_docs_search_tags_public_id_seq OWNED BY lsif_data_docs_search_tags_public.id;

CREATE TABLE IF NOT EXISTS lsif_data_documentation_mappings (
    dump_id integer NOT NULL,
    path_id text NOT NULL,
    result_id integer NOT NULL,
    file_path text,
    PRIMARY KEY(dump_id, path_id)
);

CREATE TABLE IF NOT EXISTS lsif_data_documentation_pages (
    dump_id integer NOT NULL,
    path_id text NOT NULL,
    data bytea,
    search_indexed boolean DEFAULT false,
    PRIMARY KEY(dump_id, path_id)
);

CREATE TABLE IF NOT EXISTS lsif_data_documentation_path_info (
    dump_id integer NOT NULL,
    path_id text NOT NULL,
    data bytea,
    PRIMARY KEY(dump_id, path_id)
);

CREATE INDEX IF NOT EXISTS lsif_data_docs_search_current_private_last_cleanup_scan_at ON lsif_data_docs_search_current_private USING btree (last_cleanup_scan_at);
CREATE INDEX IF NOT EXISTS lsif_data_docs_search_current_private_lookup ON lsif_data_docs_search_current_private USING btree (repo_id, dump_root, lang_name_id, created_at) INCLUDE (dump_id);
CREATE INDEX IF NOT EXISTS lsif_data_docs_search_current_public_last_cleanup_scan_at ON lsif_data_docs_search_current_public USING btree (last_cleanup_scan_at);
CREATE INDEX IF NOT EXISTS lsif_data_docs_search_current_public_lookup ON lsif_data_docs_search_current_public USING btree (repo_id, dump_root, lang_name_id, created_at) INCLUDE (dump_id);
CREATE INDEX IF NOT EXISTS lsif_data_docs_search_lang_names_private_tsv_idx ON lsif_data_docs_search_lang_names_private USING gin (tsv);
CREATE INDEX IF NOT EXISTS lsif_data_docs_search_lang_names_public_tsv_idx ON lsif_data_docs_search_lang_names_public USING gin (tsv);
CREATE INDEX IF NOT EXISTS lsif_data_docs_search_private_dump_id_idx ON lsif_data_docs_search_private USING btree (dump_id);
CREATE INDEX IF NOT EXISTS lsif_data_docs_search_private_dump_root_idx ON lsif_data_docs_search_private USING btree (dump_root);
CREATE INDEX IF NOT EXISTS lsif_data_docs_search_private_label_reverse_tsv_idx ON lsif_data_docs_search_private USING gin (label_reverse_tsv);
CREATE INDEX IF NOT EXISTS lsif_data_docs_search_private_label_tsv_idx ON lsif_data_docs_search_private USING gin (label_tsv);
CREATE INDEX IF NOT EXISTS lsif_data_docs_search_private_repo_id_idx ON lsif_data_docs_search_private USING btree (repo_id);
CREATE INDEX IF NOT EXISTS lsif_data_docs_search_private_search_key_reverse_tsv_idx ON lsif_data_docs_search_private USING gin (search_key_reverse_tsv);
CREATE INDEX IF NOT EXISTS lsif_data_docs_search_private_search_key_tsv_idx ON lsif_data_docs_search_private USING gin (search_key_tsv);
CREATE INDEX IF NOT EXISTS lsif_data_docs_search_public_dump_id_idx ON lsif_data_docs_search_public USING btree (dump_id);
CREATE INDEX IF NOT EXISTS lsif_data_docs_search_public_dump_root_idx ON lsif_data_docs_search_public USING btree (dump_root);
CREATE INDEX IF NOT EXISTS lsif_data_docs_search_public_label_reverse_tsv_idx ON lsif_data_docs_search_public USING gin (label_reverse_tsv);
CREATE INDEX IF NOT EXISTS lsif_data_docs_search_public_label_tsv_idx ON lsif_data_docs_search_public USING gin (label_tsv);
CREATE INDEX IF NOT EXISTS lsif_data_docs_search_public_repo_id_idx ON lsif_data_docs_search_public USING btree (repo_id);
CREATE INDEX IF NOT EXISTS lsif_data_docs_search_public_search_key_reverse_tsv_idx ON lsif_data_docs_search_public USING gin (search_key_reverse_tsv);
CREATE INDEX IF NOT EXISTS lsif_data_docs_search_public_search_key_tsv_idx ON lsif_data_docs_search_public USING gin (search_key_tsv);
CREATE INDEX IF NOT EXISTS lsif_data_docs_search_repo_names_private_reverse_tsv_idx ON lsif_data_docs_search_repo_names_private USING gin (reverse_tsv);
CREATE INDEX IF NOT EXISTS lsif_data_docs_search_repo_names_private_tsv_idx ON lsif_data_docs_search_repo_names_private USING gin (tsv);
CREATE INDEX IF NOT EXISTS lsif_data_docs_search_repo_names_public_reverse_tsv_idx ON lsif_data_docs_search_repo_names_public USING gin (reverse_tsv);
CREATE INDEX IF NOT EXISTS lsif_data_docs_search_repo_names_public_tsv_idx ON lsif_data_docs_search_repo_names_public USING gin (tsv);
CREATE INDEX IF NOT EXISTS lsif_data_docs_search_tags_private_tsv_idx ON lsif_data_docs_search_tags_private USING gin (tsv);
CREATE INDEX IF NOT EXISTS lsif_data_docs_search_tags_public_tsv_idx ON lsif_data_docs_search_tags_public USING gin (tsv);
CREATE UNIQUE INDEX IF NOT EXISTS lsif_data_documentation_mappings_inverse_unique_idx ON lsif_data_documentation_mappings USING btree (dump_id, result_id);
CREATE INDEX IF NOT EXISTS lsif_data_documentation_pages_dump_id_unindexed ON lsif_data_documentation_pages USING btree (dump_id) WHERE (NOT search_indexed);

DROP TRIGGER IF EXISTS lsif_data_docs_search_private_delete ON lsif_data_docs_search_private;
CREATE TRIGGER lsif_data_docs_search_private_delete AFTER DELETE ON lsif_data_docs_search_private REFERENCING OLD TABLE AS oldtbl FOR EACH STATEMENT EXECUTE FUNCTION lsif_data_docs_search_private_delete();
DROP TRIGGER IF EXISTS lsif_data_docs_search_private_insert ON lsif_data_docs_search_private;
CREATE TRIGGER lsif_data_docs_search_private_insert AFTER INSERT ON lsif_data_docs_search_private REFERENCING NEW TABLE AS newtbl FOR EACH STATEMENT EXECUTE FUNCTION lsif_data_docs_search_private_insert();
DROP TRIGGER IF EXISTS lsif_data_docs_search_public_delete ON lsif_data_docs_search_public;
CREATE TRIGGER lsif_data_docs_search_public_delete AFTER DELETE ON lsif_data_docs_search_public REFERENCING OLD TABLE AS oldtbl FOR EACH STATEMENT EXECUTE FUNCTION lsif_data_docs_search_public_delete();
DROP TRIGGER IF EXISTS lsif_data_docs_search_public_insert ON lsif_data_docs_search_public;
CREATE TRIGGER lsif_data_docs_search_public_insert AFTER INSERT ON lsif_data_docs_search_public REFERENCING NEW TABLE AS newtbl FOR EACH STATEMENT EXECUTE FUNCTION lsif_data_docs_search_public_insert();
DROP TRIGGER IF EXISTS lsif_data_documentation_pages_delete ON lsif_data_documentation_pages;
CREATE TRIGGER lsif_data_documentation_pages_delete AFTER DELETE ON lsif_data_documentation_pages REFERENCING OLD TABLE AS oldtbl FOR EACH STATEMENT EXECUTE FUNCTION lsif_data_documentation_pages_delete();
DROP TRIGGER IF EXISTS lsif_data_documentation_pages_insert ON lsif_data_documentation_pages;
CREATE TRIGGER lsif_data_documentation_pages_insert AFTER INSERT ON lsif_data_documentation_pages REFERENCING NEW TABLE AS newtbl FOR EACH STATEMENT EXECUTE FUNCTION lsif_data_documentation_pages_insert();
DROP TRIGGER IF EXISTS lsif_data_documentation_pages_update ON lsif_data_documentation_pages;
CREATE TRIGGER lsif_data_documentation_pages_update AFTER UPDATE ON lsif_data_documentation_pages REFERENCING OLD TABLE AS oldtbl NEW TABLE AS newtbl FOR EACH STATEMENT EXECUTE FUNCTION lsif_data_documentation_pages_update();

COMMENT ON TABLE lsif_data_docs_search_current_private IS 'A table indicating the most current search index for a unique repository, root, and language.';
COMMENT ON COLUMN lsif_data_docs_search_current_private.repo_id IS 'The repository identifier of the associated dump.';
COMMENT ON COLUMN lsif_data_docs_search_current_private.dump_root IS 'The root of the associated dump.';
COMMENT ON COLUMN lsif_data_docs_search_current_private.lang_name_id IS 'The interned index name of the associated dump.';
COMMENT ON COLUMN lsif_data_docs_search_current_private.dump_id IS 'The associated dump identifier.';
COMMENT ON COLUMN lsif_data_docs_search_current_private.last_cleanup_scan_at IS 'The last time this record was checked as part of a data retention scan.';
COMMENT ON COLUMN lsif_data_docs_search_current_private.created_at IS 'The time this record was inserted. The records with the latest created_at value for the same repository, root, and language is the only visible one and others will be deleted asynchronously.';

COMMENT ON TABLE lsif_data_docs_search_current_public IS 'A table indicating the most current search index for a unique repository, root, and language.';
COMMENT ON COLUMN lsif_data_docs_search_current_public.repo_id IS 'The repository identifier of the associated dump.';
COMMENT ON COLUMN lsif_data_docs_search_current_public.dump_root IS 'The root of the associated dump.';
COMMENT ON COLUMN lsif_data_docs_search_current_public.lang_name_id IS 'The interned index name of the associated dump.';
COMMENT ON COLUMN lsif_data_docs_search_current_public.dump_id IS 'The associated dump identifier.';
COMMENT ON COLUMN lsif_data_docs_search_current_public.last_cleanup_scan_at IS 'The last time this record was checked as part of a data retention scan.';
COMMENT ON COLUMN lsif_data_docs_search_current_public.created_at IS 'The time this record was inserted. The records with the latest created_at value for the same repository, root, and language is the only visible one and others will be deleted asynchronously.';

COMMENT ON TABLE lsif_data_docs_search_lang_names_private IS 'Each unique language name being stored in the API docs search index.';
COMMENT ON COLUMN lsif_data_docs_search_lang_names_private.id IS 'The ID of the language name.';
COMMENT ON COLUMN lsif_data_docs_search_lang_names_private.lang_name IS 'The lowercase language name (go, java, etc.) OR, if unknown, the LSIF indexer name.';
COMMENT ON COLUMN lsif_data_docs_search_lang_names_private.tsv IS 'Indexed tsvector for the lang_name field. Crafted for ordered, case, and punctuation sensitivity, see data_write_documentation.go:textSearchVector.';

COMMENT ON TABLE lsif_data_docs_search_lang_names_public IS 'Each unique language name being stored in the API docs search index.';
COMMENT ON COLUMN lsif_data_docs_search_lang_names_public.id IS 'The ID of the language name.';
COMMENT ON COLUMN lsif_data_docs_search_lang_names_public.lang_name IS 'The lowercase language name (go, java, etc.) OR, if unknown, the LSIF indexer name.';
COMMENT ON COLUMN lsif_data_docs_search_lang_names_public.tsv IS 'Indexed tsvector for the lang_name field. Crafted for ordered, case, and punctuation sensitivity, see data_write_documentation.go:textSearchVector.';

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

COMMENT ON TABLE lsif_data_docs_search_repo_names_private IS 'Each unique repository name being stored in the API docs search index.';
COMMENT ON COLUMN lsif_data_docs_search_repo_names_private.id IS 'The ID of the repository name.';
COMMENT ON COLUMN lsif_data_docs_search_repo_names_private.repo_name IS 'The fully qualified name of the repository.';
COMMENT ON COLUMN lsif_data_docs_search_repo_names_private.tsv IS 'Indexed tsvector for the lang_name field. Crafted for ordered, case, and punctuation sensitivity, see data_write_documentation.go:textSearchVector.';
COMMENT ON COLUMN lsif_data_docs_search_repo_names_private.reverse_tsv IS 'Indexed tsvector for the reverse of the lang_name field, for suffix lexeme/word matching. Crafted for ordered, case, and punctuation sensitivity, see data_write_documentation.go:textSearchVector.';

COMMENT ON TABLE lsif_data_docs_search_repo_names_public IS 'Each unique repository name being stored in the API docs search index.';
COMMENT ON COLUMN lsif_data_docs_search_repo_names_public.id IS 'The ID of the repository name.';
COMMENT ON COLUMN lsif_data_docs_search_repo_names_public.repo_name IS 'The fully qualified name of the repository.';
COMMENT ON COLUMN lsif_data_docs_search_repo_names_public.tsv IS 'Indexed tsvector for the lang_name field. Crafted for ordered, case, and punctuation sensitivity, see data_write_documentation.go:textSearchVector.';
COMMENT ON COLUMN lsif_data_docs_search_repo_names_public.reverse_tsv IS 'Indexed tsvector for the reverse of the lang_name field, for suffix lexeme/word matching. Crafted for ordered, case, and punctuation sensitivity, see data_write_documentation.go:textSearchVector.';

COMMENT ON TABLE lsif_data_docs_search_tags_private IS 'Each uniques sequence of space-separated tags being stored in the API docs search index.';
COMMENT ON COLUMN lsif_data_docs_search_tags_private.id IS 'The ID of the tags.';
COMMENT ON COLUMN lsif_data_docs_search_tags_private.tags IS 'The full sequence of space-separated tags. See protocol/documentation.go:Documentation';
COMMENT ON COLUMN lsif_data_docs_search_tags_private.tsv IS 'Indexed tsvector for the tags field. Crafted for ordered, case, and punctuation sensitivity, see data_write_documentation.go:textSearchVector.';

COMMENT ON TABLE lsif_data_docs_search_tags_public IS 'Each uniques sequence of space-separated tags being stored in the API docs search index.';
COMMENT ON COLUMN lsif_data_docs_search_tags_public.id IS 'The ID of the tags.';
COMMENT ON COLUMN lsif_data_docs_search_tags_public.tags IS 'The full sequence of space-separated tags. See protocol/documentation.go:Documentation';
COMMENT ON COLUMN lsif_data_docs_search_tags_public.tsv IS 'Indexed tsvector for the tags field. Crafted for ordered, case, and punctuation sensitivity, see data_write_documentation.go:textSearchVector.';

COMMENT ON TABLE lsif_data_documentation_mappings IS 'Maps documentation path IDs to their corresponding integral documentationResult vertex IDs, which are unique within a dump.';
COMMENT ON COLUMN lsif_data_documentation_mappings.dump_id IS 'The identifier of the associated dump in the lsif_uploads table (state=completed).';
COMMENT ON COLUMN lsif_data_documentation_mappings.path_id IS 'The documentation page path ID, see see GraphQL codeintel.schema:documentationPage for what this is.';
COMMENT ON COLUMN lsif_data_documentation_mappings.result_id IS 'The documentationResult vertex ID.';
COMMENT ON COLUMN lsif_data_documentation_mappings.file_path IS 'The document file path for the documentationResult, if any. e.g. the path to the file where the symbol described by this documentationResult is located, if it is a symbol.';

COMMENT ON TABLE lsif_data_documentation_pages IS 'Associates documentation pathIDs to their documentation page hierarchy chunk.';
COMMENT ON COLUMN lsif_data_documentation_pages.dump_id IS 'The identifier of the associated dump in the lsif_uploads table (state=completed).';
COMMENT ON COLUMN lsif_data_documentation_pages.path_id IS 'The documentation page path ID, see see GraphQL codeintel.schema:documentationPage for what this is.';
COMMENT ON COLUMN lsif_data_documentation_pages.data IS 'A gob-encoded payload conforming to a `type DocumentationPageData struct` pointer (lib/codeintel/semantic/types.go)';

COMMENT ON TABLE lsif_data_documentation_path_info IS 'Associates documentation page pathIDs to information about what is at that pathID, its immediate children, etc.';
COMMENT ON COLUMN lsif_data_documentation_path_info.dump_id IS 'The identifier of the associated dump in the lsif_uploads table (state=completed).';
COMMENT ON COLUMN lsif_data_documentation_path_info.path_id IS 'The documentation page path ID, see see GraphQL codeintel.schema:documentationPage for what this is.';
COMMENT ON COLUMN lsif_data_documentation_path_info.data IS 'A gob-encoded payload conforming to a `type DocumentationPathInoData struct` pointer (lib/codeintel/semantic/types.go)';
