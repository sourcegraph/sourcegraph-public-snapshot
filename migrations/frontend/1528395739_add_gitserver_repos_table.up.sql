BEGIN;

CREATE TYPE gitserver_repo_state AS ENUM (
    'not_cloned'
    'cloning',
    'cloned'
);

CREATE TABLE IF NOT EXISTS gitserver_repos (
  id int REFERENCES repo(id) PRIMARY KEY,
  shard_id text NOT NULL,
  state text NOT NULL, -- not cloned, cloning, cloned
  error text,
  updated_at timestamptz NOT NULL DEFAULT NOW()
);

COMMIT;
