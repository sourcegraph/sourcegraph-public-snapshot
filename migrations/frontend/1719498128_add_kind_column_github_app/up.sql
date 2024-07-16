DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'github_app_kind') THEN
        CREATE TYPE github_app_kind AS ENUM (
            'COMMIT_SIGNING',
            'REPO_SYNC',
            'USER_CREDENTIAL',
            'SITE_CREDENTIAL'
        );
    END IF;
END
$$;

ALTER TABLE IF EXISTS github_apps
    ADD COLUMN IF NOT EXISTS kind github_app_kind DEFAULT 'REPO_SYNC';

UPDATE github_apps
SET kind = 'COMMIT_SIGNING'
WHERE domain = 'batches';

-- This is expected to fail if any row is using an unknown value that is not repos or batches.
-- We only allow repos or batches at this time.
ALTER TABLE IF EXISTS github_apps
ALTER COLUMN kind SET NOT NULL;
