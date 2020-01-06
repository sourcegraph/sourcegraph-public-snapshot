BEGIN;

-- don't need to undo the changes to default_repos. Just need to undo our
-- index changes.

CREATE UNIQUE INDEX repo_external_service_unique_idx ON repo USING btree (external_service_type, external_service_id, external_id) WHERE ((external_service_type IS NOT NULL) AND (external_service_id IS NOT NULL) AND (external_id IS NOT NULL));

DROP INDEX IF EXISTS repo_external_unique_idx;

COMMIT;
