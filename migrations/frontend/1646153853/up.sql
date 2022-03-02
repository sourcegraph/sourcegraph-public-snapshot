BEGIN;

ALTER TABLE IF EXISTS external_service_repos
    ADD COLUMN IF NOT EXISTS created_at timestamp with time zone DEFAULT transaction_timestamp();

COMMIT;
