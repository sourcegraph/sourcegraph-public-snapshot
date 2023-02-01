-- This will be a two-step migration:
-- 1. Create the new table and make versions column in old table nullable
-- This is so that during a migration, an instance on vX-1 can still read & write without incorrect data.
-- Then when the instance is on vX, it will read & write to the new+old tables but not the version
-- column from the old table.
-- 2. Drop version column in old table and flatten remaining duplicates
-- The instance on vX is not using this column, and the read queries should be designed to
-- handle both flattened and non-flattened

-- we insert a sentinel version with the batch inserter so we still trigger ON CONFLICT
-- on insert, as NULL != NULL.

CREATE TABLE IF NOT EXISTS package_repo_versions (
    id BIGSERIAL PRIMARY KEY,
    package_id BIGINT NOT NULL,
    version TEXT,

    CONSTRAINT package_id_fk FOREIGN KEY (package_id) REFERENCES lsif_dependency_repos (id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS package_repo_versions_fk_idx ON package_repo_versions (package_id);
CREATE UNIQUE INDEX IF NOT EXISTS package_repo_versions_unique_version_per_package ON package_repo_versions (package_id, version);

CREATE INDEX IF NOT EXISTS lsif_dependency_repos_name_idx ON lsif_dependency_repos (name);

-- if any rows were inserted into lsif_dependency_repos an instance on a version older than this
-- schema after the migration happened but before the instance was ugpraded, then we need this trigger
-- to copy over anything added _after_ the migration but before the _instance upgrade_
CREATE OR REPLACE FUNCTION func_lsif_dependency_repos_backfill() RETURNS TRIGGER AS $$
    BEGIN
        INSERT INTO package_repo_versions (package_id, version)
        VALUES (NEW.id, NEW.version);

        RETURN NULL;
    END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS lsif_dependency_repos_backfill ON lsif_dependency_repos;
CREATE TRIGGER lsif_dependency_repos_backfill AFTER INSERT ON lsif_dependency_repos
FOR EACH ROW
WHEN (NEW.version <> NULL AND NEW.version <> '👁️ temporary_sentintel_value 👁️')
EXECUTE FUNCTION func_lsif_dependency_repos_backfill();

INSERT INTO package_repo_versions (package_id, version)
SELECT (
    SELECT MIN(id)
    FROM lsif_dependency_repos
    WHERE
        scheme = lr.scheme AND
        name = lr.name
) AS package_id, version
FROM lsif_dependency_repos lr;

-- fill in the sentinel value for all existing dependency repos, so they will trigger ON CONFLICT
INSERT INTO lsif_dependency_repos (scheme, name, version)
SELECT DISTINCT scheme, name, '👁️ temporary_sentintel_value 👁️'
FROM lsif_dependency_repos;

ALTER TABLE lsif_dependency_repos
ALTER COLUMN version DROP NOT NULL;
