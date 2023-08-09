CREATE TABLE IF NOT EXISTS repo_clone_jobs
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
    gitserver_address TEXT    NOT NULL,
    repo_name         TEXT    NOT NULL,
    update_after      INTEGER NOT NULL,
    clone             BOOLEAN NOT NULL         DEFAULT FALSE
);

CREATE INDEX IF NOT EXISTS repo_clone_jobs_state_gitserver_address_idx ON repo_clone_jobs (state, gitserver_address);
CREATE INDEX IF NOT EXISTS repo_clone_jobs_repo_name_idx ON repo_clone_jobs (repo_name);
