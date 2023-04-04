---------------------------------------------------------
-- PACKAGE REPO FILTERS
---------------------------------------------------------

CREATE TABLE IF NOT EXISTS package_repo_filters (
    id SERIAL PRIMARY KEY NOT NULL,
    behaviour TEXT NOT NULL,
    scheme TEXT NOT NULL,
    matcher JSONB NOT NULL,
    deleted_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT statement_timestamp()
);

CREATE OR REPLACE FUNCTION func_package_repo_filters_updated_at() RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = statement_timestamp();
    RETURN NEW;
END $$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trigger_package_repo_filters_updated_at ON package_repo_filters;
CREATE TRIGGER trigger_package_repo_filters_updated_at
BEFORE UPDATE ON package_repo_filters
FOR EACH ROW
WHEN (OLD.* IS DISTINCT FROM NEW.*)
EXECUTE PROCEDURE func_package_repo_filters_updated_at();

ALTER TABLE package_repo_filters
    DROP CONSTRAINT IF EXISTS package_repo_filters_valid_oneof_glob,
    ADD CONSTRAINT package_repo_filters_valid_oneof_glob CHECK (
        (
            (matcher ? 'VersionGlob' AND matcher->>'VersionGlob' <> '' AND matcher->>'PackageName' <> '' AND NOT(matcher ? 'PackageGlob')) OR
            (matcher ? 'PackageGlob' AND matcher->>'PackageGlob' <> '' AND NOT(matcher ? 'VersionGlob'))
        )
    );

-- because creating types is unnecessarily awkward with idempotency
ALTER TABLE package_repo_filters
    DROP CONSTRAINT IF EXISTS package_repo_filters_is_pkgrepo_scheme,
    ADD CONSTRAINT package_repo_filters_is_pkgrepo_scheme CHECK (
        scheme = ANY('{"semanticdb","npm","go","python","rust-analyzer","scip-ruby"}')
    );

ALTER TABLE package_repo_filters
    DROP CONSTRAINT IF EXISTS package_repo_filters_behaviour_is_allow_or_block,
    ADD CONSTRAINT package_repo_filters_behaviour_is_allow_or_block CHECK (
        behaviour = ANY('{"BLOCK","ALLOW"}')
    );

CREATE UNIQUE INDEX IF NOT EXISTS package_repo_filters_unique_matcher_per_scheme
ON package_repo_filters (scheme, matcher);

---------------------------------------------------------
-- PACKAGE REPOS & VERSIONS BLOCK AND LAST CHECK DATES
---------------------------------------------------------

ALTER TABLE lsif_dependency_repos
ADD COLUMN IF NOT EXISTS blocked BOOLEAN NOT NULL DEFAULT FALSE,
ADD COLUMN IF NOT EXISTS last_checked_at TIMESTAMPTZ;

CREATE INDEX IF NOT EXISTS lsif_dependency_repos_blocked
ON lsif_dependency_repos USING btree (blocked);

CREATE INDEX IF NOT EXISTS lsif_dependency_repos_last_checked_at
ON lsif_dependency_repos USING btree (last_checked_at);

ALTER TABLE package_repo_versions
ADD COLUMN IF NOT EXISTS blocked BOOLEAN NOT NULL DEFAULT FALSE,
ADD COLUMN IF NOT EXISTS last_checked_at TIMESTAMPTZ;

CREATE INDEX IF NOT EXISTS package_repo_versions_blocked
ON package_repo_versions USING btree (blocked);

CREATE INDEX IF NOT EXISTS package_repo_versions_last_checked_at
ON package_repo_versions USING btree (last_checked_at);
