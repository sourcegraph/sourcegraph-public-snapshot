BEGIN;

-- We're completely changing the API docs search table schema, so we'll reindex everything
-- from scratch. Reset our OOB migration's progress entirely.
--
-- IMPORTANT: Dropping the column and recreating it is nearly instant, updating the table
-- to set the column to 'false' again can take several minutes.
ALTER TABLE lsif_data_documentation_pages DROP COLUMN search_indexed;
ALTER TABLE lsif_data_documentation_pages ADD COLUMN search_indexed boolean DEFAULT 'false';

-- We're completely redefining the table.
DROP TABLE IF EXISTS lsif_data_documentation_search_public;

-- Each unique language name being stored in the search index.
--
-- Contains a tsvector index for matching a logical OR of query terms against the language name
-- (e.g. "http router go" to match "go" without knowing if "http", "router", or "go" are actually
-- a language name or not.)
CREATE TABLE lsif_data_docs_search_lang_names_public (
    id BIGSERIAL PRIMARY KEY,
    lang_name TEXT NOT NULL UNIQUE,
    tsv TSVECTOR NOT NULL
);
CREATE INDEX lsif_data_docs_search_lang_names_public_tsv_idx ON lsif_data_docs_search_lang_names_public USING GIN(tsv);

COMMENT ON TABLE lsif_data_docs_search_lang_names_public IS 'Each unique language name being stored in the API docs search index.';
COMMENT ON COLUMN lsif_data_docs_search_lang_names_public.id IS 'The ID of the language name.';
COMMENT ON COLUMN lsif_data_docs_search_lang_names_public.lang_name IS 'The lowercase language name (go, java, etc.) OR, if unknown, the LSIF indexer name.';
COMMENT ON COLUMN lsif_data_docs_search_lang_names_public.tsv IS 'Indexed tsvector for the lang_name field. Crafted for ordered, case, and punctuation sensitivity, see data_write_documentation.go:textSearchVector.';

-- Each unique repository name being stored in the search index.
--
-- Contains a tsvector index for matching against repository names, with both prefix and suffix
-- (reverse prefix) matching within lexemes (words).
CREATE TABLE lsif_data_docs_search_repo_names_public (
    id BIGSERIAL PRIMARY KEY,
    repo_name TEXT NOT NULL UNIQUE,
    tsv TSVECTOR NOT NULL,
    reverse_tsv TSVECTOR NOT NULL
);
CREATE INDEX lsif_data_docs_search_repo_names_public_tsv_idx ON lsif_data_docs_search_repo_names_public USING GIN(tsv);
CREATE INDEX lsif_data_docs_search_repo_names_public_reverse_tsv_idx ON lsif_data_docs_search_repo_names_public USING GIN(reverse_tsv);

COMMENT ON TABLE lsif_data_docs_search_repo_names_public IS 'Each unique repository name being stored in the API docs search index.';
COMMENT ON COLUMN lsif_data_docs_search_repo_names_public.id IS 'The ID of the repository name.';
COMMENT ON COLUMN lsif_data_docs_search_repo_names_public.repo_name IS 'The fully qualified name of the repository.';
COMMENT ON COLUMN lsif_data_docs_search_repo_names_public.tsv IS 'Indexed tsvector for the lang_name field. Crafted for ordered, case, and punctuation sensitivity, see data_write_documentation.go:textSearchVector.';
COMMENT ON COLUMN lsif_data_docs_search_repo_names_public.reverse_tsv IS 'Indexed tsvector for the reverse of the lang_name field, for suffix lexeme/word matching. Crafted for ordered, case, and punctuation sensitivity, see data_write_documentation.go:textSearchVector.';

