BEGIN;

CREATE TABLE rockskip_ancestry (
    id       VARCHAR(40) PRIMARY KEY,
    height   INTEGER     NOT NULL,
    ancestor VARCHAR(40) NOT NULL
);

CREATE TABLE rockskip_blobs (
    id           SERIAL        PRIMARY KEY,
    commit       VARCHAR(40)   NOT NULL,
    path         TEXT          NOT NULL,
    added        VARCHAR(40)[] NOT NULL,
    deleted      VARCHAR(40)[] NOT NULL,
    symbol_names TEXT[]        NOT NULL,
    symbol_data  JSONB         NOT NULL
);

CREATE INDEX rockskip_blobs_path ON rockskip_blobs(path);

CREATE INDEX rockskip_blobs_added_deleted_symbol_names ON rockskip_blobs USING GIN (added, deleted, symbol_names);

COMMIT;
