BEGIN;

-- Undo the changes corresponding to https://github.com/sourcegraph/sourcegraph/pull/25715
ALTER TABLE lsif_data_documentation_search_public DROP COLUMN IF EXISTS dump_root;
ALTER TABLE lsif_data_documentation_search_private DROP COLUMN IF EXISTS dump_root;

COMMIT;
