BEGIN;

-- todo: this is not exactly backwards compatible

DROP TABLE IF EXISTS agents;

DROP TABLE IF EXISTS action_jobs;

DROP TABLE IF EXISTS action_executions;

DROP TABLE IF EXISTS actions;

COMMIT;
