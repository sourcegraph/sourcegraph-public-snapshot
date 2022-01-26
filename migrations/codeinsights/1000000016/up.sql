-- +++
-- parent: 1000000015
-- +++

BEGIN;

-- Remove any already existing duplicates.
DELETE FROM
    dashboard_insight_view a
        USING dashboard_insight_view b
WHERE
	a.id > b.id
    AND a.dashboard_id = b.dashboard_id
    AND a.insight_view_id = b.insight_view_id;

ALTER TABLE dashboard_insight_view
ADD CONSTRAINT unique_dashboard_id_insight_view_id
UNIQUE (dashboard_id, insight_view_id);

COMMIT;
