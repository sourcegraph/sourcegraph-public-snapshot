BEGIN;

ALTER TABLE lsif_data_documentation_search_public DROP COLUMN dump_root;
ALTER TABLE lsif_data_documentation_search_private DROP COLUMN dump_root;

COMMIT;
