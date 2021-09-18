BEGIN;

CREATE INDEX IF NOT EXISTS series_points_series_id_repo_id_time_idx ON series_points (series_id, repo_id, time);

-- Clear the dirty flag in case the operator timed out and isn't around to clear it.
UPDATE codeinsights_schema_migrations SET dirty = 'f'
COMMIT;
