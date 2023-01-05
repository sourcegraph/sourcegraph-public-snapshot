CREATE TABLE IF NOT EXISTS archived_series_points (
    series_id text NOT NULL,
    "time" timestamp with time zone NOT NULL,
    value double precision NOT NULL,
    metadata_id integer,
    repo_id integer,
    repo_name_id integer,
    original_repo_name_id integer,
    capture text,
    CONSTRAINT check_repo_fields_specifity CHECK ((((repo_id IS NULL) AND (repo_name_id IS NULL) AND (original_repo_name_id IS NULL)) OR ((repo_id IS NOT NULL) AND (repo_name_id IS NOT NULL) AND (original_repo_name_id IS NOT NULL))))
); -- this structure should be kept the same as series_points.

CREATE TABLE IF NOT EXISTS archived_insight_series_recording_times (
    insight_series_id integer,
    recording_time timestamp with time zone,
    snapshot boolean,
    UNIQUE (insight_series_id, recording_time),
    CONSTRAINT insight_series_id_fkey FOREIGN KEY (insight_series_id) REFERENCES insight_series (id) ON DELETE CASCADE
); -- this structure should be kept the same as insight_series_recording_times.
