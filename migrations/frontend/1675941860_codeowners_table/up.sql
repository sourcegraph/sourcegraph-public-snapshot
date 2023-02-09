-- Perform migration here.
--
-- See /migrations/README.md. Highlights:
--  * Make migrations idempotent (use IF EXISTS)
--  * Make migrations backwards-compatible (old readers/writers must continue to work)
--  * If you are using CREATE INDEX CONCURRENTLY, then make sure that only one statement
--    is defined per file, and that each such statement is NOT wrapped in a transaction.
--    Each such migration must also declare "createIndexConcurrently: true" in their
--    associated metadata.yaml file.
--  * If you are modifying Postgres extensions, you must also declare "privileged: true"
--    in the associated metadata.yaml file.

CREATE TABLE IF NOT EXISTS own_blame_jobs (
  id                SERIAL PRIMARY KEY,
  state             text DEFAULT 'queued',
  failure_message   text,
  queued_at         timestamp with time zone DEFAULT NOW(),
  started_at        timestamp with time zone,
  finished_at       timestamp with time zone,
  process_after     timestamp with time zone,
  num_resets        integer not null default 0,
  num_failures      integer not null default 0,
  last_heartbeat_at timestamp with time zone,
  execution_logs    json[],
  worker_hostname   text not null default '',
  cancel            boolean not null default false,

  repository_id integer not null,
  absolute_file_path text not null
);

CREATE OR REPLACE VIEW own_blame_jobs_with_repository_name AS
  SELECT j.*, r.name
  FROM own_blame_jobs j
  JOIN repo r ON r.id = j.repository_id;

CREATE TABLE IF NOT EXISTS own_artifacts (
  id serial PRIMARY KEY,
  repo_id integer NOT NULL REFERENCES repo(id) ON DELETE CASCADE DEFERRABLE,
  absolute_path text NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS own_artifacts_index_repo_path ON own_artifacts USING btree (repo_id, absolute_path);

CREATE TABLE IF NOT EXISTS own_signals (
  id serial PRIMARY KEY,
  artifact_id integer NOT NULL REFERENCES own_artifacts(id) ON DELETE CASCADE DEFERRABLE,
  who text NOT NULL,
  method text NOT NULL,
  importance_indicator int NOT NULL,
  updated_at timestamp with time zone DEFAULT now() NOT NULL,
  created_at timestamp with time zone DEFAULT now() NOT NULL
);
