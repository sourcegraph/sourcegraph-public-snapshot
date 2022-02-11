BEGIN;

CREATE INDEX external_service_repos_external_service_id ON external_service_repos USING btree (external_service_id);

COMMIT;
