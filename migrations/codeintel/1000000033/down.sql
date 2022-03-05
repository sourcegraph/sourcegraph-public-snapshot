DROP INDEX IF EXISTS rockskip_symbols_gin;

ALTER TABLE rockskip_symbols DROP COLUMN file_extension;

DROP FUNCTION IF EXISTS get_file_extension;

CREATE INDEX IF NOT EXISTS rockskip_symbols_gin ON rockskip_symbols USING GIN (
    singleton_integer(repo_id) gin__int_ops,
    added gin__int_ops,
    deleted gin__int_ops,
    singleton(path),
    path_prefixes(path),
    singleton(name),
    name gin_trgm_ops
);
