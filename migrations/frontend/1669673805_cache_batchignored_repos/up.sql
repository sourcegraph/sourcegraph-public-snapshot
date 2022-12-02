CREATE TABLE IF NOT EXISTS batch_changes_repo_metadata
(
  -- Actual fields related to metadata.
  repo_id INTEGER NOT NULL PRIMARY KEY REFERENCES repo (id) ON DELETE CASCADE ON UPDATE CASCADE,
  created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
  ignored BOOLEAN NOT NULL DEFAULT FALSE,

  -- Worker fields.
  state TEXT NOT NULL DEFAULT 'queued',
  failure_message TEXT,
  queued_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
  started_at TIMESTAMP WITH TIME ZONE,
  finished_at TIMESTAMP WITH TIME ZONE,
  process_after TIMESTAMP WITH TIME ZONE,
  num_resets INTEGER NOT NULL DEFAULT 0,
  num_failures INTEGER NOT NULL DEFAULT 0,
  last_heartbeat_at TIMESTAMP WITH TIME ZONE,
  execution_logs JSON[],
  worker_hostname TEXT NOT NULL DEFAULT '',
  cancel BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE INDEX IF NOT EXISTS
  batch_changes_repo_metadata_updated_at_idx
ON
  batch_changes_repo_metadata (updated_at);

CREATE OR REPLACE VIEW batch_changes_repo_metadata_with_repo_name AS
  SELECT
    batch_changes_repo_metadata.*,
    repo.name
  FROM
    batch_changes_repo_metadata
  INNER JOIN
    repo
  ON
    repo.id = batch_changes_repo_metadata.repo_id;

INSERT INTO
  batch_changes_repo_metadata
  (repo_id)
SELECT
  id
FROM
  repo
WHERE
  deleted_at IS NULL;
