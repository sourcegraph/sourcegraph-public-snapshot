BEGIN;

CREATE TABLE query_runner_state AS SELECT * FROM saved_queries;

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

COMMIT;
