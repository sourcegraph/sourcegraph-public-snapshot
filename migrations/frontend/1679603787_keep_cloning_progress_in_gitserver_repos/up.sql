ALTER TABLE IF EXISTS gitserver_repos
ADD COLUMN IF NOT EXISTS cloning_progress text DEFAULT '';
