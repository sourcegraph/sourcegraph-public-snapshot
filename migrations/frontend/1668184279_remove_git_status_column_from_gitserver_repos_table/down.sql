-- Undo the changes made in the up migration
ALTER TABLE gitserver_repos
ADD COLUMN IF NOT EXISTS repo_status text;
