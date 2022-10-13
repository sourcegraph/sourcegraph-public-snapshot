-- Undo the changes made in the up migration

ALTER TABLE gitserver_repos
      DROP COLUMN IF EXISTS repo_status;

