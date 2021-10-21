BEGIN;

--
-- Public

-- Drop new columns
ALTER TABLE lsif_data_docs_search_current_public DROP COLUMN IF EXISTS created_at;
ALTER TABLE lsif_data_docs_search_current_public DROP COLUMN IF EXISTS id;

-- Re-create old primary key
ALTER TABLE lsif_data_docs_search_current_public ADD PRIMARY KEY (repo_id, dump_root, lang_name_id);

-- Restore old comment
COMMENT ON COLUMN lsif_data_docs_search_current_public.dump_id IS 'The most recent dump identifier for this key. See associated content in the lsif_data_docs_search_public table.';

--
-- Private

-- Drop new columns
ALTER TABLE lsif_data_docs_search_current_private DROP COLUMN IF EXISTS created_at;
ALTER TABLE lsif_data_docs_search_current_private DROP COLUMN IF EXISTS id;

-- Re-create old primary key
ALTER TABLE lsif_data_docs_search_current_private ADD PRIMARY KEY (repo_id, dump_root, lang_name_id);

-- Restore old comment
COMMENT ON COLUMN lsif_data_docs_search_current_private.dump_id IS 'The most recent dump identifier for this key. See associated content in the lsif_data_docs_search_public table.';

COMMIT;
