CREATE INDEX IF NOT EXISTS package_repo_versions_fk_idx ON package_repo_versions USING btree (package_id);
