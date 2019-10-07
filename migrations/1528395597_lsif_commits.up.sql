-- Note: `commit` is a reserved word, so it's quoted.

SELECT remote_exec('_lsif', '
    CREATE TABLE IF NOT EXISTS commits (
        id SERIAL PRIMARY KEY,
        repository text NOT NULL,
        "commit" text NOT NULL,
        parent_commit text NOT NULL
    );

    CREATE TABLE IF NOT EXISTS lsif_data_markers (
        repository text NOT NULL,
        "commit" text NOT NULL,
        PRIMARY KEY (repository, "commit")
    );

    CREATE UNIQUE INDEX IF NOT EXISTS commits_repo_commit_parent_commit_unique ON commits(repository, "commit", parent_commit);
    CREATE INDEX IF NOT EXISTS commits_repo_commit ON commits(repository, "commit");
    CREATE INDEX IF NOT EXISTS commits_repo_parent_commit ON commits(repository, "commit");

    CREATE OR REPLACE VIEW commits_with_lsif_data_markers AS
        SELECT
            c.repository,
            c."commit",
            c.parent_commit,
            EXISTS (
                SELECT 1
                FROM lsif_data_markers m
                WHERE m.repository = c.repository
                AND m."commit" = c."commit"
            ) AS has_lsif_data
        FROM commits c;
');
