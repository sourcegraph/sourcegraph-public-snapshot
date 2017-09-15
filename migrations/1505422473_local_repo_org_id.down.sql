ALTER TABLE local_repos DROP COLUMN org_id;
ALTER TABLE local_repos ADD CONSTRAINT local_repos_remote_uri_access_token_key UNIQUE (remote_uri, access_token);