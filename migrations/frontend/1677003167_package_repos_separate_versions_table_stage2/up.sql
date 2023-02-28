DROP TRIGGER IF EXISTS lsif_dependency_repos_backfill ON lsif_dependency_repos;
DROP FUNCTION IF EXISTS func_lsif_dependency_repos_backfill;

ALTER TABLE lsif_dependency_repos
DROP COLUMN IF EXISTS version;

DELETE FROM lsif_dependency_repos
WHERE id IN (
    SELECT lr.id
    FROM lsif_dependency_repos lr
    LEFT JOIN package_repo_versions prv
    ON lr.id = prv.package_id
    WHERE prv.package_id IS NULL
);
