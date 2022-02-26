BEGIN;

CREATE TABLE rockskip_repos (
    repo             TEXT PRIMARY KEY,
    last_accessed_at TIMESTAMP WITH TIME ZONE NOT NULL
);

CREATE TABLE rockskip_ancestry (
    commit_id   VARCHAR(40),
    repo        TEXT        NOT NULL,
    height      INTEGER     NOT NULL,
    ancestor_id VARCHAR(40) NOT NULL,
    PRIMARY KEY (repo, commit_id)
);

CREATE TABLE rockskip_blobs (
    id           SERIAL        PRIMARY KEY,
    repo         TEXT          NOT NULL,
    commit_id    VARCHAR(40)   NOT NULL,
    path         TEXT          NOT NULL,
    added        VARCHAR(40)[] NOT NULL,
    deleted      VARCHAR(40)[] NOT NULL,
    symbol_names TEXT[]        NOT NULL,
    symbol_data  JSONB         NOT NULL,
    UNIQUE (repo, commit_id, path)
);

CREATE OR REPLACE FUNCTION singleton(value TEXT) RETURNS TEXT[] AS $$ BEGIN
    RETURN ARRAY[value];
END; $$ IMMUTABLE language plpgsql;

CREATE INDEX rockskip_repos_repo ON rockskip_repos(repo);

CREATE INDEX rockskip_repos_last_accessed_at ON rockskip_repos(last_accessed_at);

CREATE INDEX rockskip_blobs_gin ON rockskip_blobs USING GIN (
    singleton(repo),
    added,
    deleted,
    singleton(path),
    symbol_names
);

COMMIT;
