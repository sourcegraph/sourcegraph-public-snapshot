BEGIN;

LOCK TABLE external_service_repos;

-- Drop trigger as we don't want it to fire
DROP TRIGGER IF EXISTS trig_soft_delete_orphan_repo_by_external_service_repo
ON external_service_repos;

WITH dups AS (SELECT external_service_id, repo_id, min(ctid)
              FROM external_service_repos
              GROUP BY external_service_id, repo_id
              HAVING count(*) > 1
)
DELETE FROM external_service_repos
USING dups
WHERE (external_service_repos.external_service_id, external_service_repos.repo_id) = (dups.external_service_id, dups.repo_id)
AND external_service_repos.ctid <> dups.min;

-- Add unique constraint
ALTER TABLE external_service_repos
ADD CONSTRAINT external_service_repos_repo_id_external_service_id_unique
UNIQUE (repo_id, external_service_id);

-- Recreate trigger
CREATE TRIGGER trig_soft_delete_orphan_repo_by_external_service_repo
AFTER DELETE ON external_service_repos
FOR EACH ROW EXECUTE PROCEDURE soft_delete_orphan_repo_by_external_service_repos();

COMMIT;