-- Each unique sequence of space-separated tags being stored in the search index. This could be as
-- many rows as the search table itself, because in theory each result could have a unique string
-- of tags. In practice, though, they are frequently similar sequences.
--
-- The space separated tags have a tsvector for matching a logcal OR of query terms against, for
-- the same reason as the lang_names table. e.g. so that we can have a query for "go private function net router"
-- match the tags string "private function" without knowing which query terms are tags or not.
--
-- The entire sequence of space-separated tags are stored, in part so that lookups in the search table
-- are faster (single ID lookup rather than array ALL lookup) and partly to allow for more complex
-- tag matching options in the future.
CREATE TABLE lsif_data_docs_search_tags_public (
    id BIGSERIAL PRIMARY KEY,
    tags TEXT NOT NULL UNIQUE,
    tsv TSVECTOR NOT NULL
);
CREATE INDEX lsif_data_docs_search_tags_public_tsv_idx ON lsif_data_docs_search_tags_public USING GIN(tsv);

COMMENT ON TABLE lsif_data_docs_search_tags_public IS 'Each uniques sequence of space-separated tags being stored in the API docs search index.';
COMMENT ON COLUMN lsif_data_docs_search_tags_public.id IS 'The ID of the tags.';
COMMENT ON COLUMN lsif_data_docs_search_tags_public.tags IS 'The full sequence of space-separated tags. See protocol/documentation.go:Documentation';
COMMENT ON COLUMN lsif_data_docs_search_tags_public.tsv IS 'Indexed tsvector for the tags field. Crafted for ordered, case, and punctuation sensitivity, see data_write_documentation.go:textSearchVector.';

-- The actual search index over API docs, one entry per symbol/section of API docs.
CREATE TABLE lsif_data_docs_search_public (
    -- Metadata fields
    id BIGSERIAL PRIMARY KEY,
    repo_id INTEGER NOT NULL,
    dump_id INTEGER NOT NULL,
    dump_root TEXT NOT NULL,
    path_id TEXT NOT NULL,
    detail TEXT NOT NULL,
    lang_name_id INTEGER NOT NULL,
    repo_name_id INTEGER NOT NULL,
    tags_id INTEGER NOT NULL,

    -- FTS-enabled fields
    search_key TEXT NOT NULL,
    search_key_tsv TSVECTOR NOT NULL,
    search_key_reverse_tsv TSVECTOR NOT NULL,

    label TEXT NOT NULL,
    label_tsv TSVECTOR NOT NULL,
    label_reverse_tsv TSVECTOR NOT NULL,

    CONSTRAINT lsif_data_docs_search_public_lang_name_id_fk FOREIGN KEY(lang_name_id) REFERENCES lsif_data_docs_search_lang_names_public(id),
    CONSTRAINT lsif_data_docs_search_public_repo_name_id_fk FOREIGN KEY(repo_name_id) REFERENCES lsif_data_docs_search_repo_names_public(id),
    CONSTRAINT lsif_data_docs_search_public_tags_id_fk FOREIGN KEY(tags_id) REFERENCES lsif_data_docs_search_tags_public(id)
);

-- This pair of fields is used to purge stale data from the search index, so use a btree index on it.
CREATE INDEX lsif_data_docs_search_public_repo_id_idx ON lsif_data_docs_search_public USING BTREE(repo_id);
CREATE INDEX lsif_data_docs_search_public_dump_id_idx ON lsif_data_docs_search_public USING BTREE(dump_id);
CREATE INDEX lsif_data_docs_search_public_dump_root_idx ON lsif_data_docs_search_public USING BTREE(dump_root);

-- tsvector indexes
CREATE INDEX lsif_data_docs_search_public_search_key_tsv_idx ON lsif_data_docs_search_public USING BTREE(search_key_tsv);
CREATE INDEX lsif_data_docs_search_public_search_key_reverse_tsv_idx ON lsif_data_docs_search_public USING BTREE(search_key_reverse_tsv);
CREATE INDEX lsif_data_docs_search_public_label_tsv_idx ON lsif_data_docs_search_public USING BTREE(label_tsv);
CREATE INDEX lsif_data_docs_search_public_label_reverse_tsv_idx ON lsif_data_docs_search_public USING BTREE(label_reverse_tsv);

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

