BEGIN;

ALTER TABLE insight_view
DROP COLUMN IF EXISTS other_threshold;

COMMIT;
