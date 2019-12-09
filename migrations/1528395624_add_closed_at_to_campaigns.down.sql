BEGIN;

ALTER TABLE campaigns DROP COLUMN closed_at;

COMMIT;
