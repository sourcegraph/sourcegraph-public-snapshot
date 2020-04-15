BEGIN;

ALTER TABLE global_state DROP COLUMN IF EXISTS initialized_pre_315;

COMMIT;
