BEGIN;

-- Drop trigger as we don't want it to fire
DROP TRIGGER IF EXISTS trig_soft_delete_orphan_repo_by_external_service_repo
    ON external_service_repos;

-- Create temporary table to store duplicates
CREATE TEMPORARY TABLE deduped_external_service_repos (
  external_service_id bigint,
  repo_id integer,
  clone_url text
);

-- Get all existing duplicates
INSERT INTO deduped_external_service_repos
SELECT repo_id, external_service_id, clone_url
FROM external_service_repos
GROUP BY (repo_id, external_service_id, clone_url)
HAVING count(*) > 1;

-- Delete duplicates
DELETE FROM external_service_repos
WHERE (repo_id, external_service_id, clone_url)
          IN (SELECT * FROM deduped_external_service_repos);

-- Add unique constraint
ALTER TABLE external_service_repos
ADD CONSTRAINT external_service_repos_repo_id_external_service_id_unique
UNIQUE (repo_id, external_service_id);

-- Add single rows back
INSERT INTO external_service_repos
SELECT repo_id, external_service_id, clone_url
FROM deduped_external_service_repos;

-- Drop temporary table
DROP TABLE deduped_external_service_repos;

-- Recreate trigger
CREATE TRIGGER trig_soft_delete_orphan_repo_by_external_service_repo
    AFTER DELETE ON external_service_repos
    FOR EACH ROW EXECUTE PROCEDURE soft_delete_orphan_repo_by_external_service_repos();

COMMIT;
