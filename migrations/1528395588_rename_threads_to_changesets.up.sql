BEGIN;

DROP TABLE IF EXISTS threads;

CREATE TABLE IF NOT EXISTS changesets (
  id bigserial PRIMARY KEY,
  campaign_ids jsonb NOT NULL DEFAULT '{}'
    CHECK (jsonb_typeof(campaign_ids) = 'object'),
  repo_id integer NOT NULL REFERENCES repo(id)
    ON DELETE CASCADE DEFERRABLE INITIALLY IMMEDIATE,
  created_at timestamp with time zone NOT NULL DEFAULT now(),
  updated_at timestamp with time zone NOT NULL DEFAULT now(),
  metadata jsonb NOT NULL DEFAULT '{}'
    CHECK (jsonb_typeof(metadata) = 'object')
);

ALTER TABLE campaigns RENAME COLUMN thread_ids TO changeset_ids;
ALTER INDEX campaigns_thread_ids_gin_idx RENAME TO campaigns_changeset_ids_gin_idx;
ALTER TABLE campaigns RENAME CONSTRAINT campaigns_thread_ids_check TO campaigns_changeset_ids_check;

COMMIT;
