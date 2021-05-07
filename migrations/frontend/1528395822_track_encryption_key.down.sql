BEGIN;

-- Where we're going, we don't need indexes. (Yet.)

DROP INDEX IF EXISTS
    batch_changes_site_credentials_credential_enc_idx,
    batch_changes_site_credentials_credential_idx,
    user_credentials_credential_enc_idx,
    user_credentials_credential_idx;

-- Reinstate the old credential field.

ALTER TABLE
    batch_changes_site_credentials
RENAME COLUMN
    credential TO credential_enc;

ALTER TABLE
    batch_changes_site_credentials
ADD COLUMN IF NOT EXISTS
    credential TEXT NULL DEFAULT NULL,
ALTER COLUMN
    credential_enc DROP NOT NULL,
DROP CONSTRAINT IF EXISTS
    batch_changes_site_credentials_there_can_be_only_one,
ADD CONSTRAINT
    batch_changes_site_credentials_there_can_be_only_one
    CHECK
    (num_nonnulls(credential, credential_enc) = 1);

ALTER TABLE
    user_credentials
RENAME COLUMN
    credential TO credential_enc;

ALTER TABLE
    user_credentials
ADD COLUMN IF NOT EXISTS
    credential TEXT NULL DEFAULT NULL,
ALTER COLUMN
    credential_enc DROP NOT NULL,
DROP CONSTRAINT IF EXISTS
    user_credentials_there_can_be_only_one,
ADD CONSTRAINT
    user_credentials_there_can_be_only_one
    CHECK
    (num_nonnulls(credential, credential_enc) = 1);

UPDATE
    batch_changes_site_credentials
SET
    credential = ENCODE(credential_enc, 'escape'),
    credential_enc = NULL
WHERE
    encryption_key_id = '';

UPDATE
    user_credentials
SET
    credential = ENCODE(credential_enc, 'escape'),
    credential_enc = NULL
WHERE
    encryption_key_id = '';

-- Put the indexes back.

CREATE INDEX IF NOT EXISTS
    user_credentials_credential_enc_idx
ON
    user_credentials ((credential_enc IS NULL));

CREATE INDEX IF NOT EXISTS
    batch_changes_site_credentials_credential_enc_idx
ON
    batch_changes_site_credentials ((credential_enc IS NULL));

-- Get rid of the new encryption_key_id field.

ALTER TABLE
    batch_changes_site_credentials
DROP COLUMN IF EXISTS
    encryption_key_id;

ALTER TABLE
    user_credentials
DROP COLUMN IF EXISTS
    encryption_key_id;

COMMIT;
