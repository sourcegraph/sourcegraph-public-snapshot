BEGIN;

CREATE TABLE IF NOT EXISTS actions (
    id SERIAL PRIMARY KEY,
    campaign integer REFERENCES campaigns(id) ON UPDATE CASCADE,
    schedule text,
    cancel_previous boolean NOT NULL DEFAULT false,
    saved_search integer REFERENCES saved_searches(id) ON UPDATE CASCADE,
    steps text NOT NULL,
    env json NOT NULL DEFAULT '[]'::json
);
CREATE UNIQUE INDEX IF NOT EXISTS actions_pkey ON actions(id int4_ops);

CREATE TABLE IF NOT EXISTS action_executions (
    id SERIAL PRIMARY KEY,
    steps text NOT NULL,
    env json,
    invokation_reason text NOT NULL,
    campaign_plan integer REFERENCES campaign_plans(id) ON UPDATE CASCADE,
    action integer NOT NULL REFERENCES actions(id) ON UPDATE CASCADE
);
CREATE UNIQUE INDEX IF NOT EXISTS action_executions_pkey ON action_executions(id int4_ops);

CREATE TABLE IF NOT EXISTS action_jobs (
    id SERIAL PRIMARY KEY,
    log text,
    execution_start timestamp with time zone,
    execution_end timestamp with time zone,
    runner_seen_at timestamp with time zone,
    patch text,
    state text NOT NULL DEFAULT 'PENDING'::text,
    repository integer NOT NULL REFERENCES repo(id) ON UPDATE CASCADE
);
CREATE UNIQUE INDEX IF NOT EXISTS action_jobs_pkey ON action_jobs(id int4_ops);

COMMIT;
