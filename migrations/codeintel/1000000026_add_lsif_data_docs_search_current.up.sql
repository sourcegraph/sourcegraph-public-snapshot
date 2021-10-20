BEGIN;

CREATE TABLE IF NOT EXISTS lsif_data_docs_search_current_public (
    repo_id INTEGER NOT NULL,
    dump_root TEXT NOT NULL,
    lang_name_id INTEGER NOT NULL,
    dump_id INTEGER NOT NULL,
    last_cleanup_scan_at timestamp with time zone NOT NULL,

    PRIMARY KEY (repo_id, dump_root, lang_name_id)
);

-- TODO - comments
-- TODO - fill default values?

CREATE TABLE IF NOT EXISTS lsif_data_docs_search_current_private (
    repo_id INTEGER NOT NULL,
    dump_root TEXT NOT NULL,
    lang_name_id INTEGER NOT NULL,
    dump_id INTEGER NOT NULL,
    last_cleanup_scan_at timestamp with time zone NOT NULL,

    PRIMARY KEY (repo_id, dump_root, lang_name_id)
);

-- TODO - comments
-- TODO - fill default values?

COMMIT;
