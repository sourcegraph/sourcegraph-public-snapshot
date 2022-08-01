ALTER TABLE lsif_configuration_policies
  ADD COLUMN IF NOT EXISTS lockfile_indexing_enabled boolean NOT NULL DEFAULT false;

COMMENT ON COLUMN lsif_configuration_policies.lockfile_indexing_enabled IS 'Whether to index the lockfiles in the repositories matched by this policy';

CREATE TABLE codeintel_lockfile_references (
    id SERIAL PRIMARY KEY,
    repository_name text NOT NULL,
    revspec text NOT NULL,
    package_scheme text NOT NULL,
    package_name text NOT NULL,
    package_version text NOT NULL,
    repository_id integer,
    commit_bytea bytea,
    last_check_at timestamp with time zone,
    depends_on integer[] DEFAULT '{}'::integer[],
    resolution_lockfile text,
    resolution_repository_id integer,
    resolution_commit_bytea bytea
);

COMMENT ON TABLE codeintel_lockfile_references IS 'Tracks a lockfile dependency that might be resolvable to a specific repository-commit pair.';
COMMENT ON COLUMN codeintel_lockfile_references.repository_name IS 'Encodes `reposource.PackageDependency.RepoName`. A name that is "globally unique" for a Sourcegraph instance. Used in `repo:...` queries.';
COMMENT ON COLUMN codeintel_lockfile_references.revspec IS 'Encodes `reposource.PackageDependency.GitTagFromVersion`. Returns the git tag associated with the given dependency version, used in `rev:` or `repo:foo@rev` queries.';
COMMENT ON COLUMN codeintel_lockfile_references.package_scheme IS 'Encodes `reposource.PackageDependency.Scheme`. The scheme of the dependency (e.g., semanticdb, npm).';
COMMENT ON COLUMN codeintel_lockfile_references.package_name IS 'Encodes `reposource.PackageDependency.PackageSyntax`. The name of the dependency as used by the package manager, excluding version information.';
COMMENT ON COLUMN codeintel_lockfile_references.package_version IS 'Encodes `reposource.PackageDependency.PackageVersion`. The version of the package.';
COMMENT ON COLUMN codeintel_lockfile_references.repository_id IS 'The identifier of the repo that resolves the associated name, if it is resolvable on this instance.';
COMMENT ON COLUMN codeintel_lockfile_references.commit_bytea IS 'The resolved 40-char revhash of the associated revspec, if it is resolvable on this instance.';
COMMENT ON COLUMN codeintel_lockfile_references.last_check_at IS 'Timestamp when background job last checked this row for repository resolution';
COMMENT ON COLUMN codeintel_lockfile_references.depends_on IS 'IDs of other `codeintel_lockfile_references` this package depends on in the context of this `codeintel_lockfile_references.resolution_id`.';
COMMENT ON COLUMN codeintel_lockfile_references.resolution_lockfile IS 'Relative path of lockfile in which this package was referenced. Corresponds to `codeintel_lockfiles.lockfile`.';
COMMENT ON COLUMN codeintel_lockfile_references.resolution_repository_id IS 'ID of the repository in which lockfile was resolved. Corresponds to `codeintel_lockfiles.repository_id`.';
COMMENT ON COLUMN codeintel_lockfile_references.resolution_commit_bytea IS 'Commit at which lockfile was resolved. Corresponds to `codeintel_lockfiles.commit_bytea`.';

CREATE TABLE codeintel_lockfiles (
    id SERIAL PRIMARY KEY,
    repository_id integer NOT NULL,
    commit_bytea bytea NOT NULL,
    codeintel_lockfile_reference_ids integer[] NOT NULL,
    lockfile text,
    fidelity text DEFAULT 'flat'::text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);

COMMENT ON TABLE codeintel_lockfiles IS 'Associates a repository-commit pair with the set of repository-level dependencies parsed from lockfiles.';
COMMENT ON COLUMN codeintel_lockfiles.commit_bytea IS 'A 40-char revhash. Note that this commit may not be resolvable in the future.';
COMMENT ON COLUMN codeintel_lockfiles.codeintel_lockfile_reference_ids IS 'A key to a resolved repository name-revspec pair. Not all repository names and revspecs are resolvable.';
COMMENT ON COLUMN codeintel_lockfiles.lockfile IS 'Relative path of a lockfile in the given repository and the given commit.';
COMMENT ON COLUMN codeintel_lockfiles.fidelity IS 'Fidelity of the dependency graph thats persisted, whether it is a flat list, a whole graph, circular graph, ...';
COMMENT ON COLUMN codeintel_lockfiles.created_at IS 'Time when lockfile was indexed';
COMMENT ON COLUMN codeintel_lockfiles.updated_at IS 'Time when lockfile index was updated';
