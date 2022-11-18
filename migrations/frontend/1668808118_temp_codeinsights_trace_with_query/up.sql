ALTER TABLE IF EXISTS insights_query_runner_jobs
    ADD COLUMN IF NOT EXISTS trace_id TEXT;
