ALTER TABLE local_repos ADD COLUMN org_id INTEGER;
ALTER TABLE local_repos DROP CONSTRAINT local_repos_remote_uri_access_token_key;