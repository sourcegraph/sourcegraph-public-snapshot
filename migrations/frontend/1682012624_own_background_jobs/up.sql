CREATE TABLE IF NOT EXISTS own_background_jobs
(
    id                SERIAL PRIMARY KEY,
    state             TEXT                     DEFAULT 'queued',
    failure_message   TEXT,
    queued_at         TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    started_at        TIMESTAMP WITH TIME ZONE,
    finished_at       TIMESTAMP WITH TIME ZONE,
    process_after     TIMESTAMP WITH TIME ZONE,
    num_resets        INTEGER NOT NULL         DEFAULT 0,
    num_failures      INTEGER NOT NULL         DEFAULT 0,
    last_heartbeat_at TIMESTAMP WITH TIME ZONE,
    execution_logs    JSON[],
    worker_hostname   TEXT    NOT NULL         DEFAULT '',
    cancel            BOOLEAN NOT NULL         DEFAULT FALSE,
    repo_id           INT     NOT NULL,
    job_type          INT     NOT NULL
);

CREATE INDEX IF NOT EXISTS own_background_jobs_state_idx ON own_background_jobs (state);
CREATE INDEX IF NOT EXISTS own_background_jobs_repo_id_idx ON own_background_jobs (repo_id);
