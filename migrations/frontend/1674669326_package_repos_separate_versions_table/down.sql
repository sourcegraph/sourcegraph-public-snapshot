DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.tables
        WHERE
            table_name = 'package_repo_versions' AND
            table_schema = current_schema()
    ) THEN
        WITH matched_triplets AS (
            SELECT lr.scheme, lr.name, rpv.version
            FROM package_repo_versions rpv
            JOIN lsif_dependency_repos lr
            ON rpv.package_id = lr.id
        )
        INSERT INTO lsif_dependency_repos (scheme, name, version)
        SELECT * FROM matched_triplets
        ON CONFLICT DO NOTHING;
    END IF;
END
$$;

DELETE FROM lsif_dependency_repos
WHERE version = '👁️temporary_sentinel_value👁️';

DROP INDEX IF EXISTS package_repo_versions_fk_idx;
DROP INDEX IF EXISTS package_repo_versions_unique_version_per_package;

DROP INDEX IF EXISTS lsif_dependency_repos_name_idx;

DROP TABLE IF EXISTS package_repo_versions;

DROP TRIGGER IF EXISTS lsif_dependency_repos_backfill ON lsif_dependency_repos;
DROP FUNCTION IF EXISTS func_lsif_dependency_repos_backfill;
