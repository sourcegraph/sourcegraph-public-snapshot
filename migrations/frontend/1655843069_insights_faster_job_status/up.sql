CREATE INDEX CONCURRENTLY IF NOT EXISTS insights_query_runner_jobs_series_id_state ON insights_query_runner_jobs (series_id, state);
