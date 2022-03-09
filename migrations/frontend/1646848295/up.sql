ALTER TABLE lsif_uploads ADD COLUMN IF NOT EXISTS indexer_version text;
ALTER TABLE lsif_indexes ADD COLUMN IF NOT EXISTS indexer_version text;

COMMENT ON COLUMN lsif_uploads.indexer_version IS 'The version of the indexer that produced the index file. If not supplied by the user it will be pulled from the index metadata.';
COMMENT ON COLUMN lsif_indexes.indexer_version IS 'The version of the indexer that produced the index file. If not supplied by the user it will be pulled from the index metadata.';
