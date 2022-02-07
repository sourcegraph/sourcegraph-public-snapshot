BEGIN;

CREATE TABLE commit_index
(
	committed_at TIMESTAMPTZ NOT NULL,
	repo_id INT NOT NULL,
	commit_bytea bytea NOT NULL,

	PRIMARY KEY (committed_at, repo_id, commit_bytea)
);

CREATE INDEX commit_index_repo_id_idx ON commit_index USING btree (repo_id, committed_at);

CREATE TABLE commit_index_metadata
(
    repo_id INT NOT NULL PRIMARY KEY,
    enabled BOOLEAN NOT NULL DEFAULT 'y',
    last_indexed_at TIMESTAMPTZ NOT NULL DEFAULT '1900-01-01'
);

COMMIT;
