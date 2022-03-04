CREATE TABLE IF NOT EXISTS rockskip_repos (
    id               SERIAL    PRIMARY KEY,
    repo             TEXT      NOT NULL,
    last_accessed_at TIMESTAMP WITH TIME ZONE NOT NULL,
    UNIQUE (repo)
);

CREATE TABLE IF NOT EXISTS rockskip_ancestry (
    id          SERIAL      PRIMARY KEY,
    repo_id     INTEGER     NOT NULL,
    commit_id   VARCHAR(40) NOT NULL,
    height      INTEGER     NOT NULL,
    ancestor    INTEGER     NOT NULL,
    UNIQUE (repo_id, commit_id)
);

-- Insert the null commit. repo_id 0 will not conflict with other repos because SERIAL's MINVALUE
-- defaults to 1.
INSERT INTO rockskip_ancestry
       (id, commit_id                                 , repo_id    , height, ancestor)
VALUES (0 , '0000000000000000000000000000000000000000', 0          , 0     , 0       )
ON CONFLICT DO NOTHING;

CREATE TABLE IF NOT EXISTS rockskip_symbols (
    -- Globally unique ID of this instance of the symbol.
    id           SERIAL        PRIMARY KEY,
    added        INTEGER[]     NOT NULL,
    deleted      INTEGER[]     NOT NULL,

    -- Since we only support searching by symbol name and we re-parse the file at query time, symbols
    -- with the same name in the same file only need to be stored once. Upon re-parsing the file at query
    -- time we will discover all symbols that match.
    repo_id      INTEGER       NOT NULL,
    path         TEXT          NOT NULL,
    name         TEXT          NOT NULL
);

CREATE OR REPLACE FUNCTION singleton(value TEXT) RETURNS TEXT[] AS $$ BEGIN
    RETURN ARRAY[value];
END; $$ IMMUTABLE language plpgsql;

CREATE OR REPLACE FUNCTION singleton_integer(value INTEGER) RETURNS INTEGER[] AS $$ BEGIN
    RETURN ARRAY[value];
END; $$ IMMUTABLE language plpgsql;

CREATE OR REPLACE FUNCTION path_prefixes(path TEXT) RETURNS TEXT[] AS $$ BEGIN
    RETURN (
        SELECT array_agg(array_to_string(components[:len], '/')) prefixes
        FROM
            (SELECT regexp_split_to_array(path, E'/') components) t,
            generate_series(1, array_length(components, 1)) AS len
    );
END; $$ IMMUTABLE language plpgsql;

CREATE INDEX IF NOT EXISTS rockskip_repos_repo ON rockskip_repos(repo);

CREATE INDEX IF NOT EXISTS rockskip_repos_last_accessed_at ON rockskip_repos(last_accessed_at);

CREATE INDEX IF NOT EXISTS rockskip_ancestry_repo_commit_id ON rockskip_ancestry(repo_id, commit_id);

CREATE INDEX IF NOT EXISTS rockskip_symbols_repo_id_path_name ON rockskip_symbols(repo_id, path, name);

CREATE INDEX IF NOT EXISTS rockskip_symbols_gin ON rockskip_symbols USING GIN (
    singleton_integer(repo_id) gin__int_ops,
    added gin__int_ops,
    deleted gin__int_ops,
    singleton(path),
    path_prefixes(path),
    singleton(name),
    name gin_trgm_ops
);
