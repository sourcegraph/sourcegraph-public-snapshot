BEGIN;

-- faster to supply default than manual update
ALTER TABLE lsif_data_documents ADD COLUMN schema_version int DEFAULT 1 NOT NULL;
ALTER TABLE lsif_data_documents ADD COLUMN num_diagnostics int DEFAULT 0 NOT NULL;

-- drop default after all existing columns have been set
ALTER TABLE lsif_data_documents ALTER COLUMN schema_version DROP DEFAULT;
ALTER TABLE lsif_data_documents ALTER COLUMN num_diagnostics DROP DEFAULT;

COMMENT ON COLUMN lsif_data_documents.schema_version IS 'The schema version of this row - used to determine presence and encoding of data.';
COMMENT ON COLUMN lsif_data_documents.num_diagnostics IS 'The number of diagnostics stored in the data field.';

COMMIT;
