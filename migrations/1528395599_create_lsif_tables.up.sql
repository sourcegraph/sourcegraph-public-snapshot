-- Note: `commit` is a reserved word, so it's quoted.

BEGIN;

CREATE TABLE IF NOT EXISTS lsif_packages (
    id SERIAL PRIMARY KEY,
    scheme text NOT NULL,
    name text NOT NULL,
    version text,
    repository text NOT NULL,
    "commit" text NOT NULL
);

CREATE TABLE IF NOT EXISTS lsif_references (
    id SERIAL PRIMARY KEY,
    scheme text NOT NULL,
    name text NOT NULL,
    version text,
    repository text NOT NULL,
    "commit" text NOT NULL,
    filter bytea NOT NULL
);

CREATE TABLE IF NOT EXISTS lsif_commits (
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

CREATE UNIQUE INDEX IF NOT EXISTS lsif_packages_package_unique ON lsif_packages(scheme, name, version);
CREATE INDEX IF NOT EXISTS lsif_packages_repo_commit ON lsif_packages(repository, "commit");
CREATE INDEX IF NOT EXISTS lsif_references_package ON lsif_references(scheme, name, version);
CREATE INDEX IF NOT EXISTS lsif_references_repo_commit ON lsif_references(repository, "commit");
CREATE UNIQUE INDEX IF NOT EXISTS lsif_commits_repo_commit_parent_commit_unique ON lsif_commits(repository, "commit", parent_commit);
CREATE INDEX IF NOT EXISTS lsif_commits_repo_commit ON lsif_commits(repository, "commit");
CREATE INDEX IF NOT EXISTS lsif_commits_repo_parent_commit ON lsif_commits(repository, parent_commit);

CREATE OR REPLACE VIEW lsif_commits_with_lsif_data_markers AS
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
    FROM lsif_commits c;

COMMIT;
