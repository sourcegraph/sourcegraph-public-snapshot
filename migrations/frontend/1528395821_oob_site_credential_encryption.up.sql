BEGIN;

INSERT INTO out_of_band_migrations (id, team, component, description, introduced, non_destructive)
VALUES (
    10,                                         -- This must be consistent across all Sourcegraph instances
    'batch-changes',                            -- Team owning migration
    'frontend-db.site-credentials',             -- Component being migrated
    'Encrypt batch changes site credentials',   -- Description
    '3.28.0',                                   -- The next minor release
    false                                       -- Can be read with previous version without down migration
)
ON CONFLICT DO NOTHING;

ALTER TABLE
    batch_changes_site_credentials
ADD COLUMN IF NOT EXISTS
    credential_enc BYTEA NULL,
ALTER COLUMN
    credential DROP NOT NULL,
DROP CONSTRAINT IF EXISTS
    batch_changes_site_credentials_there_can_be_only_one,
ADD CONSTRAINT
    batch_changes_site_credentials_there_can_be_only_one
    CHECK
    (num_nonnulls(credential, credential_enc) = 1);

-- Create an index on credential_enc, since we want to quickly check its null
-- state when calculating the progress of the OOB migration. Note that we can't
-- apply an index to the actual field because it may be (and in many cases
-- probably is) beyond the limit for a B-tree index.
CREATE INDEX IF NOT EXISTS
    batch_changes_site_credentials_credential_enc_idx
ON
    batch_changes_site_credentials ((credential_enc IS NULL));

COMMIT;
