BEGIN;

ALTER TABLE
    changeset_specs
DROP COLUMN IF EXISTS
    raw_spec;

COMMIT;
