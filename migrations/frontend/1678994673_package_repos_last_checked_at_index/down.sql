DROP INDEX IF EXISTS lsif_dependency_repos_last_checked_at;
DROP INDEX IF EXISTS package_repo_versions_last_checked_at;

CREATE INDEX IF NOT EXISTS lsif_dependency_repos_last_checked_at
ON lsif_dependency_repos
USING btree (last_checked_at);

CREATE INDEX IF NOT EXISTS package_repo_versions_last_checked_at
ON package_repo_versions
USING btree (last_checked_at);
