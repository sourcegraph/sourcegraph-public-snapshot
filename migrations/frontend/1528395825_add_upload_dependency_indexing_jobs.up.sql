BEGIN;

CREATE TABLE lsif_dependency_indexing_jobs (
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
    upload_id integer REFERENCES lsif_uploads(id) ON DELETE CASCADE
);

COMMENT ON TABLE lsif_dependency_indexing_jobs IS 'Tracks jobs that scan imports of indexes to schedule auto-index jobs.';
COMMENT ON COLUMN upload_id.id IS 'The identifier of the triggering upload record.';

COMMIT;
