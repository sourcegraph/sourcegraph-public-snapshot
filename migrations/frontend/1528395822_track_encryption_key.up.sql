BEGIN;

-- We're going to unify the credential columns here. This means that we need to
-- preserve the value in credential if the OOB migrator hasn't run since the
-- previous credential migrations.
--
-- Since this will break the constraint we added previously, first we have to
-- drop that.

ALTER TABLE
    batch_changes_site_credentials
ADD COLUMN IF NOT EXISTS
    encryption_key_id TEXT NOT NULL DEFAULT '',
DROP CONSTRAINT IF EXISTS
    batch_changes_site_credentials_there_can_be_only_one;

ALTER TABLE
    user_credentials
ADD COLUMN IF NOT EXISTS
    encryption_key_id TEXT NOT NULL DEFAULT '',
DROP CONSTRAINT IF EXISTS
    user_credentials_there_can_be_only_one;

-- Previously upgraded credentials with encryption need a placeholder encryption
-- ID so that we can replace it with a real one later in the OOB migrator.
--
-- Unfortunately, the lack of inline metadata means that we have to use a
-- heuristic to determine if the credential was _actually_ encrypted or not.
-- Practically speaking, this only matters for users who (a) enabled encryption
-- for Batch Changes, and (b) ran a version of Sourcegraph between May 4 and May
-- 6. That's only going to be two developers on the Batch Changes team, so this
-- leaky heuristic should be fine.

UPDATE
    batch_changes_site_credentials
SET
    encryption_key_id = 'previously-migrated'
WHERE
    credential_enc IS NOT NULL
    AND NOT (
        LEFT(ENCODE(credential_enc, 'escape'), 1) = '{'
        AND RIGHT(ENCODE(credential_enc, 'escape'), 1) = '}'
    );

UPDATE
    user_credentials
SET
    encryption_key_id = 'previously-migrated'
WHERE
    credential_enc IS NOT NULL
    AND NOT (
        LEFT(ENCODE(credential_enc, 'escape'), 1) = '{'
        AND RIGHT(ENCODE(credential_enc, 'escape'), 1) = '}'
    );

-- Now we shift credentials into the new field.

UPDATE
    batch_changes_site_credentials
SET
    credential_enc = CONVERT_TO(credential, 'UTF8')
WHERE
    credential_enc IS NULL;

UPDATE
    user_credentials
SET
    credential_enc = CONVERT_TO(credential, 'UTF8')
WHERE
    credential_enc IS NULL;

-- And finally we can rename the field and update the indexes on the tables.

DROP INDEX IF EXISTS
    batch_changes_site_credentials_credential_enc_idx,
    batch_changes_site_credentials_credential_idx,
    user_credentials_credential_enc_idx,
    user_credentials_credential_idx;

ALTER TABLE
    batch_changes_site_credentials
DROP COLUMN IF EXISTS
    credential,
ALTER COLUMN
    credential_enc SET NOT NULL;

ALTER TABLE
    batch_changes_site_credentials
RENAME COLUMN
    credential_enc TO credential;

ALTER TABLE
    user_credentials
DROP COLUMN IF EXISTS
    credential,
ALTER COLUMN
    credential_enc SET NOT NULL;

ALTER TABLE
    user_credentials
RENAME COLUMN
    credential_enc TO credential;

CREATE INDEX IF NOT EXISTS
    batch_changes_site_credentials_credential_idx
ON
    batch_changes_site_credentials ((encryption_key_id IN ('', 'previously-migrated')));

CREATE INDEX IF NOT EXISTS
    user_credentials_credential_idx
ON
    user_credentials ((encryption_key_id IN ('', 'previously-migrated')));

COMMIT;
