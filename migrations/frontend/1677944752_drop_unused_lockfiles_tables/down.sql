CREATE TABLE IF NOT EXISTS codeintel_lockfile_references (
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

CREATE TABLE IF NOT EXISTS codeintel_lockfiles (
    id SERIAL PRIMARY KEY,
    repository_id integer NOT NULL,
    commit_bytea bytea NOT NULL,
    codeintel_lockfile_reference_ids integer[] NOT NULL,
    lockfile text,
    fidelity text DEFAULT 'flat'::text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);

CREATE INDEX IF NOT EXISTS codeintel_lockfile_references_last_check_at ON codeintel_lockfile_references USING btree (last_check_at);
CREATE INDEX IF NOT EXISTS codeintel_lockfile_references_repository_id_commit_bytea ON codeintel_lockfile_references USING btree (repository_id, commit_bytea) WHERE ((repository_id IS NOT NULL) AND (commit_bytea IS NOT NULL));
CREATE UNIQUE INDEX IF NOT EXISTS codeintel_lockfile_references_repository_name_revspec_package_r ON codeintel_lockfile_references USING btree (repository_name, revspec, package_scheme, package_name, package_version, resolution_lockfile, resolution_repository_id, resolution_commit_bytea);
CREATE INDEX IF NOT EXISTS codeintel_lockfiles_codeintel_lockfile_reference_ids ON codeintel_lockfiles USING gin (codeintel_lockfile_reference_ids gin__int_ops);
CREATE INDEX IF NOT EXISTS codeintel_lockfiles_references_depends_on ON codeintel_lockfile_references USING gin (depends_on gin__int_ops);
CREATE UNIQUE INDEX IF NOT EXISTS codeintel_lockfiles_repository_id_commit_bytea_lockfile ON codeintel_lockfiles USING btree (repository_id, commit_bytea, lockfile);

CREATE TABLE IF NOT EXISTS last_lockfile_scan (
    repository_id integer NOT NULL PRIMARY KEY,
    last_lockfile_scan_at timestamp with time zone NOT NULL
);

COMMENT ON TABLE last_lockfile_scan IS 'Tracks the last time repository was checked for lockfile indexing.';

COMMENT ON COLUMN last_lockfile_scan.last_lockfile_scan_at IS 'The last time this repository was considered for lockfile indexing.';
