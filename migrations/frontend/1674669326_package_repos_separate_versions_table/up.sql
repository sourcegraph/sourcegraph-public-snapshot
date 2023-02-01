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

CREATE INDEX IF NOT EXISTS package_repo_versions_fk_idx
ON package_repo_versions (package_id);

CREATE UNIQUE INDEX IF NOT EXISTS package_repo_versions_unique_version_per_package
on package_repo_versions (package_id, version);

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
