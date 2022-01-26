-- +++
-- parent: 1000000018
-- +++

BEGIN;

ALTER TABLE insight_view
    ADD COLUMN IF NOT EXISTS other_threshold FLOAT4;

COMMENT ON COLUMN insight_view.other_threshold IS 'Percent threshold for grouping series under "other"';

COMMIT;
