DROP INDEX IF EXISTS lsif_dependency_repos_unique_scheme_name;
DROP INDEX IF EXISTS lsif_dependency_repos_scheme_id;
DROP INDEX IF EXISTS lsif_dependency_repos_name_id;
CREATE INDEX IF NOT EXISTS lsif_dependency_repos_name_idx ON lsif_dependency_repos USING btree (name);
