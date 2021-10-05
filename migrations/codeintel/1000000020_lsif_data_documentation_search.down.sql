BEGIN;

ALTER TABLE lsif_data_documentation_pages DROP COLUMN search_indexed;
DROP TABLE IF EXISTS lsif_data_documentation_search;

COMMIT;
