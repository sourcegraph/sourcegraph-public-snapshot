
BEGIN;

-- Insert migration here. See README.md. Highlights:
--  * Always use IF EXISTS. eg: DROP TABLE IF EXISTS global_dep_private;
--  * All migrations must be backward-compatible. Old versions of Sourcegraph
--    need to be able to read/write post migration.
--  * Historically we advised against transactions since we thought the
--    migrate library handled it. However, it does not! /facepalm

-- This table is a queue of records that need processing for code insights. Historically this queue grows
-- unbounded because the historical backfiller operated without state - every time it executed it would
-- requeue all of the work again. Since then we have added enough state to the backfiller to remove this problem,
-- but customer instances are going to be full of millions of records that will need processing before we can start
-- fresh. To avoid this problem, we are going to ship a 'reset' in 3.31 that will clear this queue entirely.
-- Note: This data is by design ephemeral, so there is no risk of permanent data loss here.

TRUNCATE insights_query_runner_jobs CASCADE;
COMMIT;
