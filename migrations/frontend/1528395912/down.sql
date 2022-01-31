BEGIN;

-- The OOB migration doesn't add any new tables or columns or anything, so we don't need to do
-- anything on down migration. It migrates data from lsif_data_documentation_pages -> the new
-- lsif_data_documentation_search_* tables - but it's fine to just leave those.

COMMIT;
