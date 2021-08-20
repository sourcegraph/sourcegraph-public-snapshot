BEGIN;

CREATE TABLE lsif_retention_configuration (
    id serial PRIMARY KEY,
    repository_id integer UNIQUE NOT NULL REFERENCES repo(id) ON DELETE CASCADE,
    max_age_for_non_stale_branches_seconds integer NOT NULL,
    max_age_for_non_stale_tags_seconds integer NOT NULL
);

COMMENT ON TABLE lsif_retention_configuration IS 'Stores the retention policy of code intellience data for a repository.';
COMMENT ON COLUMN lsif_retention_configuration.max_age_for_non_stale_branches_seconds IS 'The number of seconds since the last modification of a branch until it is considered stale.';
COMMENT ON COLUMN lsif_retention_configuration.max_age_for_non_stale_tags_seconds IS 'The nujmber of seconds since the commit date of a tagged commit until it is considered stale.';

COMMIT;
