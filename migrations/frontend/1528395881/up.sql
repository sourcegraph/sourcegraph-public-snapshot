-- +++
-- parent: 1528395880
-- +++

BEGIN;

CREATE TABLE IF NOT EXISTS batch_spec_resolution_jobs (
  id              BIGSERIAL PRIMARY KEY,

  batch_spec_id     INTEGER REFERENCES batch_specs(id) ON DELETE CASCADE DEFERRABLE,
  allow_unsupported BOOLEAN NOT NULL DEFAULT FALSE,
  allow_ignored     BOOLEAN NOT NULL DEFAULT FALSE,

  state             TEXT DEFAULT 'queued',
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

CREATE TABLE IF NOT EXISTS batch_spec_workspaces (
  id              BIGSERIAL PRIMARY KEY,

  batch_spec_id      INTEGER REFERENCES batch_specs(id) ON DELETE CASCADE DEFERRABLE,
  changeset_spec_ids JSONB DEFAULT '{}'::jsonb,

  repo_id integer      REFERENCES repo(id) DEFERRABLE,
  branch               TEXT NOT NULL,
  commit               TEXT NOT NULL,
  path                 TEXT NOT NULL,
  file_matches         TEXT[] NOT NULL,
  only_fetch_workspace BOOLEAN NOT NULL DEFAULT FALSE,
  steps                JSONB DEFAULT '[]'::jsonb CHECK (jsonb_typeof(steps) = 'array'),

  created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS batch_spec_workspace_execution_jobs (
  id              BIGSERIAL PRIMARY KEY,

  batch_spec_workspace_id  INTEGER REFERENCES batch_spec_workspaces(id) ON DELETE CASCADE DEFERRABLE,

  state             TEXT DEFAULT 'queued',
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

COMMIT;
