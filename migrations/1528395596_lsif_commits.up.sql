SELECT remote_exec('_lsif', '
    CREATE TABLE IF NOT EXISTS "commits" (
        "id" SERIAL PRIMARY KEY,
        "repository" text NOT NULL,
        "commit" text NOT NULL,
        "parentCommit" text NOT NULL
    );

    CREATE TABLE IF NOT EXISTS "lsifDataMarkers" (
        "repository" text NOT NULL,
        "commit" text NOT NULL,
        PRIMARY KEY ("repository", "commit")
    );

    CREATE UNIQUE INDEX IF NOT EXISTS "commits_repo_commit_parentCommit_unique" ON "commits"("repository", "commit", "parentCommit");
    CREATE INDEX IF NOT EXISTS "commits_repo_commit" ON "commits"("repository", "commit");
    CREATE INDEX IF NOT EXISTS "commits_repo_parentCommit" ON "commits"("repository", "commit");

    CREATE OR REPLACE VIEW "commitWithLsifMarkers" AS
        SELECT
            c."repository",
            c."commit",
            c."parentCommit",
            EXISTS (
                SELECT 1
                FROM "lsifDataMarkers" m
                WHERE m."repository" = c."repository"
                AND m."commit" = c."commit"
            ) AS hasLsifData
        FROM "commits" c;
');
