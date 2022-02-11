BEGIN;

ALTER TABLE dashboard_insight_view
DROP CONSTRAINT unique_dashboard_id_insight_view_id;

COMMIT;
