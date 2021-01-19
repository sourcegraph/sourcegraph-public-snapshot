BEGIN;

CREATE EXTENSION IF NOT EXISTS timescaledb;

-- Records events over time associated with a repository (or none, i.e. globally) where a single
-- numerical value is going arbitrarily up and down.
--
-- Repository association is based on both repository ID and name. The ID can be used to refer to
-- a specific repository, or lookup the current name of a repository after it has been e.g. renamed.
-- The name can be used to refer to the name of the repository at the time of the event's creation,
-- for example to trace the change in a gauge back to a repository being renamed.
CREATE TABLE gauge_events (
    -- The timestamp of the recorded event.
    time TIMESTAMPTZ NOT NULL,

    -- The floating point value at the time of the event.
    value double precision NOT NULL,

    -- Metadata about this event, this can be any arbitrary JSON metadata which will be returned
    -- when querying events, but cannot be filtered on.
    metadata jsonb NOT NULL,

    -- If the event was for a single repository (usually the case) then this field should indicate
    -- the repository ID at the time the event was created. Note that the repository may no longer
    -- exist / be valid at query time, however.
    repo_id integer,

    -- If the event was for a single repository (usually the case) then this field should indicate
    -- the repository name at the time the event was created. Note that the repository name may
    -- have changed since the event was created (e.g. if the repo was renamed), in which case this
    -- describes the outdated repository na,e.
    repo_name citext
);

-- Create hypertable, partitioning histogram events by time.
-- See https://docs.timescale.com/latest/using-timescaledb/hypertables
SELECT create_hypertable('gauge_events', 'time');

COMMIT;
