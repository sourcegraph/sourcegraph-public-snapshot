CREATE TABLE IF NOT EXISTS insight_series_recording_times (
	series_id int,
	recording_time timestamptz,
	UNIQUE (series_id, recording_time),
	CONSTRAINT series_id_fkey FOREIGN KEY (series_id) REFERENCES insight_series (id) ON DELETE CASCADE
);