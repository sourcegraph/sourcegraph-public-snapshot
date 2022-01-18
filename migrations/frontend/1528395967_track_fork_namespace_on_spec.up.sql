-- +++
-- parent: 1528395966
-- +++

BEGIN;

ALTER TABLE
  changeset_specs
ADD COLUMN IF NOT EXISTS
  fork_namespace CITEXT NULL;

COMMIT;
