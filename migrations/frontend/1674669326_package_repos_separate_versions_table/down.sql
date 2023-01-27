
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
        SELECT * FROM matched_triplets;

        DELETE FROM lsif_dependency_repos
        WHERE version = '👁️ temporary_sentintel_value 👁️';
    END IF;
END
$$;

ALTER TABLE lsif_dependency_repos
ALTER COLUMN version SET NOT NULL;

DROP INDEX IF EXISTS package_repo_versions_fk_idx;

DROP TABLE IF EXISTS package_repo_versions;
