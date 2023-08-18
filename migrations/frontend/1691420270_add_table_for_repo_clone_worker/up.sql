CREATE TABLE IF NOT EXISTS repo_update_jobs
(
    id                      SERIAL PRIMARY KEY,
    state                   TEXT                     DEFAULT 'queued',
    failure_message         TEXT,
    queued_at               TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    started_at              TIMESTAMP WITH TIME ZONE,
    finished_at             TIMESTAMP WITH TIME ZONE,
    process_after           TIMESTAMP WITH TIME ZONE,
    num_resets              INTEGER NOT NULL         DEFAULT 0,
    num_failures            INTEGER NOT NULL         DEFAULT 0,
    last_heartbeat_at       TIMESTAMP WITH TIME ZONE,
    execution_logs          JSON[],
    worker_hostname         TEXT    NOT NULL         DEFAULT '',
    cancel                  BOOLEAN NOT NULL         DEFAULT FALSE,

    repo_id                 INTEGER NOT NULL REFERENCES repo (id) ON DELETE CASCADE,
    priority                INTEGER NOT NULL         DEFAULT 0,
    -- True if this is a clone job and we know it in advance. (Use case: reclone of an already cloned repo).
    clone                   BOOLEAN NOT NULL         DEFAULT FALSE,
    -- True if an existing clone should be overwritten.
    overwrite_clone         BOOLEAN NOT NULL         DEFAULT FALSE,
    -- If provided, then the fetch (repo update) will be performed for a given revision.
    fetch_revision          TEXT                     DEFAULT '',
    -- Populated after the job is processed.
    last_fetched            TIMESTAMP WITH TIME ZONE,
    -- Populated after the job is processed.
    last_changed            TIMESTAMP WITH TIME ZONE,
    -- Used for jobs scheduling.
    update_interval_seconds INTEGER
);

CREATE INDEX IF NOT EXISTS repo_update_jobs_state_idx ON repo_update_jobs (state);
-- Index used for jobs dequeue ordering.
CREATE INDEX IF NOT EXISTS repo_update_jobs_priority_process_after_idx ON repo_update_jobs (priority, process_after, id);

-- Only one queued repo ID at a time.
CREATE UNIQUE INDEX IF NOT EXISTS repo_update_jobs_repo_id_queued_idx ON repo_update_jobs (repo_id) WHERE state = 'queued';

DROP VIEW IF EXISTS repo_update_jobs_with_repo_name;

-- repo_update_jobs_with_repo_name view is used to dequeue repo update jobs and get information about repo name
CREATE VIEW repo_update_jobs_with_repo_name AS
SELECT j.id,
       j.state,
       j.failure_message,
       j.queued_at,
       j.started_at,
       j.finished_at,
       j.process_after,
       j.num_resets,
       j.num_failures,
       j.last_heartbeat_at,
       j.execution_logs,
       j.worker_hostname,
       j.cancel,
       j.repo_id,
       j.priority,
       j.clone,
       j.overwrite_clone,
       j.fetch_revision,
       j.last_fetched,
       j.last_changed,
       j.update_interval_seconds,
       r.name         AS repository_name,
       g.pool_repo_id AS pool_repo_id
FROM repo_update_jobs j
         JOIN gitserver_repos g ON g.repo_id = j.repo_id
         JOIN repo r ON r.id = COALESCE(g.pool_repo_id, j.repo_id)
WHERE r.deleted_at IS NULL;

-- addr_for_repo is used to resolve a gitserver instance address where a repo with a given repo_name should be located/cloned.
CREATE OR REPLACE FUNCTION addr_for_repo(repo_name TEXT, addrs TEXT[]) RETURNS TEXT AS
$$
DECLARE
    md5_hash     BYTEA;
    server_index BIGINT;
BEGIN
    -- Compute the MD5 hash of the repo name.
    md5_hash := decode(md5(repo_name), 'hex');

    -- Convert the first 8 bytes of the MD5 hash to a BIGINT.
    server_index := get_byte(md5_hash, 0)::BIGINT << 56 |
                    get_byte(md5_hash, 1)::BIGINT << 48 |
                    get_byte(md5_hash, 2)::BIGINT << 40 |
                    get_byte(md5_hash, 3)::BIGINT << 32 |
                    get_byte(md5_hash, 4)::BIGINT << 24 |
                    get_byte(md5_hash, 5)::BIGINT << 16 |
                    get_byte(md5_hash, 6)::BIGINT << 8 |
                    get_byte(md5_hash, 7)::BIGINT;

    -- Use modulo to get the index and fetch the address from the array.
    RETURN addrs[(server_index % array_length(addrs, 1)) + 1];
END;
$$ LANGUAGE plpgsql;
