BEGIN;

ALTER TABLE query_runner_state RENAME TO saved_queries;

DROP TABLE IF EXISTS "saved_searches";
DROP TYPE IF EXISTS user_or_org;

-- Insert migration here. See README.md. Highlights:
--  * Always use IF EXISTS. eg: DROP TABLE IF EXISTS global_dep_private;
--  * All migrations must be backward-compatible. Old versions of Sourcegraph
--    need to be able to read/write post migration.
--  * Historically we advised against transactions since we thought the
--    migrate library handled it. However, it does not! /facepalm

COMMIT;
