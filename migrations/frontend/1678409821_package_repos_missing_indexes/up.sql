CREATE UNIQUE INDEX IF NOT EXISTS lsif_dependency_repos_unique_scheme_name
ON lsif_dependency_repos USING btree (scheme, name);

CREATE INDEX IF NOT EXISTS lsif_dependency_repos_scheme_id
ON lsif_dependency_repos USING btree (scheme, id);

DROP INDEX IF EXISTS lsif_dependency_repos_name_idx;
CREATE INDEX IF NOT EXISTS lsif_dependency_repos_name_id
ON lsif_dependency_repos USING btree (name, id);
