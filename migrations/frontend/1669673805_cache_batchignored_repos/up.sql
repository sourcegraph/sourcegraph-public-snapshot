CREATE TABLE IF NOT EXISTS
  batch_changes_repo_metadata
  (
    repo_id INTEGER NOT NULL PRIMARY KEY REFERENCES repo (id) ON DELETE CASCADE ON UPDATE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    ignored BOOLEAN NOT NULL DEFAULT FALSE
  );

CREATE INDEX IF NOT EXISTS
  batch_changes_repo_metadata_updated_at_idx
ON
  batch_changes_repo_metadata (updated_at);
