BEGIN;

-- Insert migration here. See README.md. Highlights:
--  * Always use IF EXISTS. eg: DROP TABLE IF EXISTS global_dep_private;
--  * All migrations must be backward-compatible. Old versions of Sourcegraph
--    need to be able to read/write post migration.
--  * Historically we advised against transactions since we thought the
--    migrate library handled it. However, it does not! /facepalm

CREATE TABLE IF NOT EXISTS cm_trigger_jobs
(
    id              SERIAL PRIMARY KEY,
    query           int8 NOT NULL,
    state           text default 'queued',
    failure_message text,
    started_at      timestamptz,
    finished_at     timestamptz,
    process_after   timestamptz,
    num_resets      int4 NOT NULL default 0,
    num_failures    int4 NOT NULL default 0,
    log_contents    text,
    CONSTRAINT cm_trigger_jobs_query_fk
        FOREIGN KEY (query)
            REFERENCES cm_queries (id)
            ON DELETE CASCADE
);
COMMIT;
