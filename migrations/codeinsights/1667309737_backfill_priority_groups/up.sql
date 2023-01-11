CREATE OR REPLACE VIEW insights_jobs_backfill_in_progress AS
SELECT
    jobs.*,
    isb.state AS backfill_state,
    isb.estimated_cost,
    width_bucket(isb.estimated_cost, 0, max(isb.estimated_cost+1) over (), 4) cost_bucket
FROM insights_background_jobs jobs
         JOIN insight_series_backfill isb ON jobs.backfill_id = isb.id
WHERE isb.state = 'processing';
