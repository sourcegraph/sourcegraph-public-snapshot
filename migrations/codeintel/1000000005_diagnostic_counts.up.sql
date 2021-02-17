BEGIN;

-- faster to supply default than manual update
ALTER TABLE lsif_data_documents ADD COLUMN schema_version int DEFAULT 1 NOT NULL;
ALTER TABLE lsif_data_documents ADD COLUMN num_diagnostics int DEFAULT 0 NOT NULL;

-- drop default after all existing columns have been set
ALTER TABLE lsif_data_documents ALTER COLUMN schema_version DROP DEFAULT;
ALTER TABLE lsif_data_documents ALTER COLUMN num_diagnostics DROP DEFAULT;

COMMENT ON COLUMN lsif_data_documents.schema_version IS 'The version of the current row. Used to track out-of-band migrations.';
COMMENT ON COLUMN lsif_data_documents.num_diagnostics IS 'The number of diagnostics stored in the data field.';

COMMIT;
