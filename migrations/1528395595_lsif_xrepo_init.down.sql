-- Performs migration in LSIF database

SELECT dblink_exec('dbname=sourcegraph_lsif user=' || current_user, '
    BEGIN;

    DROP INDEX IF EXISTS "packages_package_unique";
    DROP INDEX IF EXISTS "packages_repo_commit";
    DROP INDEX IF EXISTS "references_package";
    DROP INDEX IF EXISTS "references_repo_commit";
    DROP TABLE IF EXISTS "packages";
    DROP TABLE IF EXISTS "references";

    COMMIT;
');
