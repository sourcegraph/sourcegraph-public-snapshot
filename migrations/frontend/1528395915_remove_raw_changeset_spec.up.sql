-- +++
-- parent: 1528395914
-- +++

BEGIN;

ALTER TABLE
    changeset_specs
DROP COLUMN IF EXISTS
    raw_spec;

COMMIT;
