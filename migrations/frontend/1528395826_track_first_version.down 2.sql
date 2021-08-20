BEGIN;

ALTER TABLE versions DROP COLUMN first_version;

COMMIT;
