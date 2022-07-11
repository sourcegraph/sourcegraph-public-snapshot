CREATE TABLE IF NOT EXISTS codeintel_lockfiles (
    id SERIAL PRIMARY KEY,
    repository_id integer NOT NULL,
    commit_bytea bytea NOT NULL,
    codeintel_lockfile_reference_ids integer [] NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS codeintel_lockfiles_repository_id_commit_bytea ON codeintel_lockfiles USING btree (repository_id, commit_bytea);

CREATE INDEX IF NOT EXISTS codeintel_lockfiles_codeintel_lockfile_reference_ids ON codeintel_lockfiles USING GIN (
    codeintel_lockfile_reference_ids gin__int_ops
);

COMMENT ON TABLE codeintel_lockfiles IS 'Associates a repository-commit pair with the set of repository-level dependencies parsed from lockfiles.';

COMMENT ON COLUMN codeintel_lockfiles.commit_bytea IS 'A 40-char revhash. Note that this commit may not be resolvable in the future.';

COMMENT ON COLUMN codeintel_lockfiles.codeintel_lockfile_reference_ids IS 'A key to a resolved repository name-revspec pair. Not all repository names and revspecs are resolvable.';

CREATE TABLE IF NOT EXISTS codeintel_lockfile_references (
    id SERIAL PRIMARY KEY,
    repository_name text NOT NULL,
    revspec text NOT NULL,
    package_scheme text NOT NULL,
    package_name text NOT NULL,
    package_version text NOT NULL,
    repository_id integer,
    commit_bytea bytea
);

CREATE UNIQUE INDEX IF NOT EXISTS codeintel_lockfile_references_repository_name_revspec_package ON codeintel_lockfile_references USING btree (
    repository_name,
    revspec,
    package_scheme,
    package_name,
    package_version
);

CREATE UNIQUE INDEX IF NOT EXISTS codeintel_lockfile_references_repository_id_commit_bytea ON codeintel_lockfile_references USING btree (repository_id, commit_bytea)
WHERE
    repository_id IS NOT NULL
    AND commit_bytea IS NOT NULL;

COMMENT ON TABLE codeintel_lockfile_references IS 'Tracks a lockfile dependency that might be resolvable to a specific repository-commit pair.';

COMMENT ON COLUMN codeintel_lockfile_references.repository_name IS 'Encodes `reposource.PackageDependency.RepoName`. A name that is "globally unique" for a Sourcegraph instance. Used in `repo:...` queries.';

COMMENT ON COLUMN codeintel_lockfile_references.revspec IS 'Encodes `reposource.PackageDependency.GitTagFromVersion`. Returns the git tag associated with the given dependency version, used in `rev:` or `repo:foo@rev` queries.';

COMMENT ON COLUMN codeintel_lockfile_references.package_scheme IS 'Encodes `reposource.PackageDependency.Scheme`. The scheme of the dependency (e.g., semanticdb, npm).';

COMMENT ON COLUMN codeintel_lockfile_references.package_name IS 'Encodes `reposource.PackageDependency.PackageSyntax`. The name of the dependency as used by the package manager, excluding version information.';

COMMENT ON COLUMN codeintel_lockfile_references.package_version IS 'Encodes `reposource.PackageDependency.PackageVersion`. The version of the package.';

COMMENT ON COLUMN codeintel_lockfile_references.repository_id IS 'The identifier of the repo that resolves the associated name, if it is resolvable on this instance.';

COMMENT ON COLUMN codeintel_lockfile_references.commit_bytea IS 'The resolved 40-char revhash of the associated revspec, if it is resolvable on this instance.';
