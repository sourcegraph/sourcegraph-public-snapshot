BEGIN;

ALTER TABLE dashboard ADD COLUMN IF NOT EXISTS save BOOLEAN;

COMMENT ON COLUMN dashboard.save IS 'TEMPORARY Do not delete this dashboard when migrating settings.';

COMMIT;
