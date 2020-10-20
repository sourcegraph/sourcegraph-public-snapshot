BEGIN;

CREATE TYPE campaign_apply_job_state AS ENUM (
    'queued',
    'processing',
    'completed',
    'errored'
);

CREATE TABLE campaign_apply_jobs(
    id bigint NOT NULL primary key,
    queued_at timestamp with time zone NOT NULL DEFAULT now(),
    state campaign_apply_job_state NOT NULL DEFAULT 'queued',
    failure_message text,
    started_at timestamp with time zone,
    finished_at timestamp with time zone,
    process_after timestamp with time zone,
    num_resets integer NOT NULL DEFAULT 0,
    num_failures integer NOT NULL DEFAULT 0,
    yaml text NOT NULL,
    log_contents text
);

COMMIT;
