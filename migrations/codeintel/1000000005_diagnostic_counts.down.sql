BEGIN;

ALTER TABLE lsif_data_documents DROP COLUMN IF EXISTS schema_version;
ALTER TABLE lsif_data_documents DROP COLUMN IF EXISTS num_diagnostics;

COMMIT;
