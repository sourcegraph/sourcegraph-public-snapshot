BEGIN;

CREATE TABLE changeset_jobs (
  id bigserial PRIMARY KEY,

  campaign_id bigint NOT NULL REFERENCES campaigns(id)
    DEFERRABLE INITIALLY IMMEDIATE,

  campaign_job_id bigint NOT NULL REFERENCES campaign_jobs(id)
    DEFERRABLE INITIALLY IMMEDIATE,

  changeset_id bigint REFERENCES changesets(id)
    DEFERRABLE INITIALLY IMMEDIATE,

  error text,

  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),

  started_at timestamptz,
  finished_at timestamptz
);

ALTER TABLE changeset_jobs ADD CONSTRAINT changeset_jobs_unique
UNIQUE (campaign_id, campaign_job_id);

CREATE INDEX changeset_jobs_started_at ON changeset_jobs(started_at);
CREATE INDEX changeset_jobs_finished_at ON changeset_jobs(finished_at);
CREATE INDEX changeset_jobs_error ON changeset_jobs(error);

COMMIT;
