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

-- Clear the dirty flag in case the operator timed out and isn't around to clear it.
UPDATE schema_migrations SET dirty = 'f'
COMMIT;
