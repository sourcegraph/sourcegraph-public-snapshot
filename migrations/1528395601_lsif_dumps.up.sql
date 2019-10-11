-- Note: `commit` is a reserved word, so it's quoted.

BEGIN;

DROP VIEW IF EXISTS lsif_commits_with_lsif_data_markers;

CREATE TABLE IF NOT EXISTS lsif_dumps (
    id SERIAL PRIMARY KEY,
    repository TEXT NOT NULL CHECK (repository != ''),
    "commit" TEXT NOT NULL CHECK (LENGTH("commit") = 40),
    UNIQUE (repository, "commit")
);
INSERT INTO lsif_dumps (repository, "commit") SELECT DISTINCT * FROM lsif_data_markers ON CONFLICT DO NOTHING;
DROP TABLE IF EXISTS lsif_data_markers;

-- add column dump_id
ALTER TABLE lsif_packages ADD COLUMN dump_id INTEGER REFERENCES lsif_dumps(id) ON DELETE CASCADE;
ALTER TABLE lsif_references ADD COLUMN dump_id INTEGER REFERENCES lsif_dumps(id) ON DELETE CASCADE;

-- delete invalid rows
DELETE FROM lsif_packages pkg WHERE NOT EXISTS (SELECT * FROM lsif_dumps dump_ WHERE dump_.repository = pkg.repository AND dump_."commit" = pkg."commit");
DELETE FROM lsif_references ref WHERE NOT EXISTS (SELECT * FROM lsif_dumps dump_ WHERE dump_.repository = ref.repository AND dump_."commit" = ref."commit");

-- populate dump_id
UPDATE lsif_packages pkg SET dump_id = dump_.id FROM lsif_dumps dump_ WHERE dump_.repository = pkg.repository AND dump_."commit" = pkg."commit";
UPDATE lsif_references ref SET dump_id = dump_.id FROM lsif_dumps dump_ WHERE dump_.repository = ref.repository AND dump_."commit" = ref."commit";

-- set NOT NULL
ALTER TABLE lsif_packages ALTER COLUMN dump_id SET NOT NULL;
ALTER TABLE lsif_references ALTER COLUMN dump_id SET NOT NULL;

-- drop old columns
ALTER TABLE lsif_packages DROP COLUMN repository;
ALTER TABLE lsif_packages DROP COLUMN "commit";
ALTER TABLE lsif_references DROP COLUMN repository;
ALTER TABLE lsif_references DROP COLUMN "commit";

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
