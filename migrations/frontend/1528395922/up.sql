CREATE INDEX IF NOT EXISTS external_service_repos_org_id_idx ON external_service_repos USING btree (org_id) WHERE org_id IS NOT NULL;
