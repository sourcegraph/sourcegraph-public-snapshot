
BEGIN;

ALTER TABLE gitserver_repos DROP COLUMN last_fetched;

COMMIT;
