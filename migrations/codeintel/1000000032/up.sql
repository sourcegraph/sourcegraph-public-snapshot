BEGIN;

CREATE TABLE rockskip_ancestry (
    commit_id   VARCHAR(40) PRIMARY KEY,
    repo        TEXT        NOT NULL,
    height      INTEGER     NOT NULL,
    ancestor_id VARCHAR(40) NOT NULL
);

CREATE TABLE rockskip_blobs (
    id           SERIAL        PRIMARY KEY,
    repo         TEXT          NOT NULL,
    commit_id    VARCHAR(40)   NOT NULL,
    path         TEXT          NOT NULL,
    added        VARCHAR(40)[] NOT NULL,
    deleted      VARCHAR(40)[] NOT NULL,
    symbol_names TEXT[]        NOT NULL,
    symbol_data  JSONB         NOT NULL
);

CREATE TABLE rockskip_repos (
    repo             TEXT PRIMARY KEY,
    last_accessed_at TIMESTAMP WITH TIME ZONE NOT NULL
);

CREATE INDEX rockskip_blobs_path ON rockskip_blobs(path);

CREATE INDEX rockskip_blobs_added_deleted_symbol_names ON rockskip_blobs USING GIN (added, deleted, symbol_names);

CREATE INDEX rockskip_repos_repo ON rockskip_repos(repo);

CREATE INDEX rockskip_repos_last_accessed_at ON rockskip_repos(last_accessed_at);

COMMIT;
