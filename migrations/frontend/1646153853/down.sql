BEGIN;

ALTER TABLE IF EXISTS external_service_repos
    DROP COLUMN IF EXISTS created_at;

COMMIT;
