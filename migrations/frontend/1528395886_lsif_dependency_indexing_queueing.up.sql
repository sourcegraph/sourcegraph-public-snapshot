
BEGIN;

CREATE TABLE IF NOT EXISTS lsif_dependency_indexing_queueing_jobs (
    id serial PRIMARY KEY,
    state text DEFAULT 'queued' NOT NULL,
    failure_message text,
    queued_at timestamp with time zone DEFAULT NOW() NOT NULL,
    started_at timestamp with time zone,
    finished_at timestamp with time zone,
    process_after timestamp with time zone,
    num_resets integer DEFAULT 0 NOT NULL,
    num_failures integer DEFAULT 0 NOT NULL,
    execution_logs json[],
    last_heartbeat_at timestamp with time zone,
    worker_hostname text NOT NULL DEFAULT '',
    upload_id integer REFERENCES lsif_uploads(id) ON DELETE CASCADE,
    external_service_kind text,
    external_service_sync timestamp with time zone
);

COMMIT;
