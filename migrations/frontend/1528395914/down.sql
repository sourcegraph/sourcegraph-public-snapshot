BEGIN;

ALTER TABLE batch_specs
  DROP COLUMN IF EXISTS created_from_raw;

COMMIT;
