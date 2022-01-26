-- +++
-- parent: 1000000022
-- +++

BEGIN;

ALTER TABLE lsif_data_documentation_search_public ADD COLUMN dump_root TEXT NOT NULL DEFAULT '';
COMMENT ON COLUMN lsif_data_documentation_search_public.dump_root IS 'Identical to lsif_dumps.root; The working directory of the indexer image relative to the repository root.';
CREATE INDEX lsif_data_documentation_search_public_dump_root_idx ON lsif_data_documentation_search_public USING BTREE(dump_root);

ALTER TABLE lsif_data_documentation_search_private ADD COLUMN dump_root TEXT NOT NULL DEFAULT '';
COMMENT ON COLUMN lsif_data_documentation_search_private.dump_root IS 'Identical to lsif_dumps.root; The working directory of the indexer image relative to the repository root.';
CREATE INDEX lsif_data_documentation_search_private_dump_root_idx ON lsif_data_documentation_search_public USING BTREE(dump_root);

-- Truncate both tables; we don't care about reindexing given so little has been indexed to date.
TRUNCATE lsif_data_documentation_search_public;
TRUNCATE lsif_data_documentation_search_private;

COMMIT;
