BEGIN;

CREATE TABLE campaign_jobs (
  id bigserial PRIMARY KEY,
  campaign_plan_id bigint NOT NULL REFERENCES campaign_plans(id)
    ON DELETE CASCADE DEFERRABLE INITIALLY IMMEDIATE,

  repo_id bigint NOT NULL REFERENCES repo(id)
    DEFERRABLE INITIALLY IMMEDIATE,

  rev text NOT NULL,

  diff text NOT NULL,
  error text NOT NULL,

  started_at timestamp with time zone,
  finished_at timestamp with time zone,

  created_at timestamp with time zone NOT NULL DEFAULT now(),
  updated_at timestamp with time zone NOT NULL DEFAULT now()
);

COMMIT;
