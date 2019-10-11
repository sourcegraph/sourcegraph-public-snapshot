-- Note: `commit` is a reserved word, so it's quoted.

BEGIN;

DROP VIEW IF EXISTS lsif_commits_with_lsif_data;

-- add columns repository and commit
ALTER TABLE lsif_packages ADD COLUMN repository TEXT;
ALTER TABLE lsif_packages ADD COLUMN "commit" TEXT;
ALTER TABLE lsif_references ADD COLUMN repository TEXT;
ALTER TABLE lsif_references ADD COLUMN "commit" TEXT;

-- populate repository and commit
UPDATE lsif_packages pkg SET repository = dump_.repository, "commit" = dump_."commit" FROM lsif_dumps dump_ WHERE dump_.id = pkg.dump_id;
UPDATE lsif_references ref SET repository = dump_.repository, "commit" = dump_."commit" FROM lsif_dumps dump_ WHERE dump_.id = ref.dump_id;

-- set NOT NULL
ALTER TABLE lsif_packages ALTER COLUMN repository SET NOT NULL;
ALTER TABLE lsif_packages ALTER COLUMN "commit" SET NOT NULL;
ALTER TABLE lsif_references ALTER COLUMN repository SET NOT NULL;
ALTER TABLE lsif_references ALTER COLUMN "commit" SET NOT NULL;

-- drop old columns
ALTER TABLE lsif_packages DROP COLUMN dump_id;
ALTER TABLE lsif_references DROP COLUMN dump_id;

CREATE TABLE IF NOT EXISTS lsif_data_markers (
    repository text NOT NULL,
    "commit" text NOT NULL,
    PRIMARY KEY (repository, "commit")
);
INSERT INTO lsif_data_markers SELECT repository, "commit" FROM lsif_dumps ON CONFLICT DO NOTHING;
DROP TABLE IF EXISTS lsif_dumps;

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

END;
