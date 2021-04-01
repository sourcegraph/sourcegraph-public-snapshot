BEGIN;

DROP INDEX external_service_user_repos_idx;
ALTER TABLE external_service_repos DROP COLUMN user_id;

COMMIT;
