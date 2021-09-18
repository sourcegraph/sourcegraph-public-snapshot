BEGIN;

ALTER TABLE lsif_data_documents DROP COLUMN ranges;
ALTER TABLE lsif_data_documents DROP COLUMN hovers;
ALTER TABLE lsif_data_documents DROP COLUMN monikers;
ALTER TABLE lsif_data_documents DROP COLUMN packages;
ALTER TABLE lsif_data_documents DROP COLUMN diagnostics;

-- Clear the dirty flag in case the operator timed out and isn't around to clear it.
UPDATE codeintel_schema_migrations SET dirty = 'f'
COMMIT;
