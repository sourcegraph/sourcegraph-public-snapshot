ALTER TABLE gitserver_localclone_jobs ADD COLUMN IF NOT EXISTS queued_at timestamptz DEFAULT NOW();

-- drop view and recreate it with the new column

DROP VIEW IF EXISTS gitserver_localclone_jobs_with_repo_name;

CREATE OR REPLACE VIEW gitserver_localclone_jobs_with_repo_name AS
  SELECT glj.*, r.name AS repo_name
  FROM gitserver_localclone_jobs glj
  JOIN repo r ON r.id = glj.repo_id;
