BEGIN;

ALTER TABLE versions ADD COLUMN first_version text;
UPDATE versions set first_version = version;
ALTER TABLE versions ALTER COLUMN first_version SET NOT NULL;

COMMIT;
