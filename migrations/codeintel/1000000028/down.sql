BEGIN;

--
-- Public

-- De-duplicate records before adding the unique index
DELETE FROM lsif_data_docs_search_current_public WHERE id NOT IN (
    SELECT MAX(id) as max_id
    FROM lsif_data_docs_search_current_public
    GROUP BY repo_id, dump_root, lang_name_id
);

-- Drop new columns
ALTER TABLE lsif_data_docs_search_current_public DROP COLUMN IF EXISTS id;
ALTER TABLE lsif_data_docs_search_current_public DROP COLUMN IF EXISTS created_at;

-- Drop new index
DROP INDEX IF EXISTS lsif_data_docs_search_current_public_last_cleanup_scan_at;

-- Re-create old primary key
ALTER TABLE lsif_data_docs_search_current_public ADD PRIMARY KEY (repo_id, dump_root, lang_name_id);

-- Restore old comment
COMMENT ON COLUMN lsif_data_docs_search_current_public.dump_id IS 'The most recent dump identifier for this key. See associated content in the lsif_data_docs_search_public table.';

-- Restore old last_cleanup_scan_at column
COMMENT ON COLUMN lsif_data_docs_search_current_public.last_cleanup_scan_at IS 'The last time outdated records in the lsif_data_docs_search_public table have been cleaned.';
ALTER TABLE lsif_data_docs_search_current_public ALTER COLUMN last_cleanup_scan_at DROP DEFAULT;

--
-- Private

-- De-duplicate records before adding the unique index
DELETE FROM lsif_data_docs_search_current_private WHERE id NOT IN (
    SELECT MAX(id) as max_id
    FROM lsif_data_docs_search_current_private
    GROUP BY repo_id, dump_root, lang_name_id
);

-- Drop new columns
ALTER TABLE lsif_data_docs_search_current_private DROP COLUMN IF EXISTS id;
ALTER TABLE lsif_data_docs_search_current_private DROP COLUMN IF EXISTS created_at;

-- Drop new index
DROP INDEX IF EXISTS lsif_data_docs_search_current_private_last_cleanup_scan_at;

-- Re-create old primary key
ALTER TABLE lsif_data_docs_search_current_private ADD PRIMARY KEY (repo_id, dump_root, lang_name_id);

-- Restore old comment
COMMENT ON COLUMN lsif_data_docs_search_current_private.dump_id IS 'The most recent dump identifier for this key. See associated content in the lsif_data_docs_search_public table.';

-- Restore old last_cleanup_scan_at column
COMMENT ON COLUMN lsif_data_docs_search_current_private.last_cleanup_scan_at IS 'The last time outdated records in the lsif_data_docs_search_public table have been cleaned.';
ALTER TABLE lsif_data_docs_search_current_private ALTER COLUMN last_cleanup_scan_at DROP DEFAULT;

COMMIT;
