ALTER TABLE gitserver_repos
	ADD COLUMN IF NOT EXISTS pool_repo_id INTEGER DEFAULT NULL;

ALTER TABLE gitserver_repos
    ADD CONSTRAINT pool_repo_id_fkey FOREIGN KEY (pool_repo_id) REFERENCES repo(id);

COMMENT ON COLUMN gitserver_repos.pool_repo_id IS 'This is used to refer to the pool repository for deduplicated repos';