-- ************************************************************************************************
-- Below here is a direct copy of the above, but with "public" replaced with "private" for the    *
-- private variant of the table.                                                                  *
-- ************************************************************************************************

-- We're completely redefining the table.
DROP TABLE IF EXISTS lsif_data_documentation_search_private;

-- Each unique language name being stored in the search index.
--
-- Contains a tsvector index for matching a logical OR of query terms against the language name
-- (e.g. "http router go" to match "go" without knowing if "http", "router", or "go" are actually
-- a language name or not.)
CREATE TABLE lsif_data_docs_search_lang_names_private (
    id BIGSERIAL PRIMARY KEY,
    lang_name TEXT NOT NULL UNIQUE,
    tsv TSVECTOR NOT NULL
);
CREATE INDEX lsif_data_docs_search_lang_names_private_tsv_idx ON lsif_data_docs_search_lang_names_private USING GIN(tsv);

COMMENT ON TABLE lsif_data_docs_search_lang_names_private IS 'Each unique language name being stored in the API docs search index.';
COMMENT ON COLUMN lsif_data_docs_search_lang_names_private.id IS 'The ID of the language name.';
COMMENT ON COLUMN lsif_data_docs_search_lang_names_private.lang_name IS 'The lowercase language name (go, java, etc.) OR, if unknown, the LSIF indexer name.';
COMMENT ON COLUMN lsif_data_docs_search_lang_names_private.tsv IS 'Indexed tsvector for the lang_name field. Crafted for ordered, case, and punctuation sensitivity, see data_write_documentation.go:textSearchVector.';

-- Each unique repository name being stored in the search index.
--
-- Contains a tsvector index for matching against repository names, with both prefix and suffix
-- (reverse prefix) matching within lexemes (words).
CREATE TABLE lsif_data_docs_search_repo_names_private (
    id BIGSERIAL PRIMARY KEY,
    repo_name TEXT NOT NULL UNIQUE,
    tsv TSVECTOR NOT NULL,
    reverse_tsv TSVECTOR NOT NULL
);
CREATE INDEX lsif_data_docs_search_repo_names_private_tsv_idx ON lsif_data_docs_search_repo_names_private USING GIN(tsv);
CREATE INDEX lsif_data_docs_search_repo_names_private_reverse_tsv_idx ON lsif_data_docs_search_repo_names_private USING GIN(reverse_tsv);

COMMENT ON TABLE lsif_data_docs_search_repo_names_private IS 'Each unique repository name being stored in the API docs search index.';
COMMENT ON COLUMN lsif_data_docs_search_repo_names_private.id IS 'The ID of the repository name.';
COMMENT ON COLUMN lsif_data_docs_search_repo_names_private.repo_name IS 'The fully qualified name of the repository.';
COMMENT ON COLUMN lsif_data_docs_search_repo_names_private.tsv IS 'Indexed tsvector for the lang_name field. Crafted for ordered, case, and punctuation sensitivity, see data_write_documentation.go:textSearchVector.';
COMMENT ON COLUMN lsif_data_docs_search_repo_names_private.reverse_tsv IS 'Indexed tsvector for the reverse of the lang_name field, for suffix lexeme/word matching. Crafted for ordered, case, and punctuation sensitivity, see data_write_documentation.go:textSearchVector.';

