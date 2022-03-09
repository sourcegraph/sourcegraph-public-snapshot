DROP INDEX IF EXISTS rockskip_symbols_gin;

CREATE OR REPLACE FUNCTION get_file_extension(path TEXT) RETURNS TEXT AS $$ BEGIN
    RETURN substring(path FROM '\.([^\.]*)$');
END; $$ IMMUTABLE language plpgsql;

CREATE INDEX IF NOT EXISTS rockskip_symbols_gin ON rockskip_symbols USING GIN (
    singleton_integer(repo_id) gin__int_ops,
    added gin__int_ops,
    deleted gin__int_ops,
    singleton(path),
    path_prefixes(path),
    singleton(name),
    name gin_trgm_ops,
    singleton(get_file_extension(path))
);
