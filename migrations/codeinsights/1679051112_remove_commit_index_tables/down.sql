CREATE TABLE IF NOT EXISTS commit_index (
    committed_at timestamp with time zone NOT NULL,
    repo_id integer NOT NULL,
    commit_bytea bytea NOT NULL,
    indexed_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    debug_field text
);

CREATE TABLE IF NOT EXISTS commit_index_metadata (
   repo_id integer NOT NULL,
   enabled boolean DEFAULT true NOT NULL,
   last_indexed_at timestamp with time zone DEFAULT '1900-01-01 00:00:00+00'::timestamp with time zone NOT NULL
);

ALTER TABLE IF EXISTS commit_index_metadata DROP CONSTRAINT IF EXISTS commit_index_metadata_pkey;
ALTER TABLE IF EXISTS ONLY commit_index_metadata
    ADD CONSTRAINT commit_index_metadata_pkey PRIMARY KEY (repo_id);

ALTER TABLE IF EXISTS commit_index DROP CONSTRAINT IF EXISTS commit_index_pkey;
ALTER TABLE IF EXISTS ONLY commit_index
    ADD CONSTRAINT commit_index_pkey PRIMARY KEY (committed_at, repo_id, commit_bytea);

CREATE INDEX IF NOT EXISTS commit_index_repo_id_idx ON commit_index USING btree (repo_id, committed_at);