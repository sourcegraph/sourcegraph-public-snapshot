BEGIN;

CREATE EXTENSION IF NOT EXISTS timescaledb;

CREATE TABLE histogram_events (
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
SELECT create_hypertable('histogram_events', 'time');

COMMIT;
