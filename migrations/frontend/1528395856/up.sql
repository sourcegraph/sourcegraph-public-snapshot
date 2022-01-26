-- +++
-- parent: 1528395855
-- +++

BEGIN;

ALTER TABLE IF EXISTS lsif_uploads ADD COLUMN IF NOT EXISTS execution_logs JSON[];

COMMIT;
