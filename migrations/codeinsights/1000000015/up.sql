-- +++
-- parent: 1000000014
-- +++

BEGIN;

ALTER TABLE dashboard ADD COLUMN IF NOT EXISTS save BOOLEAN NOT NULL DEFAULT FALSE;

COMMENT ON COLUMN dashboard.save IS 'TEMPORARY Do not delete this dashboard when migrating settings.';

COMMIT;
