ALTER TABLE gitserver_repos
      ADD COLUMN IF NOT EXISTS repo_status text;
