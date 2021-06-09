BEGIN;

CREATE TABLE commit_index
(
	commit_time TIMESTAMPTZ NOT NULL,
	repo_id INT NOT NULL,
	ref TEXT NOT NULL,

	PRIMARY KEY (commit_time, repo_id, ref)
);

CREATE INDEX commit_index_repo_id_idx ON commit_index USING btree (repo_id);

CREATE TABLE commit_index_metadata
(
    repo_id INT NOT NULL PRIMARY KEY,
    enabled BOOLEAN NOT NULL DEFAULT 'y',
    last_indexed_at TIMESTAMPTZ NOT NULL DEFAULT '1900-01-01'
);

COMMIT;
