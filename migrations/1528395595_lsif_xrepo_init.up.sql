-- Performs migration in LSIF database

SELECT dblink_exec('dbname=sourcegraph_lsif user=' || current_user, '
    -- This gets weirdly meta. We need to enable the extension
    -- on this side as well so the LSIF processes can query into
    -- the sourcegraph db to check if migrations have been
    -- applied on application startup.
    CREATE EXTENSION IF NOT EXISTS dblink;

    BEGIN;

    CREATE TABLE "packages" (
        "id" SERIAL PRIMARY KEY,
        "scheme" text NOT NULL,
        "name" text NOT NULL,
        "version" text NOT NULL,
        "repository" text,
        "commit" text NOT NULL
    );

    CREATE TABLE "references" (
        "id" SERIAL PRIMARY KEY,
        "scheme" text NOT NULL,
        "name" text NOT NULL,
        "version" text NOT NULL,
        "repository" text,
        "commit" text NOT NULL,
        "filter" bytea NOT NULL
    );

    CREATE UNIQUE INDEX "packages_package_unique" ON "packages"("scheme", "name", "version");
    CREATE INDEX "packages_repo_commit" ON "packages"("repository", "commit");
    CREATE INDEX "references_package" ON "references"("scheme", "name", "version");
    CREATE INDEX "references_repo_commit" ON "references"("repository", "commit");

    COMMIT;
');
