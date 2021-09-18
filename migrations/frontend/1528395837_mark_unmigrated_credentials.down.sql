BEGIN;

UPDATE
    user_credentials
SET
    encryption_key_id = ''
WHERE
    encryption_key_id = 'unmigrated';

UPDATE
    batch_changes_site_credentials
SET
    encryption_key_id = ''
WHERE
    encryption_key_id = 'unmigrated';

-- Clear the dirty flag in case the operator timed out and isn't around to clear it.
UPDATE schema_migrations SET dirty = 'f'
COMMIT;
