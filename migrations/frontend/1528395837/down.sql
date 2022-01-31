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

COMMIT;
