
BEGIN;

ALTER TABLE lsif_dependency_indexing_jobs
RENAME TO lsif_dependency_syncing_jobs;

CREATE TABLE IF NOT EXISTS lsif_dependency_indexing_jobs (
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
    external_service_kind text NOT NULL DEFAULT '',
    external_service_sync timestamp with time zone
);

COMMENT ON COLUMN lsif_dependency_indexing_jobs.external_service_kind IS 'Filter the external services for this kind to wait to have synced. If empty, external_service_sync is ignored and no external services are polled for their last sync time.';
COMMENT ON COLUMN lsif_dependency_indexing_jobs.external_service_sync IS 'The sync time after which external services of the given kind will have synced/created any repositories referenced by the LSIF upload that are resolvable.';

COMMIT;
