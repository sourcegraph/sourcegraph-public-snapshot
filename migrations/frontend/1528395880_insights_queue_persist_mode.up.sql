BEGIN;

-- Insert migration here. See README.md. Highlights:
--  * Always use IF EXISTS. eg: DROP TABLE IF EXISTS global_dep_private;
--  * All migrations must be backward-compatible. Old versions of Sourcegraph
--    need to be able to read/write post migration.
--  * Historically we advised against transactions since we thought the
--    migrate library handled it. However, it does not! /facepalm

CREATE TYPE PersistMode AS ENUM ('record', 'snapshot');

ALTER TABLE insights_query_runner_jobs
    ADD persist_mode PersistMode DEFAULT 'record' NOT NULL;

COMMENT ON COLUMN insights_query_runner_jobs.persist_mode IS 'The persistence level for this query. This value will determine the lifecycle of the resulting value.';
COMMIT;
