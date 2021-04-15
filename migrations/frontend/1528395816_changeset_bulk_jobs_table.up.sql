BEGIN;

CREATE TABLE IF NOT EXISTS changeset_jobs (
    id BIGSERIAL PRIMARY KEY,
    bulk_group text NOT NULL,
    user_id integer NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    batch_change_id integer NOT NULL REFERENCES batch_changes(id) ON DELETE CASCADE,
    changeset_id integer NOT NULL REFERENCES changesets(id) ON DELETE CASCADE,

    job_type text NOT NULL,
    payload jsonb DEFAULT '{}'::jsonb CHECK (jsonb_typeof(payload) = 'object'::text),

    state text DEFAULT 'queued'::text,
    failure_message text,
    started_at timestamp with time zone,
    finished_at timestamp with time zone,
    process_after timestamp with time zone,
    num_resets integer NOT NULL DEFAULT 0,
    num_failures integer NOT NULL DEFAULT 0,
    execution_logs json[],

    created_at timestamp with time zone NOT NULL DEFAULT now(),
    updated_at timestamp with time zone NOT NULL DEFAULT now()
);

COMMIT;
