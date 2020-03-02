BEGIN;

ALTER TABLE campaigns ADD COLUMN closed_at timestamptz;

COMMIT;
