CREATE INDEX IF NOT EXISTS process_after_insights_query_runner_jobs_idx
    ON insights_query_runner_jobs (process_after);

CREATE INDEX IF NOT EXISTS finished_at_insights_query_runner_jobs_idx
    ON insights_query_runner_jobs (finished_at);
