-- +++
-- parent: 1528395862
-- +++

BEGIN;

ALTER TABLE IF EXISTS event_logs ADD COLUMN IF NOT EXISTS public_argument JSONB DEFAULT '{}'::jsonb NOT NULL;

COMMIT;
