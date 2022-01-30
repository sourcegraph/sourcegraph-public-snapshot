BEGIN;
CREATE INDEX IF NOT EXISTS lsif_data_documentation_pages_dump_id_unindexed ON lsif_data_documentation_pages(dump_id) WHERE NOT search_indexed;
COMMIT;
