BEGIN;

CREATE TABLE IF NOT EXISTS lsif_data_docs_search_current_public (
    repo_id INTEGER NOT NULL,
    dump_root TEXT NOT NULL,
    lang_name_id INTEGER NOT NULL,
    dump_id INTEGER NOT NULL,
    last_cleanup_scan_at timestamp with time zone NOT NULL,

    PRIMARY KEY (repo_id, dump_root, lang_name_id)
);

COMMENT ON TABLE lsif_data_docs_search_current_public IS 'A table indicating the most current search index for a unique repository, root, and language.';
COMMENT ON COLUMN lsif_data_docs_search_current_public.repo_id IS 'The repository identifier of the associated dump.';
COMMENT ON COLUMN lsif_data_docs_search_current_public.dump_root IS 'The root of the associated dump.';
COMMENT ON COLUMN lsif_data_docs_search_current_public.lang_name_id IS 'The interned index name of the associated dump.';
COMMENT ON COLUMN lsif_data_docs_search_current_public.dump_id IS 'The most recent dump identifier for this key. See associated content in the lsif_data_docs_search_public table.';
COMMENT ON COLUMN lsif_data_docs_search_current_public.last_cleanup_scan_at IS 'The last time outdated records in the lsif_data_docs_search_public table have been cleaned.';

CREATE TABLE IF NOT EXISTS lsif_data_docs_search_current_private (
    repo_id INTEGER NOT NULL,
    dump_root TEXT NOT NULL,
    lang_name_id INTEGER NOT NULL,
    dump_id INTEGER NOT NULL,
    last_cleanup_scan_at timestamp with time zone NOT NULL,

    PRIMARY KEY (repo_id, dump_root, lang_name_id)
);

COMMENT ON TABLE lsif_data_docs_search_current_private IS 'A table indicating the most current search index for a unique repository, root, and language.';
COMMENT ON COLUMN lsif_data_docs_search_current_private.repo_id IS 'The repository identifier of the associated dump.';
COMMENT ON COLUMN lsif_data_docs_search_current_private.dump_root IS 'The root of the associated dump.';
COMMENT ON COLUMN lsif_data_docs_search_current_private.lang_name_id IS 'The interned index name of the associated dump.';
COMMENT ON COLUMN lsif_data_docs_search_current_private.dump_id IS 'The most recent dump identifier for this key. See associated content in the lsif_data_docs_search_private table.';
COMMENT ON COLUMN lsif_data_docs_search_current_private.last_cleanup_scan_at IS 'The last time outdated records in the lsif_data_docs_search_private table have been cleaned.';

COMMIT;
