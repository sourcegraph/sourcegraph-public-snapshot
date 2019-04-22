BEGIN;

ALTER TABLE saved_queries RENAME TO query_runner_state;

CREATE TYPE user_or_org AS ENUM ('user', 'org');

CREATE TABLE IF NOT EXISTS "saved_searches" (
    "id" serial NOT NULL PRIMARY KEY,
    "description" text NOT NULL,
    "query" text NOT NULL,
    "created_at" timestamp with time zone DEFAULT now(),
    "updated_at" timestamp with time zone DEFAULT now(),
    "notify_owner" boolean,
    "notify_slack" boolean,
    "owner_kind" user_or_org NOT NULL,
    "user_id" integer REFERENCES users (id),
    "org_id" integer REFERENCES orgs (id),
    "slack_webhook_url" text
);


-- Insert migration here. See README.md. Highlights:
--  * Always use IF EXISTS. eg: DROP TABLE IF EXISTS global_dep_private;
--  * All migrations must be backward-compatible. Old versions of Sourcegraph
--    need to be able to read/write post migration.
--  * Historically we advised against transactions since we thought the
--    migrate library handled it. However, it does not! /facepalm

COMMIT;
