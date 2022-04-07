DROP INDEX IF EXISTS rockskip_symbols_gin;

CREATE OR REPLACE FUNCTION get_file_extension(path TEXT) RETURNS TEXT AS $$ BEGIN
    RETURN substring(path FROM '\.([^\.]*)$');
END; $$ IMMUTABLE language plpgsql;

CREATE INDEX IF NOT EXISTS rockskip_symbols_gin ON rockskip_symbols USING GIN (
    -- repo_id
    singleton_integer(repo_id) gin__int_ops,

    -- added,deleted
    added gin__int_ops,
    deleted gin__int_ops,

    -- name
    name gin_trgm_ops,
    singleton(name),
    singleton(lower(name)),

    -- path
    path gin_trgm_ops,
    singleton(path),
    path_prefixes(path),
    singleton(lower(path)),
    path_prefixes(lower(path)),

    -- file extension
    singleton(get_file_extension(path)),
    singleton(get_file_extension(lower(path)))
);
