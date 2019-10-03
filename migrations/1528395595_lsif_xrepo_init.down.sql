-- Tear down inital LSIF database.

SELECT remote_exec('_lsif', '
    BEGIN;

    DROP INDEX IF EXISTS "packages_package_unique";
    DROP INDEX IF EXISTS "packages_repo_commit";
    DROP INDEX IF EXISTS "references_package";
    DROP INDEX IF EXISTS "references_repo_commit";
    DROP TABLE IF EXISTS "packages";
    DROP TABLE IF EXISTS "references";

    COMMIT;
');
