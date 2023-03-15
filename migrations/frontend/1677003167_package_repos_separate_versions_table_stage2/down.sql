ALTER TABLE lsif_dependency_repos ADD COLUMN IF NOT EXISTS version TEXT DEFAULT '👁️temporary_sentinel_value👁️';

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
WHEN (NEW.version <> '👁️temporary_sentinel_value👁️')
EXECUTE FUNCTION func_lsif_dependency_repos_backfill();
