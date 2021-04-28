BEGIN;

ALTER TABLE lsif_data_documents DROP COLUMN ranges;
ALTER TABLE lsif_data_documents DROP COLUMN hovers;
ALTER TABLE lsif_data_documents DROP COLUMN monikers;
ALTER TABLE lsif_data_documents DROP COLUMN packages;
ALTER TABLE lsif_data_documents DROP COLUMN diagnostics;

COMMIT;
