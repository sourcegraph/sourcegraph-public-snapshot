CREATE TABLE IF NOT EXISTS codeintel_commit_dates(
    repository_id integer NOT NULL,
    commit_bytea bytea NOT NULL,
    committed_at text NOT NULL,
    PRIMARY KEY(repository_id, commit_bytea)
);
