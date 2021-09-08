BEGIN;

-- TODO: Remove _job suffix
CREATE TABLE IF NOT EXISTS batch_spec_workspace_jobs (
  id              BIGSERIAL PRIMARY KEY,

  batch_spec_id      INTEGER REFERENCES batch_specs(id) ON DELETE CASCADE DEFERRABLE,
  changeset_spec_ids JSONB DEFAULT '{}'::jsonb,

  repo_id integer      REFERENCES repo(id) DEFERRABLE,
  branch               TEXT NOT NULL,
  commit               TEXT NOT NULL,
  path                 TEXT NOT NULL,
  file_matches         TEXT[] NOT NULL,
  only_fetch_workspace BOOLEAN NOT NULL DEFAULT FALSE,
  steps                JSONB DEFAULT '[]'::jsonb,

  state             TEXT DEFAULT 'pending',
  failure_message   TEXT,
  started_at        TIMESTAMP WITH TIME ZONE,
  finished_at       TIMESTAMP WITH TIME ZONE,
  process_after     TIMESTAMP WITH TIME ZONE,
  num_resets        INTEGER NOT NULL DEFAULT 0,
  num_failures      INTEGER NOT NULL DEFAULT 0,
  execution_logs    JSON[],
  worker_hostname   TEXT NOT NULL DEFAULT '',
  last_heartbeat_at TIMESTAMP WITH TIME ZONE,

  created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);


ALTER TABLE IF EXISTS batch_specs
-- TODO: Do we need a default state when migrating from old model to new model?
  ADD COLUMN IF NOT EXISTS state           TEXT,

  ADD COLUMN IF NOT EXISTS failure_message   TEXT,
  ADD COLUMN IF NOT EXISTS started_at        TIMESTAMP WITH TIME ZONE,
  ADD COLUMN IF NOT EXISTS finished_at       TIMESTAMP WITH TIME ZONE,
  ADD COLUMN IF NOT EXISTS process_after     TIMESTAMP WITH TIME ZONE,
  ADD COLUMN IF NOT EXISTS last_heartbeat_at TIMESTAMP WITH TIME ZONE,
  ADD COLUMN IF NOT EXISTS num_resets        INTEGER NOT NULL DEFAULT 0,
  ADD COLUMN IF NOT EXISTS num_failures      INTEGER NOT NULL DEFAULT 0,
  ADD COLUMN IF NOT EXISTS execution_logs    JSON[],
  ADD COLUMN IF NOT EXISTS worker_hostname   TEXT NOT NULL DEFAULT ''

  -- TODO: add two "allow unsupported, ignored" options here as booleans
  ;

COMMIT;
