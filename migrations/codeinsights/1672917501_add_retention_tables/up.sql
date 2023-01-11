CREATE TABLE IF NOT EXISTS archived_series_points (
    series_id text NOT NULL,
    "time" timestamp with time zone NOT NULL,
    value double precision NOT NULL,
    repo_id integer,
    repo_name_id integer,
    original_repo_name_id integer,
    capture text,
    CONSTRAINT check_repo_fields_specifity CHECK ((((repo_id IS NULL) AND (repo_name_id IS NULL) AND (original_repo_name_id IS NULL)) OR ((repo_id IS NOT NULL) AND (repo_name_id IS NOT NULL) AND (original_repo_name_id IS NOT NULL)))),
    CONSTRAINT insight_series_series_id_fkey FOREIGN KEY (series_id) REFERENCES insight_series (series_id) ON DELETE CASCADE 
); -- any new column added to series_points should be added here too. we add a foreign key constraint for deletion.

CREATE TABLE IF NOT EXISTS archived_insight_series_recording_times (
    insight_series_id integer not null,
    recording_time timestamp with time zone not null,
    snapshot boolean not null,
    UNIQUE (insight_series_id, recording_time),
    CONSTRAINT insight_series_id_fkey FOREIGN KEY (insight_series_id) REFERENCES insight_series (id) ON DELETE CASCADE
); -- this structure should be kept the same as insight_series_recording_times.
