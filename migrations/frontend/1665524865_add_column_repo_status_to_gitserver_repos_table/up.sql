ALTER TABLE gitserver_repos
      -- repo_status indicates the "live" status of a cloned repo and
      -- is to be used in addition to clone_status. For example, it
      -- may indicate if a repo is currently being "fetched" or if it
      -- is currently being garbage collcted by one of the cleanup
      -- jobs.
      ADD COLUMN IF NOT EXISTS repo_status text;
