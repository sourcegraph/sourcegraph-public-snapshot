-- +++
-- parent: 1000000028
-- +++

BEGIN;

-- Drop the btree indexes that we intended to be GIN indexes.
-- btree indexes are no where near as performant for tsvector indexing.
DROP INDEX IF EXISTS lsif_data_docs_search_public_search_key_tsv_idx;
DROP INDEX IF EXISTS lsif_data_docs_search_public_search_key_reverse_tsv_idx;
DROP INDEX IF EXISTS lsif_data_docs_search_public_label_tsv_idx;
DROP INDEX IF EXISTS lsif_data_docs_search_public_label_reverse_tsv_idx;

DROP INDEX IF EXISTS lsif_data_docs_search_private_search_key_tsv_idx;
DROP INDEX IF EXISTS lsif_data_docs_search_private_search_key_reverse_tsv_idx;
DROP INDEX IF EXISTS lsif_data_docs_search_private_label_tsv_idx;
DROP INDEX IF EXISTS lsif_data_docs_search_private_label_reverse_tsv_idx;

-- Recreate indexes with GIN instead.
CREATE INDEX IF NOT EXISTS lsif_data_docs_search_public_search_key_tsv_idx ON lsif_data_docs_search_public USING GIN (search_key_tsv);
CREATE INDEX IF NOT EXISTS lsif_data_docs_search_public_search_key_reverse_tsv_idx ON lsif_data_docs_search_public USING GIN (search_key_reverse_tsv);
CREATE INDEX IF NOT EXISTS lsif_data_docs_search_public_label_tsv_idx ON lsif_data_docs_search_public USING GIN (label_tsv);
CREATE INDEX IF NOT EXISTS lsif_data_docs_search_public_label_reverse_tsv_idx ON lsif_data_docs_search_public USING GIN (label_reverse_tsv);

CREATE INDEX IF NOT EXISTS lsif_data_docs_search_private_search_key_tsv_idx ON lsif_data_docs_search_private USING GIN (search_key_tsv);
CREATE INDEX IF NOT EXISTS lsif_data_docs_search_private_search_key_reverse_tsv_idx ON lsif_data_docs_search_private USING GIN (search_key_reverse_tsv);
CREATE INDEX IF NOT EXISTS lsif_data_docs_search_private_label_tsv_idx ON lsif_data_docs_search_private USING GIN (label_tsv);
CREATE INDEX IF NOT EXISTS lsif_data_docs_search_private_label_reverse_tsv_idx ON lsif_data_docs_search_private USING GIN (label_reverse_tsv);

COMMIT;
