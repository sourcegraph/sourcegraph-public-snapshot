DROP INDEX IF EXISTS repo_kvps_trgm_idx ON repo_kvps
USING gin (key gin_trgm_ops, value gin_trgm_ops);
