
-- drop view
DROP VIEW IF EXISTS gitserver_localclone_jobs_with_repo_name;

-- drop the column
ALTER TABLE gitserver_localclone_jobs DROP COLUMN IF EXISTS queued_at;

-- recreate the view without the column
CREATE OR REPLACE VIEW gitserver_localclone_jobs_with_repo_name AS
  SELECT glj.*, r.name AS repo_name
  FROM gitserver_localclone_jobs glj
  JOIN repo r ON r.id = glj.repo_id;
