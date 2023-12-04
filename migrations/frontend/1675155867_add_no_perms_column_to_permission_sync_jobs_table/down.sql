-- We are changing the schema and we don't expect any rows yet so we can just drop
-- the old table.
-- We are changing the schema and we don't expect any rows yet so we can just drop
-- the old table.
DROP TABLE IF EXISTS permission_sync_jobs;

CREATE TABLE permission_sync_jobs
(
    id                   SERIAL PRIMARY KEY,
    state                TEXT                     DEFAULT 'queued',
    reason               TEXT    NOT NULL,
    failure_message      TEXT,
    queued_at            TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    started_at           TIMESTAMP WITH TIME ZONE,
    finished_at          TIMESTAMP WITH TIME ZONE,
    process_after        TIMESTAMP WITH TIME ZONE,
    num_resets           INTEGER NOT NULL         DEFAULT 0,
    num_failures         INTEGER NOT NULL         DEFAULT 0,
    last_heartbeat_at    TIMESTAMP WITH TIME ZONE,
    execution_logs       JSON[],
    worker_hostname      TEXT    NOT NULL         DEFAULT '',
    cancel               BOOLEAN NOT NULL         DEFAULT FALSE,

    repository_id        INTEGER REFERENCES repo (id) ON DELETE CASCADE,
    user_id              INTEGER REFERENCES users (id) ON DELETE CASCADE,
    triggered_by_user_id INTEGER REFERENCES users (id) ON DELETE SET NULL DEFERRABLE,

    priority             INTEGER NOT NULL         DEFAULT 0,
    invalidate_caches    BOOLEAN NOT NULL         DEFAULT FALSE,
    cancellation_reason  TEXT,
    CONSTRAINT permission_sync_jobs_for_repo_or_user CHECK (((user_id IS NULL) <> (repository_id IS NULL)))
);

CREATE INDEX IF NOT EXISTS permission_sync_jobs_state ON permission_sync_jobs (state);
CREATE INDEX IF NOT EXISTS permission_sync_jobs_process_after ON permission_sync_jobs (process_after);
CREATE INDEX IF NOT EXISTS permission_sync_jobs_repository_id ON permission_sync_jobs (repository_id);
CREATE INDEX IF NOT EXISTS permission_sync_jobs_user_id ON permission_sync_jobs (user_id);

-- this index is used as a last resort if deduplication logic fails to work.
-- we should not enqueue more that one high priority immediate sync job (process_after IS NULL) for given repo/user.
CREATE UNIQUE INDEX IF NOT EXISTS permission_sync_jobs_unique ON permission_sync_jobs
    USING btree (priority, user_id, repository_id, cancel, process_after)
    WHERE (state = 'queued');

COMMENT ON COLUMN permission_sync_jobs.reason IS 'Specifies why permissions sync job was triggered.';
COMMENT ON COLUMN permission_sync_jobs.triggered_by_user_id IS 'Specifies an ID of a user who triggered a sync.';
COMMENT ON COLUMN permission_sync_jobs.cancellation_reason IS 'Specifies why permissions sync job was cancelled.';
COMMENT ON COLUMN permission_sync_jobs.priority IS 'Specifies numeric priority for the permissions sync job.';
