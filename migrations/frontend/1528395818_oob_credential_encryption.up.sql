BEGIN;

INSERT INTO out_of_band_migrations (id, team, component, description, introduced, non_destructive)
VALUES (
    9,                                          -- This must be consistent across all Sourcegraph instances
    'batch-changes',                            -- Team owning migration
    'frontend-db.user-credentials',             -- Component being migrated
    'Encrypt batch changes user credentials',   -- Description
    '3.28.0',                                   -- The next minor release
    true                                        -- Can be read with previous version without down migration
)
ON CONFLICT DO NOTHING;

ALTER TABLE
    user_credentials
ADD COLUMN
    credential_enc BYTEA NULL,
ALTER COLUMN
    credential DROP NOT NULL,
ADD CONSTRAINT
    user_credentials_there_can_be_only_one
    CHECK
    (num_nonnulls(credential, credential_enc) = 1);

-- TODO: add another OOB migration for site credentials.

COMMIT;
