-- handles trigger_package_repo_filters_updated_at, constraints and unique index
DROP TABLE IF EXISTS package_repo_filters;

DROP FUNCTION IF EXISTS func_package_repo_filters_updated_at();

ALTER TABLE lsif_dependency_repos
    DROP COLUMN IF EXISTS blocked,
    DROP COLUMN IF EXISTS last_checked_at;

ALTER TABLE package_repo_versions
    DROP COLUMN IF EXISTS blocked,
    DROP COLUMN IF EXISTS last_checked_at;
