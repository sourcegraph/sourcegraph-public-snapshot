BEGIN;

ALTER TABLE gitserver_repos DROP COLUMN last_changed;

COMMIT;
