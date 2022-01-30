BEGIN;

CREATE INDEX IF NOT EXISTS series_points_series_id_repo_id_time_idx ON series_points (series_id, repo_id, time);

COMMIT;
