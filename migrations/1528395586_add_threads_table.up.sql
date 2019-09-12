BEGIN;

CREATE TABLE IF NOT EXISTS threads (
  id bigserial PRIMARY KEY,
  campaign_id integer NOT NULL REFERENCES campaigns(id)
    ON DELETE CASCADE DEFERRABLE INITIALLY IMMEDIATE,
  repo_id integer NOT NULL REFERENCES repo(id)
    ON DELETE CASCADE DEFERRABLE INITIALLY IMMEDIATE,
  created_at timestamp with time zone NOT NULL DEFAULT now(),
  updated_at timestamp with time zone NOT NULL DEFAULT now(),
  metadata jsonb NOT NULL DEFAULT '{}' CHECK (jsonb_typeof(metadata) = 'object')
);

COMMIT;
