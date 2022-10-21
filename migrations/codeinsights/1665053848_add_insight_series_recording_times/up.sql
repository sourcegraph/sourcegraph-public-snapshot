CREATE TABLE IF NOT EXISTS insight_series_recording_times (
	insight_series_id int,
	recording_time timestamptz,
	snapshot bool,
	UNIQUE (insight_series_id, recording_time),
	CONSTRAINT insight_series_id_fkey FOREIGN KEY (insight_series_id) REFERENCES insight_series (id) ON DELETE CASCADE
);