DROP INDEX IF EXISTS external_service_user_repos_idx;

ALTER TABLE users DROP COLUMN IF EXISTS tags;

DROP TABLE IF EXISTS user_public_repos;

-- TODO: Only run when these columns still exist.
-- TODO: Make sure this is a clean deletion.
DELETE FROM external_services WHERE namespace_user_id IS NOT NULL OR namespace_org_id IS NOT NULL;
ALTER TABLE external_services DROP COLUMN IF EXISTS namespace_user_id;
ALTER TABLE external_services DROP COLUMN IF EXISTS namespace_org_id;

DELETE FROM external_service_repos WHERE user_id IS NOT NULL OR org_id IS NOT NULL;
ALTER TABLE external_service_repos DROP COLUMN IF EXISTS user_id;
ALTER TABLE external_service_repos DROP COLUMN IF EXISTS org_id;
