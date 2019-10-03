-- Set up initial LSIF database.
-- Note: `commit` is a reserved word, so it's quoted.
-- Note: `references` is a reserved word, so it's quoted.

SELECT remote_exec('_lsif', '
    BEGIN;

    CREATE TABLE IF NOT EXISTS packages (
        id SERIAL PRIMARY KEY,
        scheme text NOT NULL,
        name text NOT NULL,
        version text,
        repository text NOT NULL,
        "commit" text NOT NULL
    );

    CREATE TABLE IF NOT EXISTS "references" (
        id SERIAL PRIMARY KEY,
        scheme text NOT NULL,
        name text NOT NULL,
        version text,
        repository text NOT NULL,
        "commit" text NOT NULL,
        filter bytea NOT NULL
    );

    CREATE UNIQUE INDEX IF NOT EXISTS packages_package_unique ON packages(scheme, name, version);
    CREATE INDEX IF NOT EXISTS packages_repo_commit ON packages(repository, "commit");
    CREATE INDEX IF NOT EXISTS references_package ON "references"(scheme, name, version);
    CREATE INDEX IF NOT EXISTS references_repo_commit ON "references"(repository, "commit");

    COMMIT;
');