-- Each unique sequence of space-separated tags being stored in the search index. This could be as
-- many rows as the search table itself, because in theory each result could have a unique string
-- of tags. In practice, though, they are frequently similar sequences.
--
-- The space separated tags have a tsvector for matching a logcal OR of query terms against, for
-- the same reason as the lang_names table. e.g. so that we can have a query for "go private function net router"
-- match the tags string "private function" without knowing which query terms are tags or not.
--
-- The entire sequence of space-separated tags are stored, in part so that lookups in the search table
-- are faster (single ID lookup rather than array ALL lookup) and partly to allow for more complex
-- tag matching options in the future.
CREATE TABLE lsif_data_docs_search_tags_private (
    id BIGSERIAL PRIMARY KEY,
    tags TEXT NOT NULL UNIQUE,
    tsv TSVECTOR NOT NULL
);
CREATE INDEX lsif_data_docs_search_tags_private_tsv_idx ON lsif_data_docs_search_tags_private USING GIN(tsv);

COMMENT ON TABLE lsif_data_docs_search_tags_private IS 'Each uniques sequence of space-separated tags being stored in the API docs search index.';
COMMENT ON COLUMN lsif_data_docs_search_tags_private.id IS 'The ID of the tags.';
COMMENT ON COLUMN lsif_data_docs_search_tags_private.tags IS 'The full sequence of space-separated tags. See protocol/documentation.go:Documentation';
COMMENT ON COLUMN lsif_data_docs_search_tags_private.tsv IS 'Indexed tsvector for the tags field. Crafted for ordered, case, and punctuation sensitivity, see data_write_documentation.go:textSearchVector.';

-- The actual search index over API docs, one entry per symbol/section of API docs.
CREATE TABLE lsif_data_docs_search_private (
    -- Metadata fields
    id BIGSERIAL PRIMARY KEY,
    repo_id INTEGER NOT NULL,
    dump_id INTEGER NOT NULL,
    dump_root TEXT NOT NULL,
    path_id TEXT NOT NULL,
    detail TEXT NOT NULL,
    lang_name_id INTEGER NOT NULL,
    repo_name_id INTEGER NOT NULL,
    tags_id INTEGER NOT NULL,

    -- FTS-enabled fields
    search_key TEXT NOT NULL,
    search_key_tsv TSVECTOR NOT NULL,
    search_key_reverse_tsv TSVECTOR NOT NULL,

    label TEXT NOT NULL,
    label_tsv TSVECTOR NOT NULL,
    label_reverse_tsv TSVECTOR NOT NULL,

    CONSTRAINT lsif_data_docs_search_private_lang_name_id_fk FOREIGN KEY(lang_name_id) REFERENCES lsif_data_docs_search_lang_names_private(id),
    CONSTRAINT lsif_data_docs_search_private_repo_name_id_fk FOREIGN KEY(repo_name_id) REFERENCES lsif_data_docs_search_repo_names_private(id),
    CONSTRAINT lsif_data_docs_search_private_tags_id_fk FOREIGN KEY(tags_id) REFERENCES lsif_data_docs_search_tags_private(id)
);

-- This pair of fields is used to purge stale data from the search index, so use a btree index on it.
CREATE INDEX lsif_data_docs_search_private_repo_id_idx ON lsif_data_docs_search_private USING BTREE(repo_id);
CREATE INDEX lsif_data_docs_search_private_dump_id_idx ON lsif_data_docs_search_private USING BTREE(dump_id);
CREATE INDEX lsif_data_docs_search_private_dump_root_idx ON lsif_data_docs_search_private USING BTREE(dump_root);

-- tsvector indexes
CREATE INDEX lsif_data_docs_search_private_search_key_tsv_idx ON lsif_data_docs_search_private USING BTREE(search_key_tsv);
CREATE INDEX lsif_data_docs_search_private_search_key_reverse_tsv_idx ON lsif_data_docs_search_private USING BTREE(search_key_reverse_tsv);
CREATE INDEX lsif_data_docs_search_private_label_tsv_idx ON lsif_data_docs_search_private USING BTREE(label_tsv);
CREATE INDEX lsif_data_docs_search_private_label_reverse_tsv_idx ON lsif_data_docs_search_private USING BTREE(label_reverse_tsv);

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

COMMIT;
