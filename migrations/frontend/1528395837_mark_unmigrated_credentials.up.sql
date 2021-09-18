BEGIN;

-- Previously, we conflated unmigrated user and site credentials with
-- unencrypted ones. Instead, we should separate these states with a placeholder
-- so the out of band migration responsible for encrypting credentials reports
-- its progress correctly.

UPDATE
    user_credentials
SET
    encryption_key_id = 'unmigrated'
WHERE
    encryption_key_id = '';

UPDATE
    batch_changes_site_credentials
SET
    encryption_key_id = 'unmigrated'
WHERE
    encryption_key_id = '';

-- Clear the dirty flag in case the operator timed out and isn't around to clear it.
UPDATE schema_migrations SET dirty = 'f'
COMMIT;
