BEGIN;

ALTER TABLE IF EXISTS lsif_uploads ADD COLUMN IF NOT EXISTS execution_logs JSON[];

-- Clear the dirty flag in case the operator timed out and isn't around to clear it.
UPDATE schema_migrations SET dirty = 'f'
COMMIT;
