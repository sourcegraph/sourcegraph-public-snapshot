BEGIN;

CREATE TABLE IF NOT EXISTS lsif_dependency_repos (
    id bigserial NOT NULL PRIMARY KEY,
    name text NOT NULL,
    version text NOT NULL,
    scheme text NOT NULL,
    CONSTRAINT lsif_dependency_repos_unique_triplet
        UNIQUE (scheme, name, version)
);

COMMIT;
