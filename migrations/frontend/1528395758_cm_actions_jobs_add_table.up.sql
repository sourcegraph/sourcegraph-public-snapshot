BEGIN;
CREATE TABLE IF NOT EXISTS cm_action_jobs
(
    id              SERIAL PRIMARY KEY,
    email           int8 NOT NULL,
    state           text default 'queued',
    failure_message text,
    started_at      timestamptz,
    finished_at     timestamptz,
    process_after   timestamptz,
    num_resets      int4 NOT NULL default 0,
    num_failures    int4 NOT NULL default 0,
    log_contents    text,
    CONSTRAINT cm_action_jobs_email_fk
        FOREIGN KEY (email)
            REFERENCES cm_emails (id)
            ON DELETE CASCADE
);
COMMIT;
