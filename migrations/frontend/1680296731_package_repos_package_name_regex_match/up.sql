CREATE INDEX IF NOT EXISTS lsif_dependency_repos_name_gin
ON lsif_dependency_repos
USING gin (name gin_trgm_ops)
