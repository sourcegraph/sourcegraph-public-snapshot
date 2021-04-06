BEGIN;

ALTER TABLE external_service_repos ADD COLUMN user_id int REFERENCES users(id) ON DELETE CASCADE DEFERRABLE;

UPDATE external_service_repos
SET user_id = es.namespace_user_id
FROM external_services es
WHERE es.id = external_service_id AND es.namespace_user_id IS NOT NULL;

CREATE INDEX external_service_user_repos_idx ON external_service_repos(user_id, repo_id) WHERE user_id IS NOT NULL;

COMMIT;
