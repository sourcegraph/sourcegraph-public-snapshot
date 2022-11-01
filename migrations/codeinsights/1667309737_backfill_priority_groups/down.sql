-- Undo the changes made in the up migration
drop view if exists insights_jobs_backfill_in_progress;

CREATE OR REPLACE VIEW insights_jobs_backfill_in_progress AS
SELECT jobs.*, isb.state AS backfill_state, isb.estimated_cost
FROM insights_background_jobs jobs
         JOIN insight_series_backfill isb ON jobs.backfill_id = isb.id
WHERE isb.state = 'processing';
