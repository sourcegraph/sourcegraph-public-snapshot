BEGIN;

ALTER TABLE global_state ADD COLUMN initialized_pre_315 boolean DEFAULT FALSE NOT NULL;

-- This migration first shipped with v3.15, so only instances initialized pre-v3.15 would already
-- have the initialized column true.
UPDATE global_state SET initialized_pre_315=true WHERE initialized;

COMMIT;
