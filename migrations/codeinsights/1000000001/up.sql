-- +++
-- parent: 1000000000
-- +++

BEGIN;

CREATE EXTENSION IF NOT EXISTS timescaledb;
CREATE EXTENSION IF NOT EXISTS pg_trgm;
CREATE EXTENSION IF NOT EXISTS citext;

-- Records repository names, both historical and present, using a unique repository _name_ ID
-- (unrelated to the repository ID.)
CREATE TABLE repo_names (
    -- The repository _name_ ID.
    id bigserial NOT NULL PRIMARY KEY,

    -- The name, trigram-indexed for fast e.g. regexp filtering.
    name citext NOT NULL,

    CONSTRAINT check_name_nonempty CHECK ((name OPERATOR(<>) ''::citext))
);

-- Enforce that names are unique.
CREATE UNIQUE INDEX repo_names_name_unique_idx ON repo_names(name);

-- Create trigram indexes for repository name filtering based on e.g. regexps.
CREATE INDEX repo_names_name_trgm ON repo_names USING gin (lower((name)::text) gin_trgm_ops);


-- Records arbitrary metadata about events. Stored in a separate table as it is often repeated
-- for multiple events.
CREATE TABLE metadata (
    -- The metadata ID.
    id bigserial NOT NULL PRIMARY KEY,

    -- Metadata about this event, this can be any arbitrary JSON metadata which will be returned
    -- when querying events, and can be filtered on and grouped using jsonb operators ?, ?&, ?|,
    -- and @>. This should be small data only, primary use case is small lists such as:
    --
    --  {"java_versions": [...]}
    --  {"languages":     [...]}
    --  {"pull_requests": [...]}
    --  {"annotations":   [...]}
    --
    metadata jsonb NOT NULL
);

-- Enforce that metadata is unique.
CREATE UNIQUE INDEX metadata_metadata_unique_idx ON metadata(metadata);

-- Index metadata to optimize WHERE clauses with jsonb ?, ?&, ?|, and @> operators.
CREATE INDEX metadata_metadata_gin ON metadata USING GIN (metadata);

-- Records events over time associated with a repository (or none, i.e. globally) where a single
-- numerical value is going arbitrarily up and down.
--
-- Repository association is based on both repository ID and name. The ID can be used to refer to
-- a specific repository, or lookup the current name of a repository after it has been e.g. renamed.
-- The name can be used to refer to the name of the repository at the time of the event's creation,
-- for example to trace the change in a gauge back to a repository being renamed.
CREATE TABLE series_points (
    -- A unique identifier for the series of data being recorded. This is not an ID from another
    -- table, but rather just a unique identifier.
    series_id integer,

    -- The timestamp of the recorded event.
    time TIMESTAMPTZ NOT NULL,

    -- The floating point value at the time of the event.
    value double precision NOT NULL,

    -- Associated metadata for this event, if any.
    metadata_id integer,

    -- The repository ID (from the main application DB) at the time the event was created. Note
    -- that the repository may no longer exist / be valid at query time, however.
    --
    -- null if the event was not for a single repository (i.e. a global gauge).
    repo_id integer,

    -- The most recently known name for the repository, updated periodically to account for e.g.
    -- repository renames. If the repository was deleted, this is still the most recently known
    -- name.
    --
    -- null if the event was not for a single repository (i.e. a global gauge).
    repo_name_id integer,

    -- The repository name as it was known at the time the event was created. It may have been renamed
    -- since.
    original_repo_name_id integer,

    -- Ensure if one repo association field is specified, all are.
    CONSTRAINT check_repo_fields_specifity CHECK (
        ((repo_id IS NULL) AND (repo_name_id IS NULL) AND (original_repo_name_id IS NULL))
        OR
        ((repo_id IS NOT NULL) AND (repo_name_id IS NOT NULL) AND (original_repo_name_id IS NOT NULL))
    ),

    FOREIGN KEY (metadata_id) REFERENCES metadata(id) ON DELETE CASCADE DEFERRABLE,
    FOREIGN KEY (repo_name_id) REFERENCES repo_names(id) ON DELETE CASCADE DEFERRABLE,
    FOREIGN KEY (original_repo_name_id) REFERENCES repo_names(id) ON DELETE CASCADE DEFERRABLE
);

-- Create hypertable, partitioning events by time.
-- See https://docs.timescale.com/latest/using-timescaledb/hypertables
SELECT create_hypertable('series_points', 'time');

-- Create btree indexes for repository filtering.
CREATE INDEX series_points_repo_id_btree ON series_points USING btree (repo_id);
CREATE INDEX series_points_repo_name_id_btree ON series_points USING btree (repo_name_id);
CREATE INDEX series_points_original_repo_name_id_btree ON series_points USING btree (original_repo_name_id);

COMMIT;
