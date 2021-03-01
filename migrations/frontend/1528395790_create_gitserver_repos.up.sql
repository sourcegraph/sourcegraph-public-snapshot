BEGIN;

CREATE TABLE IF NOT EXISTS gitserver_repos
(
    repo_id int REFERENCES repo(id) PRIMARY KEY,
    clone_status text NOT NULL default 'not_cloned',
    last_external_service bigint,
    shard_id text NOT NULL,
    last_error text,
    updated_at timestamp WITH TIME ZONE default now() not null
);

COMMIT;
