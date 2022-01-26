-- +++
-- parent: 1528395889
-- +++

BEGIN;

ALTER TABLE
    repo
DROP COLUMN IF EXISTS cloned;

COMMIT;
