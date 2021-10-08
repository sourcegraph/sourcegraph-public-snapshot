BEGIN;

ALTER TABLE lsif_data_documentation_search_public DROP COLUMN IF EXISTS dump_root;
ALTER TABLE lsif_data_documentation_search_private DROP COLUMN IF EXISTS dump_root;

COMMIT;
