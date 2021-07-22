
BEGIN;

ALTER TABLE gitserver_repos ADD COLUMN last_fetched timestamp with time zone;

COMMIT;
