ALTER TABLE gitserver_repos
      DROP CONSTRAINT IF EXISTS pool_repo_id_fkey;

ALTER TABLE gitserver_repos
	DROP COLUMN IF EXISTS pool_repo_id;
