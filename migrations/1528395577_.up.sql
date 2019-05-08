BEGIN;

CREATE TABLE query_runner_state AS SELECT * FROM saved_queries;

CREATE TABLE IF NOT EXISTS "saved_searches" (
    "id" serial NOT NULL PRIMARY KEY,
    "description" text NOT NULL,
    "query" text NOT NULL,
    "created_at" timestamp with time zone DEFAULT now(),
    "updated_at" timestamp with time zone DEFAULT now(),
    "notify_owner" boolean,
    "notify_slack" boolean,
    "user_id" integer REFERENCES users (id),
    "org_id" integer REFERENCES orgs (id),
    "slack_webhook_url" text
);

COMMIT;
