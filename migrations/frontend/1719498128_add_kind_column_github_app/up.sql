CREATE TYPE github_app_kind AS ENUM (
    'COMMIT_SIGNING',
    'REPO_SYNC',
    'USER_GITHUB_APP'
);

ALTER TABLE IF EXISTS github_apps
    ADD COLUMN IF NOT EXISTS kind github_app_kind NULL;

UPDATE github_apps
SET kind = 'COMMIT_SIGNING'
WHERE domain = 'batches';

UPDATE github_apps
SET kind = 'REPO_SYNC'
WHERE domain = 'repos';

-- This is expected to fail if any row is using an unknown value that is not repos or batches.
-- We only allow repos or batches at this time.
ALTER TABLE IF EXISTS github_apps
ALTER COLUMN kind SET NOT NULL;
