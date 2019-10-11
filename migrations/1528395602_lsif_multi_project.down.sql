BEGIN;

ALTER TABLE lsif_dumps DROP COLUMN root;
ALTER TABLE lsif_dumps ADD UNIQUE (repository, commit);

CREATE OR REPLACE VIEW lsif_commits_with_lsif_data AS
    SELECT
        c.repository,
        c."commit",
        c.parent_commit,
        EXISTS (
            SELECT 1
            FROM lsif_dumps dump
            WHERE dump.repository = c.repository
            AND dump."commit" = c."commit"
        ) AS has_lsif_data
    FROM lsif_commits c;

END;
