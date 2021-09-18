BEGIN;

ALTER TABLE versions ADD COLUMN first_version text;
UPDATE versions set first_version = version;
ALTER TABLE versions ALTER COLUMN first_version SET NOT NULL;

-- Clear the dirty flag in case the operator timed out and isn't around to clear it.
UPDATE schema_migrations SET dirty = 'f'
COMMIT;
