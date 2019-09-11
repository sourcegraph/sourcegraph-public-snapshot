BEGIN;

DROP TABLE IF EXISTS changesets;

CREATE TABLE IF NOT EXISTS threads (
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

ALTER TABLE campaigns RENAME COLUMN changeset_ids TO thread_ids;
ALTER INDEX campaigns_changeset_ids_gin_idx RENAME TO campaigns_thread_ids_gin_idx;
ALTER TABLE campaigns RENAME CONSTRAINT campaigns_changeset_ids_check TO campaigns_thread_ids_check;

COMMIT;
